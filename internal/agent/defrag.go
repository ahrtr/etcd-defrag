package agent

import (
	"context"
	"log"
	"time"
)

// defragEndpoints defragments each endpoint
func (a *Agent) defragEndpoints(ctx context.Context, endpoints []string, statuses map[string]*MemberStatus) error {
	for _, ep := range endpoints {
		if err := a.defragEndpoint(ctx, ep, statuses[ep]); err != nil {
			if !a.cfg.ContinueOnError {
				return err
			}
			log.Printf("Error defragmenting %s: %v, continuing...", ep, err)
		}

		if a.cfg.WaitBetweenDefrags > 0 {
			log.Printf("Waiting %v before next defragmentation...", a.cfg.WaitBetweenDefrags)
			time.Sleep(a.cfg.WaitBetweenDefrags)
		}
	}
	return nil
}

func (a *Agent) defragEndpoint(ctx context.Context, endpoint string, status *MemberStatus) error {
	log.Println("[Before defragmentation]")
	logStatus(endpoint, status)

	// Check if defrag is needed based on rule
	if a.evaluator != nil {
		shouldDefrag, err := a.evaluator.Evaluate(status.ToEvalVars(a.cfg.EtcdStorageQuotaBytes))
		if err != nil {
			return err
		}
		if !shouldDefrag {
			log.Printf("Evaluation result is false, so skipping endpoint: %s", endpoint)
			return nil
		}
	}

	if a.cfg.DryRun {
		log.Printf("[DRY-RUN] Would defragment endpoint %q", endpoint)
		return nil
	}

	// Perform defragmentation
	log.Printf("Defragmenting endpoint %q", endpoint)
	start := time.Now()

	cli, err := a.getClient(endpoint)
	if err != nil {
		return err
	}
	defer cli.Close()

	ctx, cancel := context.WithTimeout(ctx, a.cfg.CommandTimeout)
	defer cancel()

	_, err = cli.Defragment(ctx, endpoint)
	if err != nil {
		return err
	}

	log.Printf("Finished defragmenting etcd endpoint %q. took %v", endpoint, time.Since(start))

	// Log post-defrag status
	newStatus, err := a.getEndpointStatus(ctx, endpoint)
	if err == nil {
		log.Println("[Post defragmentation]")
		logStatus(endpoint, newStatus)
	}

	return nil
}

func logStatus(endpoint string, status *MemberStatus) {
	log.Printf("endpoint: %s, dbSize: %d, dbSizeInUse: %d, memberId: %x, leader: %x, revision: %d, term: %d, index: %d",
		endpoint, status.DbSize, status.DbSizeInUse, status.MemberID, status.LeaderID,
		status.Revision, status.RaftTerm, status.RaftIndex)
}
