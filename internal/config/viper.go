package config

import (
	"strings"
	"time"

	"github.com/spf13/viper"
)

// SetupViper configures viper for environment variable handling and sets default values
func SetupViper() {
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.SetEnvPrefix("ETCD_DEFRAG")
	viper.AutomaticEnv()
	setDefaults()
}

func setDefaults() {
	viper.SetDefault("endpoints", "127.0.0.1:2379")
	viper.SetDefault("cluster", false)
	viper.SetDefault("exclude-localhost", false)
	viper.SetDefault("move-leader", false)
	viper.SetDefault("wait-between-defrags", 0*time.Second)
	viper.SetDefault("dial-timeout", 2*time.Second)
	viper.SetDefault("command-timeout", 30*time.Second)
	viper.SetDefault("keepalive-time", 2*time.Second)
	viper.SetDefault("keepalive-timeout", 6*time.Second)
	viper.SetDefault("insecure-transport", true)
	viper.SetDefault("insecure-skip-tls-verify", false)
	viper.SetDefault("cert", "")
	viper.SetDefault("key", "")
	viper.SetDefault("cacert", "")
	viper.SetDefault("user", "")
	viper.SetDefault("password", "")
	viper.SetDefault("discovery-srv", "")
	viper.SetDefault("discovery-srv-name", "")
	viper.SetDefault("insecure-discovery", true)
	viper.SetDefault("compaction", true)
	viper.SetDefault("continue-on-error", true)
	viper.SetDefault("etcd-storage-quota-bytes", 2*1024*1024*1024)
	viper.SetDefault("defrag-rule", "")
	viper.SetDefault("version", false)
	viper.SetDefault("dry-run", false)
	viper.SetDefault("skip-healthcheck-cluster-endpoints", false)
	viper.SetDefault("auto-disalarm", false)
	viper.SetDefault("disalarm-threshold", 0.9)
}
