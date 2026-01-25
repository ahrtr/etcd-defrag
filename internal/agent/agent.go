// Package agent provides the core defragmentation logic
package agent

import (
	"context"
	"fmt"
	"log"

	clientv3 "go.etcd.io/etcd/client/v3"

	"github.com/ahrtr/etcd-defrag/internal/client"
	"github.com/ahrtr/etcd-defrag/internal/config"
	"github.com/ahrtr/etcd-defrag/internal/eval"
)

// Agent orchestrates the defragmentation process
type Agent struct {
	cfg       *config.GlobalConfig
	evaluator *eval.Evaluator
}

// New creates a new Agent
func New(cfg *config.GlobalConfig) (*Agent, error) {
	var evaluator *eval.Evaluator
	var err error

	if cfg.DefragRule != "" {
		evaluator, err = eval.New(cfg.DefragRule)
		if err != nil {
			return nil, fmt.Errorf("invalid defrag rule: %w", err)
		}
		log.Printf("Validating the defragmentation rule: %s ...", cfg.DefragRule)
		log.Println("valid")
	} else {
		log.Println("No defragmentation rule provided")
	}

	return &Agent{
		cfg:       cfg,
		evaluator: evaluator,
	}, nil
}

// Run executes the defragmentation workflow
func (a *Agent) Run(ctx context.Context) error {
	log.Println("Validating configuration.")

	// Validate configuration
	if errs := a.cfg.Validate(); len(errs) > 0 {
		for _, err := range errs {
			log.Printf("Configuration error: %v", err)
		}
		return errs[0]
	}

	// Get endpoints
	endpoints, err := a.resolveEndpoints(ctx)
	if err != nil {
		return err
	}

	// Health check
	if err := a.checkHealth(ctx, endpoints); err != nil {
		return err
	}

	// Get member status
	log.Println("Getting members status")
	statuses, err := a.getMemberStatus(ctx, endpoints)
	if err != nil {
		return err
	}

	// Run compaction
	if a.cfg.Compaction {
		if err := a.runCompaction(ctx, endpoints, statuses); err != nil {
			return err
		}
	}

	// Sort endpoints (leader last)
	endpoints = a.sortLeaderLast(endpoints, statuses)
	log.Printf("%d endpoint(s) need to be defragmented: %v", len(endpoints), endpoints)

	// Defragment
	if err := a.defragEndpoints(ctx, endpoints, statuses); err != nil {
		return err
	}

	// Auto-disalarm
	if a.cfg.AutoDisalarm {
		if err := a.autoDisalarm(ctx, endpoints, statuses); err != nil {
			return err
		}
	}

	log.Println("The defragmentation is successful.")
	return nil
}

// getClient creates an etcd client for the given endpoint
func (a *Agent) getClient(endpoint string) (*clientv3.Client, error) {
	return client.NewFromEndpoint(endpoint, &client.Config{
		Endpoints:         []string{endpoint},
		DialTimeout:       a.cfg.DialTimeout,
		KeepaliveTime:     a.cfg.KeepaliveTime,
		KeepaliveTimeout:  a.cfg.KeepaliveTimeout,
		CaCert:            a.cfg.CaCert,
		Cert:              a.cfg.Cert,
		Key:               a.cfg.Key,
		InsecureTransport: a.cfg.InsecureTransport,
		InsecureSkipVerify: a.cfg.InsecureSkipVerify,
		Username:          a.cfg.User,
		Password:          a.cfg.Password,
	})
}

// resolveEndpoints returns the list of endpoints to defragment
func (a *Agent) resolveEndpoints(ctx context.Context) ([]string, error) {
	// TODO: Implement cluster discovery, SRV discovery, etc.
	// For now, just return configured endpoints
	return a.cfg.Endpoints, nil
}

// sortLeaderLast sorts endpoints with leader at the end
func (a *Agent) sortLeaderLast(endpoints []string, statuses map[string]*MemberStatus) []string {
	var nonLeaders []string
	var leader string

	for _, ep := range endpoints {
		if status, ok := statuses[ep]; ok && status.IsLeader {
			leader = ep
		} else {
			nonLeaders = append(nonLeaders, ep)
		}
	}

	if leader != "" {
		return append(nonLeaders, leader)
	}
	return endpoints
}
