package endpoint

import (
	"context"

	clientv3 "go.etcd.io/etcd/client/v3"

	"github.com/ahrtr/etcd-defrag/internal/client"
	"github.com/ahrtr/etcd-defrag/internal/config"
)

// Manager handles endpoint resolution and management
type Manager struct {
	cfg *config.GlobalConfig
}

// NewManager creates a new endpoint manager
func NewManager(cfg *config.GlobalConfig) *Manager {
	return &Manager{cfg: cfg}
}

// Resolve returns the list of endpoints based on configuration
func (m *Manager) Resolve(ctx context.Context) ([]string, error) {
	if !m.cfg.Cluster {
		return m.endpointsFromCmd()
	}
	return m.endpointsFromCluster(ctx)
}

// endpointsFromCmd returns endpoints from command-line or DNS discovery
func (m *Manager) endpointsFromCmd() ([]string, error) {
	eps, err := m.endpointsFromDNSDiscovery()
	if err != nil {
		return nil, err
	}

	if len(eps) == 0 {
		eps = m.cfg.Endpoints
	}

	if len(eps) == 0 {
		return nil, ErrNoEndpoints
	}

	return eps, nil
}

// endpointsFromCluster gets endpoints from cluster member list
func (m *Manager) endpointsFromCluster(ctx context.Context) ([]string, error) {
	memberlistResp, err := m.memberList(ctx)
	if err != nil {
		return nil, err
	}

	return m.extractEndpoints(memberlistResp)
}

// memberList retrieves the cluster member list
func (m *Manager) memberList(ctx context.Context) (*clientv3.MemberListResponse, error) {
	cli, err := client.New(&client.Config{
		Endpoints:          m.cfg.Endpoints,
		DialTimeout:        m.cfg.DialTimeout,
		KeepaliveTime:      m.cfg.KeepaliveTime,
		KeepaliveTimeout:   m.cfg.KeepaliveTimeout,
		CaCert:             m.cfg.CaCert,
		Cert:               m.cfg.Cert,
		Key:                m.cfg.Key,
		InsecureTransport:  m.cfg.InsecureTransport,
		InsecureSkipVerify: m.cfg.InsecureSkipVerify,
		Username:           m.cfg.User,
		Password:           m.cfg.Password,
	})
	if err != nil {
		return nil, err
	}
	defer cli.Close()

	return cli.MemberList(ctx)
}
