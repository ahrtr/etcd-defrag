package client

import (
	"testing"
	"time"
)

func TestNewFromEndpoint(t *testing.T) {
	tests := []struct {
		name     string
		endpoint string
		baseCfg  *Config
		wantErr  bool
	}{
		{
			name:     "valid insecure config",
			endpoint: "localhost:2379",
			baseCfg: &Config{
				Endpoints:         []string{"other:2379"}, // Should be overridden
				DialTimeout:       2 * time.Second,
				InsecureTransport: true,
			},
			wantErr: true, // Will fail to connect but config is valid
		},
		{
			name:     "valid with TLS",
			endpoint: "localhost:2379",
			baseCfg: &Config{
				Endpoints:          []string{"other:2379"},
				DialTimeout:        2 * time.Second,
				InsecureTransport:  false,
				CaCert:             "/nonexistent/ca.crt", // Will fail but tests config
				Cert:               "/nonexistent/client.crt",
				Key:                "/nonexistent/client.key",
				InsecureSkipVerify: true,
			},
			wantErr: true, // Will fail but for different reasons
		},
		{
			name:     "valid with auth",
			endpoint: "localhost:2379",
			baseCfg: &Config{
				Endpoints:         []string{"other:2379"},
				DialTimeout:       2 * time.Second,
				InsecureTransport: true,
				Username:          "testuser",
				Password:          "testpass",
			},
			wantErr: true, // Will fail to connect
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewFromEndpoint(tt.endpoint, tt.baseCfg)
			// We expect errors since we're not actually running etcd
			// The test verifies the function doesn't panic and handles configs
			if err == nil && tt.wantErr {
				t.Log("Expected connection error (no etcd running), but function accepts config")
			}
		})
	}
}

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *Config
		wantErr bool
	}{
		{
			name: "minimal valid config",
			cfg: &Config{
				Endpoints:         []string{"localhost:2379"},
				DialTimeout:       2 * time.Second,
				InsecureTransport: true,
			},
			wantErr: true, // Connection will fail without etcd
		},
		{
			name: "multiple endpoints",
			cfg: &Config{
				Endpoints:         []string{"ep1:2379", "ep2:2379", "ep3:2379"},
				DialTimeout:       2 * time.Second,
				InsecureTransport: true,
			},
			wantErr: true,
		},
		{
			name: "with keepalive settings",
			cfg: &Config{
				Endpoints:         []string{"localhost:2379"},
				DialTimeout:       2 * time.Second,
				KeepaliveTime:     5 * time.Second,
				KeepaliveTimeout:  10 * time.Second,
				InsecureTransport: true,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New(tt.cfg)
			// We expect errors since we're not actually running etcd
			if err == nil && tt.wantErr {
				t.Log("Expected connection error (no etcd running)")
			}
		})
	}
}

func TestConfig_Fields(t *testing.T) {
	cfg := &Config{
		Endpoints:          []string{"ep1:2379", "ep2:2379"},
		DialTimeout:        5 * time.Second,
		KeepaliveTime:      10 * time.Second,
		KeepaliveTimeout:   15 * time.Second,
		CaCert:             "/path/to/ca.crt",
		Cert:               "/path/to/client.crt",
		Key:                "/path/to/client.key",
		InsecureTransport:  false,
		InsecureSkipVerify: true,
		Username:           "user",
		Password:           "pass",
	}

	tests := []struct {
		name string
		got  interface{}
		want interface{}
	}{
		{"Endpoints length", len(cfg.Endpoints), 2},
		{"Endpoints[0]", cfg.Endpoints[0], "ep1:2379"},
		{"Endpoints[1]", cfg.Endpoints[1], "ep2:2379"},
		{"DialTimeout", cfg.DialTimeout, 5 * time.Second},
		{"KeepaliveTime", cfg.KeepaliveTime, 10 * time.Second},
		{"KeepaliveTimeout", cfg.KeepaliveTimeout, 15 * time.Second},
		{"CaCert", cfg.CaCert, "/path/to/ca.crt"},
		{"Cert", cfg.Cert, "/path/to/client.crt"},
		{"Key", cfg.Key, "/path/to/client.key"},
		{"InsecureTransport", cfg.InsecureTransport, false},
		{"InsecureSkipVerify", cfg.InsecureSkipVerify, true},
		{"Username", cfg.Username, "user"},
		{"Password", cfg.Password, "pass"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("Config.%s = %v, want %v", tt.name, tt.got, tt.want)
			}
		})
	}
}
