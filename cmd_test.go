package main

import (
	"os"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"

	"github.com/ahrtr/etcd-defrag/internal/config"
)

func TestAllFlags(t *testing.T) {
	testCases := []struct {
		name string
		env  map[string]string
		cli  []string
		want config.GlobalConfig
	}{
		{
			name: "defaults only",
			env:  nil,
			cli:  nil,
			want: config.GlobalConfig{
				Endpoints:             []string{"127.0.0.1:2379"},
				Cluster:               false,
				ExcludeLocalhost:      false,
				MoveLeader:            false,
				DialTimeout:           2 * time.Second,
				CommandTimeout:        30 * time.Second,
				KeepaliveTime:         2 * time.Second,
				KeepaliveTimeout:      6 * time.Second,
				InsecureTransport:     true,
				InsecureSkipVerify:    false,
				Cert:                  "",
				Key:                   "",
				CaCert:                "",
				User:                  "",
				Password:              "",
				DiscoverySrv:          "",
				DiscoverySrvName:      "",
				InsecureDiscovery:     true,
				Compaction:            true,
				ContinueOnError:       true,
				EtcdStorageQuotaBytes: 2 * 1024 * 1024 * 1024,
				DefragRule:            "",
				PrintVersion:          false,
				DryRun:                false,
				AutoDisalarm:          false,
				DisalarmThreshold:     0.9,
			},
		},
		{
			name: "all from environment",
			env: map[string]string{
				"ETCD_DEFRAG_ENDPOINTS":                "10.0.0.1:2379,10.0.0.2:2379",
				"ETCD_DEFRAG_CLUSTER":                  "true",
				"ETCD_DEFRAG_EXCLUDE_LOCALHOST":        "true",
				"ETCD_DEFRAG_MOVE_LEADER":              "true",
				"ETCD_DEFRAG_DIAL_TIMEOUT":             "5s",
				"ETCD_DEFRAG_COMMAND_TIMEOUT":          "45s",
				"ETCD_DEFRAG_KEEPALIVE_TIME":           "3s",
				"ETCD_DEFRAG_KEEPALIVE_TIMEOUT":        "8s",
				"ETCD_DEFRAG_INSECURE_TRANSPORT":       "false",
				"ETCD_DEFRAG_INSECURE_SKIP_TLS_VERIFY": "true",
				"ETCD_DEFRAG_CERT":                     "/path/to/cert",
				"ETCD_DEFRAG_KEY":                      "/path/to/key",
				"ETCD_DEFRAG_CACERT":                   "/path/to/ca",
				"ETCD_DEFRAG_USER":                     "envuser",
				"ETCD_DEFRAG_PASSWORD":                 "envpassword",
				"ETCD_DEFRAG_DISCOVERY_SRV":            "mydomain.com",
				"ETCD_DEFRAG_DISCOVERY_SRV_NAME":       "etcd",
				"ETCD_DEFRAG_INSECURE_DISCOVERY":       "false",
				"ETCD_DEFRAG_COMPACTION":               "false",
				"ETCD_DEFRAG_CONTINUE_ON_ERROR":        "false",
				"ETCD_DEFRAG_ETCD_STORAGE_QUOTA_BYTES": "1073741824",
				"ETCD_DEFRAG_DEFRAG_RULE":              "size(db) > 500MB",
				"ETCD_DEFRAG_VERSION":                  "true",
				"ETCD_DEFRAG_DRY_RUN":                  "true",
				"ETCD_DEFRAG_AUTO_DISALARM":            "false",
				"ETCD_DEFRAG_DISALARM_THRESHOLD":       "0.9",
			},
			cli: nil,
			want: config.GlobalConfig{
				Endpoints:             []string{"10.0.0.1:2379", "10.0.0.2:2379"},
				Cluster:               true,
				ExcludeLocalhost:      true,
				MoveLeader:            true,
				DialTimeout:           5 * time.Second,
				CommandTimeout:        45 * time.Second,
				KeepaliveTime:         3 * time.Second,
				KeepaliveTimeout:      8 * time.Second,
				InsecureTransport:     false,
				InsecureSkipVerify:    true,
				Cert:                  "/path/to/cert",
				Key:                   "/path/to/key",
				CaCert:                "/path/to/ca",
				User:                  "envuser",
				Password:              "envpassword",
				DiscoverySrv:          "mydomain.com",
				DiscoverySrvName:      "etcd",
				InsecureDiscovery:     false,
				Compaction:            false,
				ContinueOnError:       false,
				EtcdStorageQuotaBytes: 1073741824,
				DefragRule:            "size(db) > 500MB",
				PrintVersion:          true,
				DryRun:                true,
				AutoDisalarm:          false,
				DisalarmThreshold:     0.9,
			},
		},
		{
			name: "all from CLI (override environment)",
			env: map[string]string{
				"ETCD_DEFRAG_ENDPOINTS":          "shouldBeOverridden:9999",
				"ETCD_DEFRAG_CLUSTER":            "false",
				"ETCD_DEFRAG_MOVE_LEADER":        "false",
				"ETCD_DEFRAG_INSECURE_TRANSPORT": "true",
			},
			cli: []string{
				"--endpoints=192.168.1.100:2379,192.168.1.101:2379",
				"--cluster=true",
				"--exclude-localhost=true",
				"--move-leader=true",
				"--dial-timeout=7s",
				"--command-timeout=50s",
				"--keepalive-time=4s",
				"--keepalive-timeout=10s",
				"--insecure-transport=false",
				"--insecure-skip-tls-verify=true",
				"--cert=/cli/cert",
				"--key=/cli/key",
				"--cacert=/cli/ca",
				"--user=cliuser",
				"--password=clipass",
				"--discovery-srv=cli.mydomain",
				"--discovery-srv-name=clietcd",
				"--insecure-discovery=false",
				"--compaction=false",
				"--continue-on-error=false",
				"--etcd-storage-quota-bytes=999999999",
				"--defrag-rule=size(db) >= 1GB",
				"--version=true",
				"--dry-run=true",
			},
			want: config.GlobalConfig{
				Endpoints:             []string{"192.168.1.100:2379", "192.168.1.101:2379"},
				Cluster:               true,
				ExcludeLocalhost:      true,
				MoveLeader:            true,
				DialTimeout:           7 * time.Second,
				CommandTimeout:        50 * time.Second,
				KeepaliveTime:         4 * time.Second,
				KeepaliveTimeout:      10 * time.Second,
				InsecureTransport:     false,
				InsecureSkipVerify:    true,
				Cert:                  "/cli/cert",
				Key:                   "/cli/key",
				CaCert:                "/cli/ca",
				User:                  "cliuser",
				Password:              "clipass",
				DiscoverySrv:          "cli.mydomain",
				DiscoverySrvName:      "clietcd",
				InsecureDiscovery:     false,
				Compaction:            false,
				ContinueOnError:       false,
				EtcdStorageQuotaBytes: 999999999,
				DefragRule:            "size(db) >= 1GB",
				PrintVersion:          true,
				DryRun:                true,
				AutoDisalarm:          false,
				DisalarmThreshold:     0.9,
			},
		},
		{
			name: "mixed env + CLI",
			env: map[string]string{
				"ETCD_DEFRAG_ENDPOINTS":                "env:2379",
				"ETCD_DEFRAG_CLUSTER":                  "false",
				"ETCD_DEFRAG_MOVE_LEADER":              "true",
				"ETCD_DEFRAG_COMPACTION":               "false",
				"ETCD_DEFRAG_ETCD_STORAGE_QUOTA_BYTES": "555555555",
			},
			cli: []string{
				"--exclude-localhost=true", // override the default
				"--dial-timeout=10s",       // override the default
				"--compaction=true",        // override the env
			},
			want: config.GlobalConfig{
				Endpoints:             []string{"env:2379"},
				Cluster:               false, // env sets cluster=false
				ExcludeLocalhost:      true,  // from CLI
				MoveLeader:            true,  // from env
				DialTimeout:           10 * time.Second,
				CommandTimeout:        30 * time.Second, // default
				KeepaliveTime:         2 * time.Second,  // default
				KeepaliveTimeout:      6 * time.Second,  // default
				InsecureTransport:     true,             // default
				InsecureSkipVerify:    false,            // default
				Cert:                  "",
				Key:                   "",
				CaCert:                "",
				User:                  "",
				Password:              "",
				DiscoverySrv:          "",
				DiscoverySrvName:      "",
				InsecureDiscovery:     true,
				Compaction:            true,      // CLI override
				ContinueOnError:       true,      // default
				EtcdStorageQuotaBytes: 555555555, // from env
				DefragRule:            "",
				PrintVersion:          false, // default
				DryRun:                false, // default
				AutoDisalarm:          false, // default
				DisalarmThreshold:     0.9,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			viper.Reset()
			os.Clearenv()
			// Reset globalCfg to defaults before each test
			globalCfg = config.GlobalConfig{}

			for key, val := range tc.env {
				if err := os.Setenv(key, val); err != nil {
					t.Fatalf("failed to set env %s=%s: %v", key, val, err)
				}
			}

			cmd := newDefragCommand()
			// Set empty Run function to avoid actual execution
			cmd.Run = func(cmd *cobra.Command, args []string) {}

			if tc.cli != nil {
				cmd.SetArgs(tc.cli)
			}

			if err := cmd.Execute(); err != nil {
				t.Fatalf("command execution failed: %v", err)
			}

			require.Equal(t, tc.want, globalCfg)
		})
	}
}
