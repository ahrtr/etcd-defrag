package agent

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"
)

// checkHealth performs health checks on all endpoints
func (a *Agent) checkHealth(ctx context.Context, endpoints []string) error {
	log.Println("Performing health check.")

	for _, ep := range endpoints {
		if err := a.checkEndpointHealth(ctx, ep); err != nil {
			return err
		}
	}
	return nil
}

func (a *Agent) checkEndpointHealth(ctx context.Context, endpoint string) error {
	start := time.Now()

	cli, err := a.getClient(endpoint)
	if err != nil {
		log.Printf("endpoint: %s, health: false, took: %v, error: %v", endpoint, time.Since(start), err)
		return fmt.Errorf("failed to create client for %s: %w", endpoint, err)
	}
	defer cli.Close()

	ctx, cancel := context.WithTimeout(ctx, a.cfg.CommandTimeout)
	defer cancel()

	_, err = cli.Status(ctx, endpoint)
	if err != nil {
		// Check if it's NOSPACE alarm (which we ignore)
		if isNospaceError(err) {
			log.Printf("endpoint: %s, health: true (NOSPACE alarm ignored), took: %v, error: ", endpoint, time.Since(start))
			return nil
		}
		log.Printf("endpoint: %s, health: false, took: %v, error: %v", endpoint, time.Since(start), err)
		return fmt.Errorf("endpoint %s is unhealthy: %w", endpoint, err)
	}

	log.Printf("endpoint: %s, health: true, took: %v, error: ", endpoint, time.Since(start))
	return nil
}

func isNospaceError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "NOSPACE")
}
