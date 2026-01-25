package config

import (
	"errors"
	"log"

	"github.com/spf13/cobra"
)

// Validate validates the configuration and returns an error if validation fails
func (c GlobalConfig) Validate(cmd *cobra.Command) error {
	if c.Cert == "" && cmd.Flags().Changed("cert") {
		return errors.New("empty string is passed to --cert option")
	}

	if c.Key == "" && cmd.Flags().Changed("key") {
		return errors.New("empty string is passed to --key option")
	}

	if c.CaCert == "" && cmd.Flags().Changed("cacert") {
		return errors.New("empty string is passed to --cacert option")
	}

	if c.SkipHealthcheckClusterEndpoints && len(c.Endpoints) == 0 {
		return errors.New("--skip-healthcheck-cluster-endpoints requires explicit endpoints to be provided via --endpoints flag")
	}

	if c.SkipHealthcheckClusterEndpoints && c.Cluster {
		return errors.New("--skip-healthcheck-cluster-endpoints and --cluster flags are mutually exclusive")
	}

	if c.SkipHealthcheckClusterEndpoints && c.DiscoverySrv != "" {
		return errors.New("--skip-healthcheck-cluster-endpoints and --discovery-srv flags are mutually exclusive")
	}

	if c.AutoDisalarm && (c.DisalarmThreshold <= 0 || c.DisalarmThreshold >= 1) {
		return errors.New("--disalarm-threshold must be greater than 0 and less than 1.0 when --auto-disalarm is enabled")
	}

	if c.DisalarmThreshold != 0 && !c.AutoDisalarm {
		log.Println("Warning: --disalarm-threshold is set but --auto-disalarm is disabled, threshold will be ignored")
	}

	return nil
}
