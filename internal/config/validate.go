package config

import (
	"fmt"
	"time"
)

// ValidationError represents a configuration validation error
type ValidationError struct {
	Field   string
	Value   interface{}
	Message string
	Hint    string
}

func (e ValidationError) Error() string {
	if e.Hint != "" {
		return fmt.Sprintf("%s: %s (hint: %s)", e.Field, e.Message, e.Hint)
	}
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// Validate validates the configuration and returns any errors
func (c *GlobalConfig) Validate() []ValidationError {
	var errs []ValidationError

	// Endpoint validation
	if len(c.Endpoints) == 0 && c.DiscoverySrv == "" && !c.Cluster {
		errs = append(errs, ValidationError{
			Field:   "endpoints",
			Message: "no endpoints specified",
			Hint:    "use --endpoints, --discovery-srv, or --cluster",
		})
	}

	// Disalarm threshold validation
	if c.AutoDisalarm && (c.DisalarmThreshold <= 0 || c.DisalarmThreshold >= 1) {
		errs = append(errs, ValidationError{
			Field:   "disalarm-threshold",
			Value:   c.DisalarmThreshold,
			Message: "must be between 0 and 1 (exclusive)",
			Hint:    "try --disalarm-threshold=0.9",
		})
	}

	// Timeout validation
	if c.CommandTimeout < 5*time.Second {
		errs = append(errs, ValidationError{
			Field:   "command-timeout",
			Value:   c.CommandTimeout,
			Message: "may be too short for defragmentation",
			Hint:    "defrag can take 30s+, try --command-timeout=60s",
		})
	}

	// Storage quota validation
	if c.EtcdStorageQuotaBytes < 100*1024*1024 {
		errs = append(errs, ValidationError{
			Field:   "etcd-storage-quota-bytes",
			Value:   c.EtcdStorageQuotaBytes,
			Message: "quota seems too small (< 100MB)",
			Hint:    "ensure this matches your etcd --quota-backend-bytes",
		})
	}

	return errs
}
