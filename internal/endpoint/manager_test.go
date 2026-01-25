package endpoint

import (
	"context"
	"testing"

	"github.com/ahrtr/etcd-defrag/internal/config"
)

func TestNewManager(t *testing.T) {
	cfg := &config.GlobalConfig{
		Endpoints: []string{"ep1:2379", "ep2:2379"},
	}

	mgr := NewManager(cfg)
	if mgr == nil {
		t.Fatal("NewManager() returned nil")
	}
	if mgr.cfg != cfg {
		t.Error("NewManager() did not set config correctly")
	}
}

func TestManager_EndpointsFromCmd(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *config.GlobalConfig
		want    []string
		wantErr bool
	}{
		{
			name: "simple endpoints",
			cfg: &config.GlobalConfig{
				Endpoints: []string{"ep1:2379", "ep2:2379"},
			},
			want:    []string{"ep1:2379", "ep2:2379"},
			wantErr: false,
		},
		{
			name: "single endpoint",
			cfg: &config.GlobalConfig{
				Endpoints: []string{"localhost:2379"},
			},
			want:    []string{"localhost:2379"},
			wantErr: false,
		},
		{
			name: "no endpoints",
			cfg: &config.GlobalConfig{
				Endpoints: []string{},
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr := NewManager(tt.cfg)
			got, err := mgr.endpointsFromCmd()
			if (err != nil) != tt.wantErr {
				t.Errorf("endpointsFromCmd() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if len(got) != len(tt.want) {
					t.Errorf("endpointsFromCmd() returned %d endpoints, want %d", len(got), len(tt.want))
					return
				}
				for i := range got {
					if got[i] != tt.want[i] {
						t.Errorf("endpointsFromCmd()[%d] = %v, want %v", i, got[i], tt.want[i])
					}
				}
			}
		})
	}
}

func TestManager_EndpointsFromDNSDiscovery(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *config.GlobalConfig
		wantNil bool
		wantErr bool
	}{
		{
			name: "no DNS discovery",
			cfg: &config.GlobalConfig{
				DiscoverySrv: "",
			},
			wantNil: true,
			wantErr: false,
		},
		{
			name: "with invalid DNS discovery",
			cfg: &config.GlobalConfig{
				DiscoverySrv:     "nonexistent.invalid.domain.example",
				DiscoverySrvName: "etcd-client",
			},
			wantNil: false,
			wantErr: true, // Will fail DNS lookup
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr := NewManager(tt.cfg)
			got, err := mgr.endpointsFromDNSDiscovery()
			if (err != nil) != tt.wantErr {
				t.Errorf("endpointsFromDNSDiscovery() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantNil && got != nil {
				t.Errorf("endpointsFromDNSDiscovery() = %v, want nil", got)
			}
		})
	}
}

func TestManager_Resolve(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *config.GlobalConfig
		want    []string
		wantErr bool
	}{
		{
			name: "non-cluster mode",
			cfg: &config.GlobalConfig{
				Endpoints: []string{"ep1:2379", "ep2:2379"},
				Cluster:   false,
			},
			want:    []string{"ep1:2379", "ep2:2379"},
			wantErr: false,
		},
		{
			name: "cluster mode will try to connect",
			cfg: &config.GlobalConfig{
				Endpoints: []string{"nonexistent:2379"},
				Cluster:   true,
			},
			want:    nil,
			wantErr: true, // Will fail to connect
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr := NewManager(tt.cfg)
			ctx := context.Background()
			got, err := mgr.Resolve(ctx)
			if (err != nil) != tt.wantErr {
				if tt.wantErr {
					// Expected error (no etcd running)
					return
				}
				t.Errorf("Resolve() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if len(got) != len(tt.want) {
					t.Errorf("Resolve() returned %d endpoints, want %d", len(got), len(tt.want))
					return
				}
				for i := range got {
					if got[i] != tt.want[i] {
						t.Errorf("Resolve()[%d] = %v, want %v", i, got[i], tt.want[i])
					}
				}
			}
		})
	}
}
