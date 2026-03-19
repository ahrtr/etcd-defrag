package config

import (
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// GlobalConfig holds all configuration options
type GlobalConfig struct {
	// Connection configuration
	Endpoints        []string      `mapstructure:"endpoints"`
	DialTimeout      time.Duration `mapstructure:"dial-timeout"`
	CommandTimeout   time.Duration `mapstructure:"command-timeout"`
	KeepaliveTime    time.Duration `mapstructure:"keepalive-time"`
	KeepaliveTimeout time.Duration `mapstructure:"keepalive-timeout"`

	// TLS configuration
	CaCert             string `mapstructure:"cacert"`
	Cert               string `mapstructure:"cert"`
	Key                string `mapstructure:"key"`
	InsecureTransport  bool   `mapstructure:"insecure-transport"`
	InsecureSkipVerify bool   `mapstructure:"insecure-skip-tls-verify"`

	// Discovery configuration
	Cluster           bool   `mapstructure:"cluster"`
	DiscoverySrv      string `mapstructure:"discovery-srv"`
	DiscoverySrvName  string `mapstructure:"discovery-srv-name"`
	InsecureDiscovery bool   `mapstructure:"insecure-discovery"`

	// Authentication configuration
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`

	// Behavior configuration
	Compaction      bool   `mapstructure:"compaction"`
	ContinueOnError bool   `mapstructure:"continue-on-error"`
	DefragRule      string `mapstructure:"defrag-rule"`
	DryRun          bool   `mapstructure:"dry-run"`
	// TODO: remove this when etcd v3.5 is end of life.
	// etcd v3.6.0 already added quota into the endpoint status response
	// in https://github.com/etcd-io/etcd/pull/17877
	EtcdStorageQuotaBytes           int64         `mapstructure:"etcd-storage-quota-bytes"`
	ExcludeLocalhost                bool          `mapstructure:"exclude-localhost"`
	MoveLeader                      bool          `mapstructure:"move-leader"`
	WaitBetweenDefrags              time.Duration `mapstructure:"wait-between-defrags"`
	SkipHealthcheckClusterEndpoints bool          `mapstructure:"skip-healthcheck-cluster-endpoints"`

	// Auto-disalarm configuration
	AutoDisalarm      bool    `mapstructure:"auto-disalarm"`
	DisalarmThreshold float64 `mapstructure:"disalarm-threshold"`

	// Other flags
	PrintVersion bool `mapstructure:"version"`
}

// ClientConfigWithoutEndpoints creates a clientv3.ConfigSpec from GlobalConfig without setting endpoints
func (c GlobalConfig) ClientConfigWithoutEndpoints() *clientv3.ConfigSpec {
	cfg := &clientv3.ConfigSpec{
		DialTimeout:      c.DialTimeout,
		KeepAliveTime:    c.KeepaliveTime,
		KeepAliveTimeout: c.KeepaliveTimeout,
		Secure:           c.SecureConfig(),
		Auth:             c.AuthConfig(),
	}
	return cfg
}

// SecureConfig creates a clientv3.SecureConfig from GlobalConfig
func (c GlobalConfig) SecureConfig() *clientv3.SecureConfig {
	return &clientv3.SecureConfig{
		Cert:               c.Cert,
		Key:                c.Key,
		Cacert:             c.CaCert,
		InsecureTransport:  c.InsecureTransport,
		InsecureSkipVerify: c.InsecureSkipVerify,
	}
}

// AuthConfig creates a clientv3.AuthConfig from GlobalConfig
func (c GlobalConfig) AuthConfig() *clientv3.AuthConfig {
	return &clientv3.AuthConfig{
		Username: c.User,
		Password: c.Password,
	}
}

// RegisterFlags registers all command-line flags
func RegisterFlags(cmd *cobra.Command, cfg *GlobalConfig) {
	// Manually splitting, because GetStringSlice has inconsistent behavior for splitting command line flags and environment variables
	// https://github.com/spf13/viper/issues/380
	cmd.Flags().StringSliceVar(&cfg.Endpoints, "endpoints", strings.Split(viper.GetString("endpoints"), ","),
		"comma separated etcd endpoints")

	// Connection flags
	cmd.Flags().DurationVar(&cfg.DialTimeout, "dial-timeout", viper.GetDuration("dial-timeout"),
		"dial timeout for client connections")
	cmd.Flags().DurationVar(&cfg.CommandTimeout, "command-timeout", viper.GetDuration("command-timeout"),
		"command timeout (excluding dial timeout)")
	cmd.Flags().DurationVar(&cfg.KeepaliveTime, "keepalive-time", viper.GetDuration("keepalive-time"),
		"keepalive time for client connections")
	cmd.Flags().DurationVar(&cfg.KeepaliveTimeout, "keepalive-timeout", viper.GetDuration("keepalive-timeout"),
		"keepalive timeout for client connections")

	// TLS flags
	cmd.Flags().StringVar(&cfg.CaCert, "cacert", viper.GetString("cacert"),
		"verify certificates of TLS-enabled secure servers using this CA bundle")
	cmd.Flags().StringVar(&cfg.Cert, "cert", viper.GetString("cert"),
		"identify secure client using this TLS certificate file")
	cmd.Flags().StringVar(&cfg.Key, "key", viper.GetString("key"),
		"identify secure client using this TLS key file")
	cmd.Flags().BoolVar(&cfg.InsecureTransport, "insecure-transport", viper.GetBool("insecure-transport"),
		"disable transport security for client connections")
	cmd.Flags().BoolVar(&cfg.InsecureSkipVerify, "insecure-skip-tls-verify", viper.GetBool("insecure-skip-tls-verify"),
		"skip server certificate verification (CAUTION: this option should be enabled only for testing purposes)")

	// Discovery flags
	cmd.Flags().BoolVar(&cfg.Cluster, "cluster", viper.GetBool("cluster"),
		"use all endpoints from the cluster member list")
	cmd.Flags().StringVarP(&cfg.DiscoverySrv, "discovery-srv", "d", viper.GetString("discovery-srv"),
		"domain name to query for SRV records describing cluster endpoints")
	cmd.Flags().StringVar(&cfg.DiscoverySrvName, "discovery-srv-name", viper.GetString("discovery-srv-name"),
		"service name to query when using DNS discovery")
	cmd.Flags().BoolVar(&cfg.InsecureDiscovery, "insecure-discovery", viper.GetBool("insecure-discovery"),
		"accept insecure SRV records describing cluster endpoints")

	// Auth flags
	cmd.Flags().StringVar(&cfg.User, "user", viper.GetString("user"),
		"username[:password] for authentication (prompt if password is not supplied)")
	cmd.Flags().StringVar(&cfg.Password, "password", viper.GetString("password"),
		"password for authentication (if this option is used, --user option shouldn't include password)")

	// Behavior flags
	cmd.Flags().BoolVar(&cfg.Compaction, "compaction", viper.GetBool("compaction"),
		"whether execute compaction before the defragmentation (defaults to true)")
	cmd.Flags().BoolVar(&cfg.ContinueOnError, "continue-on-error", viper.GetBool("continue-on-error"),
		"whether continue to defragment next endpoint if current one fails")
	cmd.Flags().StringVar(&cfg.DefragRule, "defrag-rule", viper.GetString("defrag-rule"),
		"defragmentation rule (etcd-defrag will run defragmentation if the rule is empty or it is evaluated to true)")
	cmd.Flags().BoolVar(&cfg.DryRun, "dry-run", viper.GetBool("dry-run"),
		"evaluate whether or not endpoints require defragmentation, but don't actually perform it")
	cmd.Flags().Int64Var(&cfg.EtcdStorageQuotaBytes, "etcd-storage-quota-bytes", int64(viper.GetInt("etcd-storage-quota-bytes")),
		"etcd storage quota in bytes (the value passed to etcd instance by flag --quota-backend-bytes)")
	cmd.Flags().BoolVar(&cfg.ExcludeLocalhost, "exclude-localhost", viper.GetBool("exclude-localhost"),
		"whether to exclude localhost endpoints")
	cmd.Flags().BoolVar(&cfg.MoveLeader, "move-leader", viper.GetBool("move-leader"),
		"whether to move the leadership before performing defragmentation on the leader")
	cmd.Flags().DurationVar(&cfg.WaitBetweenDefrags, "wait-between-defrags", viper.GetDuration("wait-between-defrags"),
		"wait time between consecutive defragmentation runs or after a leader movement (if --move-leader is enabled). Defaults to 0s (no wait)")
	cmd.Flags().BoolVar(&cfg.SkipHealthcheckClusterEndpoints, "skip-healthcheck-cluster-endpoints", viper.GetBool("skip-healthcheck-cluster-endpoints"),
		"skip cluster endpoint discovery during health check and only check the endpoints provided via --endpoints")

	// Auto-disalarm flags
	cmd.Flags().BoolVar(&cfg.AutoDisalarm, "auto-disalarm", viper.GetBool("auto-disalarm"),
		"automatically disalarm NOSPACE alarms after successful defragmentation")
	cmd.Flags().Float64Var(&cfg.DisalarmThreshold, "disalarm-threshold", viper.GetFloat64("disalarm-threshold"),
		"threshold ratio for automatic alarm clearing (db size / quota)")

	// Version flag
	cmd.Flags().BoolVar(&cfg.PrintVersion, "version", viper.GetBool("version"),
		"print the version and exit")
}
