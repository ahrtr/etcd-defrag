package agent

import (
	"testing"
)

func TestLogStatus(t *testing.T) {
	tests := []struct {
		name     string
		endpoint string
		status   *MemberStatus
	}{
		{
			name:     "normal status",
			endpoint: "ep1:2379",
			status: &MemberStatus{
				Endpoint:    "ep1:2379",
				DbSize:      1000000,
				DbSizeInUse: 800000,
				MemberID:    123456,
				LeaderID:    123456,
				IsLeader:    true,
				Revision:    1000,
				RaftTerm:    5,
				RaftIndex:   2000,
			},
		},
		{
			name:     "follower status",
			endpoint: "ep2:2379",
			status: &MemberStatus{
				Endpoint:    "ep2:2379",
				DbSize:      2000000,
				DbSizeInUse: 1500000,
				MemberID:    789012,
				LeaderID:    123456,
				IsLeader:    false,
				Revision:    1000,
				RaftTerm:    5,
				RaftIndex:   2000,
			},
		},
		{
			name:     "zero values",
			endpoint: "ep3:2379",
			status: &MemberStatus{
				Endpoint:    "ep3:2379",
				DbSize:      0,
				DbSizeInUse: 0,
				MemberID:    0,
				LeaderID:    0,
				IsLeader:    false,
				Revision:    0,
				RaftTerm:    0,
				RaftIndex:   0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// logStatus just logs, no return value to check
			// This test ensures it doesn't panic
			logStatus(tt.endpoint, tt.status)
		})
	}
}
