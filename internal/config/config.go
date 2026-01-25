// Package config provides configuration management for etcd-defrag
package config

import (
	"time"

	"github.com/spf13/cobra"
)

// GlobalConfig holds all configuration for etcd-defrag
type GlobalConfig struct {
	// Endpoints
	Endpoints []string

	// Timeouts
	DialTimeout      time.Duration
	CommandTimeout   time.Duration
	KeepaliveTime    time.Duration
	KeepaliveTimeout time.Duration

	// TLS
	CaCert             string
	Cert               string
	Key                string
	InsecureTransport  bool
	InsecureSkipVerify bool

	// Cluster discovery
	Cluster           bool
	DiscoverySrv      string
	DiscoverySrvName  string
	InsecureDiscovery bool

	// Auth
	User     string
	Password string

	// Defrag behavior
	Compaction                      bool
	ContinueOnError                 bool
	DefragRule                      string
	DryRun                          bool
	EtcdStorageQuotaBytes           int64
	ExcludeLocalhost                bool
	MoveLeader                      bool
	WaitBetweenDefrags              time.Duration
	SkipHealthcheckClusterEndpoints bool

	// Auto-disalarm
	AutoDisalarm      bool
	DisalarmThreshold float64

	// Version
	PrintVersion bool
}

// NewGlobalConfig creates a new config with default values
func NewGlobalConfig() *GlobalConfig {
	return &GlobalConfig{
		Endpoints:             []string{"127.0.0.1:2379"},
		DialTimeout:           2 * time.Second,
		CommandTimeout:        30 * time.Second,
		KeepaliveTime:         2 * time.Second,
		KeepaliveTimeout:      6 * time.Second,
		InsecureTransport:     true,
		InsecureDiscovery:     true,
		Compaction:            true,
		ContinueOnError:       true,
		EtcdStorageQuotaBytes: 2 * 1024 * 1024 * 1024, // 2GB
		DisalarmThreshold:     0.9,
	}
}

// RegisterFlags registers all command-line flags
func RegisterFlags(cmd *cobra.Command, cfg *GlobalConfig) {
	// Connection flags
	cmd.Flags().StringSliceVar(&cfg.Endpoints, "endpoints", cfg.Endpoints,
		"comma separated etcd endpoints")
	cmd.Flags().DurationVar(&cfg.DialTimeout, "dial-timeout", cfg.DialTimeout,
		"dial timeout for client connections")
	cmd.Flags().DurationVar(&cfg.CommandTimeout, "command-timeout", cfg.CommandTimeout,
		"command timeout (excluding dial timeout)")
	cmd.Flags().DurationVar(&cfg.KeepaliveTime, "keepalive-time", cfg.KeepaliveTime,
		"keepalive time for client connections")
	cmd.Flags().DurationVar(&cfg.KeepaliveTimeout, "keepalive-timeout", cfg.KeepaliveTimeout,
		"keepalive timeout for client connections")

	// TLS flags
	cmd.Flags().StringVar(&cfg.CaCert, "cacert", "",
		"verify certificates of TLS-enabled secure servers using this CA bundle")
	cmd.Flags().StringVar(&cfg.Cert, "cert", "",
		"identify secure client using this TLS certificate file")
	cmd.Flags().StringVar(&cfg.Key, "key", "",
		"identify secure client using this TLS key file")
	cmd.Flags().BoolVar(&cfg.InsecureTransport, "insecure-transport", cfg.InsecureTransport,
		"disable transport security for client connections")
	cmd.Flags().BoolVar(&cfg.InsecureSkipVerify, "insecure-skip-tls-verify", cfg.InsecureSkipVerify,
		"skip server certificate verification")

	// Discovery flags
	cmd.Flags().BoolVar(&cfg.Cluster, "cluster", false,
		"use all endpoints from the cluster member list")
	cmd.Flags().StringVarP(&cfg.DiscoverySrv, "discovery-srv", "d", "",
		"domain name to query for SRV records describing cluster endpoints")
	cmd.Flags().StringVar(&cfg.DiscoverySrvName, "discovery-srv-name", "",
		"service name to query when using DNS discovery")
	cmd.Flags().BoolVar(&cfg.InsecureDiscovery, "insecure-discovery", cfg.InsecureDiscovery,
		"accept insecure SRV records describing cluster endpoints")

	// Auth flags
	cmd.Flags().StringVar(&cfg.User, "user", "",
		"username[:password] for authentication (prompt if password is not supplied)")
	cmd.Flags().StringVar(&cfg.Password, "password", "",
		"password for authentication (if this option is used, --user option shouldn't include password)")

	// Behavior flags
	cmd.Flags().BoolVar(&cfg.Compaction, "compaction", cfg.Compaction,
		"whether execute compaction before the defragmentation (defaults to true)")
	cmd.Flags().BoolVar(&cfg.ContinueOnError, "continue-on-error", cfg.ContinueOnError,
		"whether continue to defragment next endpoint if current one fails")
	cmd.Flags().StringVar(&cfg.DefragRule, "defrag-rule", "",
		"defragmentation rule (etcd-defrag will run defragmentation if the rule is empty or it is evaluated to true)")
	cmd.Flags().BoolVar(&cfg.DryRun, "dry-run", false,
		"evaluate whether or not endpoints require defragmentation, but don't actually perform it")
	cmd.Flags().Int64Var(&cfg.EtcdStorageQuotaBytes, "etcd-storage-quota-bytes", cfg.EtcdStorageQuotaBytes,
		"etcd storage quota in bytes (the value passed to etcd instance by flag --quota-backend-bytes)")
	cmd.Flags().BoolVar(&cfg.ExcludeLocalhost, "exclude-localhost", false,
		"whether to exclude localhost endpoints")
	cmd.Flags().BoolVar(&cfg.MoveLeader, "move-leader", false,
		"whether to move the leadership before performing defragmentation on the leader")
	cmd.Flags().DurationVar(&cfg.WaitBetweenDefrags, "wait-between-defrags", 0,
		"wait time between consecutive defragmentation runs or after a leader movement")
	cmd.Flags().BoolVar(&cfg.SkipHealthcheckClusterEndpoints, "skip-healthcheck-cluster-endpoints", false,
		"skip cluster endpoint discovery during health check and only check the endpoints provided via --endpoints")

	// Auto-disalarm flags
	cmd.Flags().BoolVar(&cfg.AutoDisalarm, "auto-disalarm", false,
		"automatically disalarm NOSPACE alarms after successful defragmentation")
	cmd.Flags().Float64Var(&cfg.DisalarmThreshold, "disalarm-threshold", cfg.DisalarmThreshold,
		"threshold ratio for automatic alarm clearing (db size / quota)")

	// Version flag
	cmd.Flags().BoolVar(&cfg.PrintVersion, "version", false,
		"print the version and exit")
}
