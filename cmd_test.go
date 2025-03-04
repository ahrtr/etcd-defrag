package main

import (
	"os"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

func TestAllFlags(t *testing.T) {
	testCases := []struct {
		name string
		env  map[string]string
		cli  []string
		want globalConfig
	}{
		{
			name: "defaults only",
			env:  nil,
			cli:  nil,
			want: globalConfig{
				endpoints:           []string{"127.0.0.1:2379"},
				useClusterEndpoints: false,
				excludeLocalhost:    false,
				moveLeader:          false,
				dialTimeout:         2 * time.Second,
				commandTimeout:      30 * time.Second,
				keepAliveTime:       2 * time.Second,
				keepAliveTimeout:    6 * time.Second,
				insecure:            true,
				insecureSkepVerify:  false,
				certFile:            "",
				keyFile:             "",
				caFile:              "",
				username:            "",
				password:            "",
				dnsDomain:           "",
				dnsService:          "",
				insecureDiscovery:   true,
				compaction:          true,
				continueOnError:     true,
				dbQuotaBytes:        2 * 1024 * 1024 * 1024,
				defragRule:          "",
				printVersion:        false,
				dryRun:              false,
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
			},
			cli: nil,
			want: globalConfig{
				endpoints:           []string{"10.0.0.1:2379", "10.0.0.2:2379"},
				useClusterEndpoints: true,
				excludeLocalhost:    true,
				moveLeader:          true,
				dialTimeout:         5 * time.Second,
				commandTimeout:      45 * time.Second,
				keepAliveTime:       3 * time.Second,
				keepAliveTimeout:    8 * time.Second,
				insecure:            false,
				insecureSkepVerify:  true,
				certFile:            "/path/to/cert",
				keyFile:             "/path/to/key",
				caFile:              "/path/to/ca",
				username:            "envuser",
				password:            "envpassword",
				dnsDomain:           "mydomain.com",
				dnsService:          "etcd",
				insecureDiscovery:   false,
				compaction:          false,
				continueOnError:     false,
				dbQuotaBytes:        1073741824,
				defragRule:          "size(db) > 500MB",
				printVersion:        true,
				dryRun:              true,
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
			want: globalConfig{
				endpoints:           []string{"192.168.1.100:2379", "192.168.1.101:2379"},
				useClusterEndpoints: true,
				excludeLocalhost:    true,
				moveLeader:          true,
				dialTimeout:         7 * time.Second,
				commandTimeout:      50 * time.Second,
				keepAliveTime:       4 * time.Second,
				keepAliveTimeout:    10 * time.Second,
				insecure:            false,
				insecureSkepVerify:  true,
				certFile:            "/cli/cert",
				keyFile:             "/cli/key",
				caFile:              "/cli/ca",
				username:            "cliuser",
				password:            "clipass",
				dnsDomain:           "cli.mydomain",
				dnsService:          "clietcd",
				insecureDiscovery:   false,
				compaction:          false,
				continueOnError:     false,
				dbQuotaBytes:        999999999,
				defragRule:          "size(db) >= 1GB",
				printVersion:        true,
				dryRun:              true,
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
			want: globalConfig{
				endpoints:           []string{"env:2379"},
				useClusterEndpoints: false, // env sets cluster=false
				excludeLocalhost:    true,  // from CLI
				moveLeader:          true,  // from env
				dialTimeout:         10 * time.Second,
				commandTimeout:      30 * time.Second, // default
				keepAliveTime:       2 * time.Second,  // default
				keepAliveTimeout:    6 * time.Second,  // default
				insecure:            true,             // default
				insecureSkepVerify:  false,            // default
				certFile:            "",
				keyFile:             "",
				caFile:              "",
				username:            "",
				password:            "",
				dnsDomain:           "",
				dnsService:          "",
				insecureDiscovery:   true,
				compaction:          true,      // CLI override
				continueOnError:     true,      // default
				dbQuotaBytes:        555555555, // from env
				defragRule:          "",
				printVersion:        false, // default
				dryRun:              false, // default
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			viper.Reset()
			os.Clearenv()

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
