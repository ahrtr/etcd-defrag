package agent

import (
	"testing"

	"github.com/ahrtr/etcd-defrag/internal/eval"
)

func TestMemberStatus_ToEvalVars(t *testing.T) {
	tests := []struct {
		name   string
		status *MemberStatus
		quota  int64
		want   *eval.Variables
	}{
		{
			name: "normal status",
			status: &MemberStatus{
				Endpoint:    "ep1:2379",
				DbSize:      1000,
				DbSizeInUse: 600,
				MemberID:    1,
				LeaderID:    1,
				IsLeader:    true,
			},
			quota: 2000,
			want: &eval.Variables{
				DbSize:       1000,
				DbSizeInUse:  600,
				DbSizeFree:   400,
				DbQuota:      2000,
				DbQuotaUsage: 0.5,
			},
		},
		{
			name: "full database",
			status: &MemberStatus{
				Endpoint:    "ep2:2379",
				DbSize:      2000,
				DbSizeInUse: 2000,
				MemberID:    2,
				LeaderID:    1,
				IsLeader:    false,
			},
			quota: 2000,
			want: &eval.Variables{
				DbSize:       2000,
				DbSizeInUse:  2000,
				DbSizeFree:   0,
				DbQuota:      2000,
				DbQuotaUsage: 1.0,
			},
		},
		{
			name: "empty database",
			status: &MemberStatus{
				Endpoint:    "ep3:2379",
				DbSize:      0,
				DbSizeInUse: 0,
				MemberID:    3,
				LeaderID:    1,
				IsLeader:    false,
			},
			quota: 2000,
			want: &eval.Variables{
				DbSize:       0,
				DbSizeInUse:  0,
				DbSizeFree:   0,
				DbQuota:      2000,
				DbQuotaUsage: 0.0,
			},
		},
		{
			name: "large values",
			status: &MemberStatus{
				Endpoint:    "ep4:2379",
				DbSize:      8 * 1024 * 1024 * 1024, // 8GB
				DbSizeInUse: 6 * 1024 * 1024 * 1024, // 6GB
				MemberID:    4,
				LeaderID:    1,
				IsLeader:    false,
			},
			quota: 10 * 1024 * 1024 * 1024, // 10GB
			want: &eval.Variables{
				DbSize:       8 * 1024 * 1024 * 1024,
				DbSizeInUse:  6 * 1024 * 1024 * 1024,
				DbSizeFree:   2 * 1024 * 1024 * 1024,
				DbQuota:      10 * 1024 * 1024 * 1024,
				DbQuotaUsage: 0.8,
			},
		},
		{
			name: "minimal usage",
			status: &MemberStatus{
				Endpoint:    "ep5:2379",
				DbSize:      100,
				DbSizeInUse: 10,
				MemberID:    5,
				LeaderID:    1,
				IsLeader:    false,
			},
			quota: 10000,
			want: &eval.Variables{
				DbSize:       100,
				DbSizeInUse:  10,
				DbSizeFree:   90,
				DbQuota:      10000,
				DbQuotaUsage: 0.01,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.status.ToEvalVars(tt.quota)
			if got.DbSize != tt.want.DbSize {
				t.Errorf("ToEvalVars().DbSize = %d, want %d", got.DbSize, tt.want.DbSize)
			}
			if got.DbSizeInUse != tt.want.DbSizeInUse {
				t.Errorf("ToEvalVars().DbSizeInUse = %d, want %d", got.DbSizeInUse, tt.want.DbSizeInUse)
			}
			if got.DbSizeFree != tt.want.DbSizeFree {
				t.Errorf("ToEvalVars().DbSizeFree = %d, want %d", got.DbSizeFree, tt.want.DbSizeFree)
			}
			if got.DbQuota != tt.want.DbQuota {
				t.Errorf("ToEvalVars().DbQuota = %d, want %d", got.DbQuota, tt.want.DbQuota)
			}
			if got.DbQuotaUsage != tt.want.DbQuotaUsage {
				t.Errorf("ToEvalVars().DbQuotaUsage = %f, want %f", got.DbQuotaUsage, tt.want.DbQuotaUsage)
			}
		})
	}
}
