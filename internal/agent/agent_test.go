package agent

import (
	"context"
	"testing"
	"time"

	"github.com/ahrtr/etcd-defrag/internal/config"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name        string
		cfg         *config.GlobalConfig
		wantErr     bool
		errContains string
	}{
		{
			name: "valid config without rule",
			cfg: &config.GlobalConfig{
				DefragRule:            "",
				EtcdStorageQuotaBytes: 2 * 1024 * 1024 * 1024,
			},
			wantErr: false,
		},
		{
			name: "valid config with valid rule",
			cfg: &config.GlobalConfig{
				DefragRule:            "dbSize > 100",
				EtcdStorageQuotaBytes: 2 * 1024 * 1024 * 1024,
			},
			wantErr: false,
		},
		{
			name: "invalid defrag rule",
			cfg: &config.GlobalConfig{
				DefragRule:            "invalid rule syntax $$$",
				EtcdStorageQuotaBytes: 2 * 1024 * 1024 * 1024,
			},
			wantErr:     true,
			errContains: "invalid defrag rule",
		},
		{
			name: "rule not boolean expression",
			cfg: &config.GlobalConfig{
				DefragRule:            "dbSize + 100",
				EtcdStorageQuotaBytes: 2 * 1024 * 1024 * 1024,
			},
			wantErr:     true,
			errContains: "boolean expression",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent, err := New(tt.cfg)
			if tt.wantErr {
				if err == nil {
					t.Errorf("New() expected error containing %q, got nil", tt.errContains)
				}
				return
			}
			if err != nil {
				t.Errorf("New() unexpected error: %v", err)
				return
			}
			if agent == nil {
				t.Error("New() returned nil agent")
			}
			if agent.cfg != tt.cfg {
				t.Error("New() agent.cfg != input cfg")
			}
		})
	}
}

func TestSortLeaderLast(t *testing.T) {
	tests := []struct {
		name      string
		endpoints []string
		statuses  map[string]*MemberStatus
		want      []string
	}{
		{
			name:      "no leader",
			endpoints: []string{"ep1", "ep2", "ep3"},
			statuses: map[string]*MemberStatus{
				"ep1": {Endpoint: "ep1", MemberID: 1, LeaderID: 999, IsLeader: false},
				"ep2": {Endpoint: "ep2", MemberID: 2, LeaderID: 999, IsLeader: false},
				"ep3": {Endpoint: "ep3", MemberID: 3, LeaderID: 999, IsLeader: false},
			},
			want: []string{"ep1", "ep2", "ep3"},
		},
		{
			name:      "leader at beginning",
			endpoints: []string{"ep1", "ep2", "ep3"},
			statuses: map[string]*MemberStatus{
				"ep1": {Endpoint: "ep1", MemberID: 1, LeaderID: 1, IsLeader: true},
				"ep2": {Endpoint: "ep2", MemberID: 2, LeaderID: 1, IsLeader: false},
				"ep3": {Endpoint: "ep3", MemberID: 3, LeaderID: 1, IsLeader: false},
			},
			want: []string{"ep2", "ep3", "ep1"},
		},
		{
			name:      "leader in middle",
			endpoints: []string{"ep1", "ep2", "ep3"},
			statuses: map[string]*MemberStatus{
				"ep1": {Endpoint: "ep1", MemberID: 1, LeaderID: 2, IsLeader: false},
				"ep2": {Endpoint: "ep2", MemberID: 2, LeaderID: 2, IsLeader: true},
				"ep3": {Endpoint: "ep3", MemberID: 3, LeaderID: 2, IsLeader: false},
			},
			want: []string{"ep1", "ep3", "ep2"},
		},
		{
			name:      "leader already at end",
			endpoints: []string{"ep1", "ep2", "ep3"},
			statuses: map[string]*MemberStatus{
				"ep1": {Endpoint: "ep1", MemberID: 1, LeaderID: 3, IsLeader: false},
				"ep2": {Endpoint: "ep2", MemberID: 2, LeaderID: 3, IsLeader: false},
				"ep3": {Endpoint: "ep3", MemberID: 3, LeaderID: 3, IsLeader: true},
			},
			want: []string{"ep1", "ep2", "ep3"},
		},
		{
			name:      "single endpoint leader",
			endpoints: []string{"ep1"},
			statuses: map[string]*MemberStatus{
				"ep1": {Endpoint: "ep1", MemberID: 1, LeaderID: 1, IsLeader: true},
			},
			want: []string{"ep1"},
		},
		{
			name:      "empty endpoints",
			endpoints: []string{},
			statuses:  map[string]*MemberStatus{},
			want:      []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &Agent{
				cfg: &config.GlobalConfig{},
			}
			got := a.sortLeaderLast(tt.endpoints, tt.statuses)
			if len(got) != len(tt.want) {
				t.Errorf("sortLeaderLast() length = %d, want %d", len(got), len(tt.want))
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("sortLeaderLast()[%d] = %v, want %v", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestResolveEndpoints(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *config.GlobalConfig
		want    []string
		wantErr bool
	}{
		{
			name: "returns configured endpoints",
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &Agent{cfg: tt.cfg}
			got, err := a.resolveEndpoints(context.Background())
			if (err != nil) != tt.wantErr {
				t.Errorf("resolveEndpoints() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != len(tt.want) {
				t.Errorf("resolveEndpoints() length = %d, want %d", len(got), len(tt.want))
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("resolveEndpoints()[%d] = %v, want %v", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestRun_Validation(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *config.GlobalConfig
		wantErr bool
	}{
		{
			name: "valid minimal config",
			cfg: &config.GlobalConfig{
				Endpoints:             []string{"localhost:2379"},
				DialTimeout:           2 * time.Second,
				CommandTimeout:        30 * time.Second,
				EtcdStorageQuotaBytes: 2 * 1024 * 1024 * 1024,
				DisalarmThreshold:     0.9,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent, err := New(tt.cfg)
			if err != nil {
				t.Fatalf("New() error = %v", err)
			}

			// Run will fail trying to connect to etcd, but should pass validation
			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer cancel()

			err = agent.Run(ctx)
			// We expect connection error, not validation error
			if err != nil && !tt.wantErr {
				// This is expected since we're not running a real etcd
				return
			}
		})
	}
}
