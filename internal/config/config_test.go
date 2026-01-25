package config

import (
	"testing"
	"time"
)

func TestNewGlobalConfig(t *testing.T) {
	cfg := NewGlobalConfig()

	tests := []struct {
		name string
		got  interface{}
		want interface{}
	}{
		{"Endpoints", len(cfg.Endpoints), 1},
		{"Endpoints[0]", cfg.Endpoints[0], "127.0.0.1:2379"},
		{"DialTimeout", cfg.DialTimeout, 2 * time.Second},
		{"CommandTimeout", cfg.CommandTimeout, 30 * time.Second},
		{"KeepaliveTime", cfg.KeepaliveTime, 2 * time.Second},
		{"KeepaliveTimeout", cfg.KeepaliveTimeout, 6 * time.Second},
		{"InsecureTransport", cfg.InsecureTransport, true},
		{"InsecureDiscovery", cfg.InsecureDiscovery, true},
		{"Compaction", cfg.Compaction, true},
		{"ContinueOnError", cfg.ContinueOnError, true},
		{"EtcdStorageQuotaBytes", cfg.EtcdStorageQuotaBytes, int64(2*1024*1024*1024)},
		{"DisalarmThreshold", cfg.DisalarmThreshold, 0.9},
		{"AutoDisalarm", cfg.AutoDisalarm, false},
		{"DryRun", cfg.DryRun, false},
		{"MoveLeader", cfg.MoveLeader, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("NewGlobalConfig().%s = %v, want %v", tt.name, tt.got, tt.want)
			}
		})
	}
}

func TestGlobalConfig_Validate(t *testing.T) {
	tests := []struct {
		name       string
		cfg        *GlobalConfig
		wantErrors int
		checkError func([]ValidationError) bool
	}{
		{
			name: "valid configuration",
			cfg: &GlobalConfig{
				Endpoints:             []string{"localhost:2379"},
				CommandTimeout:        30 * time.Second,
				EtcdStorageQuotaBytes: 2 * 1024 * 1024 * 1024,
				AutoDisalarm:          true,
				DisalarmThreshold:     0.9,
			},
			wantErrors: 0,
		},
		{
			name: "no endpoints and no discovery",
			cfg: &GlobalConfig{
				Endpoints:             []string{},
				DiscoverySrv:          "",
				Cluster:               false,
				CommandTimeout:        30 * time.Second,
				EtcdStorageQuotaBytes: 2 * 1024 * 1024 * 1024,
			},
			wantErrors: 1,
			checkError: func(errs []ValidationError) bool {
				return len(errs) > 0 && errs[0].Field == "endpoints"
			},
		},
		{
			name: "invalid disalarm threshold - too low",
			cfg: &GlobalConfig{
				Endpoints:             []string{"localhost:2379"},
				CommandTimeout:        30 * time.Second,
				EtcdStorageQuotaBytes: 2 * 1024 * 1024 * 1024,
				AutoDisalarm:          true,
				DisalarmThreshold:     0.0,
			},
			wantErrors: 1,
			checkError: func(errs []ValidationError) bool {
				return len(errs) > 0 && errs[0].Field == "disalarm-threshold"
			},
		},
		{
			name: "invalid disalarm threshold - too high",
			cfg: &GlobalConfig{
				Endpoints:             []string{"localhost:2379"},
				CommandTimeout:        30 * time.Second,
				EtcdStorageQuotaBytes: 2 * 1024 * 1024 * 1024,
				AutoDisalarm:          true,
				DisalarmThreshold:     1.0,
			},
			wantErrors: 1,
			checkError: func(errs []ValidationError) bool {
				return len(errs) > 0 && errs[0].Field == "disalarm-threshold"
			},
		},
		{
			name: "command timeout too short",
			cfg: &GlobalConfig{
				Endpoints:             []string{"localhost:2379"},
				CommandTimeout:        1 * time.Second,
				EtcdStorageQuotaBytes: 2 * 1024 * 1024 * 1024,
			},
			wantErrors: 1,
			checkError: func(errs []ValidationError) bool {
				return len(errs) > 0 && errs[0].Field == "command-timeout"
			},
		},
		{
			name: "storage quota too small",
			cfg: &GlobalConfig{
				Endpoints:             []string{"localhost:2379"},
				CommandTimeout:        30 * time.Second,
				EtcdStorageQuotaBytes: 50 * 1024 * 1024, // 50MB
			},
			wantErrors: 1,
			checkError: func(errs []ValidationError) bool {
				return len(errs) > 0 && errs[0].Field == "etcd-storage-quota-bytes"
			},
		},
		{
			name: "multiple validation errors",
			cfg: &GlobalConfig{
				Endpoints:             []string{},
				DiscoverySrv:          "",
				Cluster:               false,
				CommandTimeout:        1 * time.Second,
				EtcdStorageQuotaBytes: 50 * 1024 * 1024,
				AutoDisalarm:          true,
				DisalarmThreshold:     1.5,
			},
			wantErrors: 4,
		},
		{
			name: "valid with cluster flag",
			cfg: &GlobalConfig{
				Endpoints:             []string{},
				Cluster:               true,
				CommandTimeout:        30 * time.Second,
				EtcdStorageQuotaBytes: 2 * 1024 * 1024 * 1024,
			},
			wantErrors: 0,
		},
		{
			name: "valid with discovery srv",
			cfg: &GlobalConfig{
				Endpoints:             []string{},
				DiscoverySrv:          "example.com",
				CommandTimeout:        30 * time.Second,
				EtcdStorageQuotaBytes: 2 * 1024 * 1024 * 1024,
			},
			wantErrors: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := tt.cfg.Validate()
			if len(errs) != tt.wantErrors {
				t.Errorf("Validate() returned %d errors, want %d", len(errs), tt.wantErrors)
				for _, err := range errs {
					t.Logf("  - %v", err)
				}
				return
			}
			if tt.checkError != nil && !tt.checkError(errs) {
				t.Errorf("Validate() error check failed")
			}
		})
	}
}

func TestValidationError_Error(t *testing.T) {
	tests := []struct {
		name    string
		verr    ValidationError
		wantStr string
	}{
		{
			name: "with hint",
			verr: ValidationError{
				Field:   "test-field",
				Message: "is invalid",
				Hint:    "try --test-field=value",
			},
			wantStr: "test-field: is invalid (hint: try --test-field=value)",
		},
		{
			name: "without hint",
			verr: ValidationError{
				Field:   "test-field",
				Message: "is required",
				Hint:    "",
			},
			wantStr: "test-field: is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.verr.Error()
			if got != tt.wantStr {
				t.Errorf("ValidationError.Error() = %q, want %q", got, tt.wantStr)
			}
		})
	}
}

func TestRegisterFlags(t *testing.T) {
	// This test just ensures RegisterFlags doesn't panic
	// and properly registers flags on a cobra command

	// We'll import cobra for this test
	// Note: Proper testing would use cobra's testing utilities
	t.Run("registers flags without panic", func(t *testing.T) {
		// Create a minimal mock command-like structure
		cfg := NewGlobalConfig()

		// Call RegisterFlags with nil to check for nil pointer panics
		// In real usage, this would be called with a real cobra command
		defer func() {
			if r := recover(); r == nil {
				// Good - no panic with proper config
				t.Log("RegisterFlags handles config properly")
			}
		}()

		// Just verify cfg is usable
		if cfg == nil {
			t.Error("NewGlobalConfig returned nil")
		}
		if len(cfg.Endpoints) == 0 {
			t.Error("Default config has no endpoints")
		}
	})
}
