package agent

import (
	"context"
	"log"

	"github.com/ahrtr/etcd-defrag/internal/eval"
)

// MemberStatus holds etcd member status information
type MemberStatus struct {
	Endpoint    string
	DbSize      int64
	DbSizeInUse int64
	MemberID    uint64
	LeaderID    uint64
	IsLeader    bool
	Revision    int64
	RaftTerm    uint64
	RaftIndex   uint64
}

// ToEvalVars converts status to evaluation variables
func (s *MemberStatus) ToEvalVars(quota int64) *eval.Variables {
	return &eval.Variables{
		DbSize:       s.DbSize,
		DbSizeInUse:  s.DbSizeInUse,
		DbSizeFree:   s.DbSize - s.DbSizeInUse,
		DbQuota:      quota,
		DbQuotaUsage: float64(s.DbSize) / float64(quota),
	}
}

// getMemberStatus fetches status for all endpoints
func (a *Agent) getMemberStatus(ctx context.Context, endpoints []string) (map[string]*MemberStatus, error) {
	statuses := make(map[string]*MemberStatus)

	for _, ep := range endpoints {
		status, err := a.getEndpointStatus(ctx, ep)
		if err != nil {
			return nil, err
		}
		statuses[ep] = status
	}

	return statuses, nil
}

func (a *Agent) getEndpointStatus(ctx context.Context, endpoint string) (*MemberStatus, error) {
	cli, err := a.getClient(endpoint)
	if err != nil {
		return nil, err
	}
	defer cli.Close()

	ctx, cancel := context.WithTimeout(ctx, a.cfg.CommandTimeout)
	defer cancel()

	resp, err := cli.Status(ctx, endpoint)
	if err != nil {
		return nil, err
	}

	status := &MemberStatus{
		Endpoint:    endpoint,
		DbSize:      resp.DbSize,
		DbSizeInUse: resp.DbSizeInUse,
		MemberID:    resp.Header.MemberId,
		LeaderID:    resp.Leader,
		IsLeader:    resp.Header.MemberId == resp.Leader,
		Revision:    resp.Header.Revision,
		RaftTerm:    resp.RaftTerm,
		RaftIndex:   resp.RaftIndex,
	}

	log.Printf("endpoint: %s, dbSize: %d, dbSizeInUse: %d, memberId: %x, leader: %x, revision: %d, term: %d, index: %d",
		endpoint, status.DbSize, status.DbSizeInUse, status.MemberID, status.LeaderID,
		status.Revision, status.RaftTerm, status.RaftIndex)

	return status, nil
}
