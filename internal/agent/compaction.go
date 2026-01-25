package agent

import (
	"context"
	"log"
)

// runCompaction performs compaction on the cluster
func (a *Agent) runCompaction(ctx context.Context, endpoints []string, statuses map[string]*MemberStatus) error {
	// Get the highest revision
	var maxRevision int64
	for _, status := range statuses {
		if status.Revision > maxRevision {
			maxRevision = status.Revision
		}
	}

	if maxRevision == 0 {
		log.Println("No revision found, skipping compaction")
		return nil
	}

	log.Printf("Running compaction until revision: %d ...", maxRevision)

	if a.cfg.DryRun {
		log.Printf("[DRY-RUN] Would compact until revision %d", maxRevision)
		log.Println("successful")
		return nil
	}

	// Use first endpoint for compaction (it's cluster-wide)
	cli, err := a.getClient(endpoints[0])
	if err != nil {
		return err
	}
	defer cli.Close()

	ctx, cancel := context.WithTimeout(ctx, a.cfg.CommandTimeout)
	defer cancel()

	_, err = cli.Compact(ctx, maxRevision)
	if err != nil {
		return err
	}

	log.Println("successful")
	return nil
}
