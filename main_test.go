package main

import (
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

func TestValidateConfig_SkipHealthcheckClusterEndpoints(t *testing.T) {
	testCases := []struct {
		name      string
		cfg       globalConfig
		expectErr bool
	}{
		{
			name: "flag disabled with no endpoints",
			cfg: globalConfig{
				skipHealthcheckClusterEndpoints: false,
				endpoints:                       []string{},
			},
			expectErr: false,
		},
		{
			name: "flag disabled with endpoints",
			cfg: globalConfig{
				skipHealthcheckClusterEndpoints: false,
				endpoints:                       []string{"127.0.0.1:2379"},
			},
			expectErr: false,
		},
		{
			name: "flag enabled with no endpoints",
			cfg: globalConfig{
				skipHealthcheckClusterEndpoints: true,
				endpoints:                       []string{},
			},
			expectErr: true,
		},
		{
			name: "flag enabled with single endpoint",
			cfg: globalConfig{
				skipHealthcheckClusterEndpoints: true,
				endpoints:                       []string{"192.168.1.10:2379"},
			},
			expectErr: false,
		},
		{
			name: "flag enabled with multiple endpoints",
			cfg: globalConfig{
				skipHealthcheckClusterEndpoints: true,
				endpoints:                       []string{"192.168.1.10:2379", "192.168.1.11:2379"},
			},
			expectErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := &cobra.Command{}
			err := validateConfig(cmd, tc.cfg)

			if tc.expectErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), "--skip-healthcheck-cluster-endpoints requires explicit endpoints")
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestAllFlags_SkipHealthcheckClusterEndpoints(t *testing.T) {
	testCases := []struct {
		name string
		env  map[string]string
		cli  []string
		want bool
	}{
		{
			name: "default value",
			env:  nil,
			cli:  nil,
			want: false,
		},
		{
			name: "from environment",
			env: map[string]string{
				"ETCD_DEFRAG_SKIP_HEALTHCHECK_CLUSTER_ENDPOINTS": "true",
			},
			cli:  nil,
			want: true,
		},
		{
			name: "from CLI",
			env:  nil,
			cli: []string{
				"--skip-healthcheck-cluster-endpoints=true",
			},
			want: true,
		},
		{
			name: "CLI overrides environment",
			env: map[string]string{
				"ETCD_DEFRAG_SKIP_HEALTHCHECK_CLUSTER_ENDPOINTS": "true",
			},
			cli: []string{
				"--skip-healthcheck-cluster-endpoints=false",
			},
			want: false,
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
			cmd.Run = func(cmd *cobra.Command, args []string) {}

			if tc.cli != nil {
				cmd.SetArgs(tc.cli)
			}

			err := cmd.Execute()
			require.NoError(t, err)

			require.Equal(t, tc.want, globalCfg.skipHealthcheckClusterEndpoints)
		})
	}
}
