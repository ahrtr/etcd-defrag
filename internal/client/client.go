// Package client provides etcd client utilities
package client

import (
	"crypto/tls"
	"time"

	"go.etcd.io/etcd/client/pkg/v3/transport"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// Config holds the configuration for creating an etcd client
type Config struct {
	Endpoints        []string
	DialTimeout      time.Duration
	KeepaliveTime    time.Duration
	KeepaliveTimeout time.Duration

	// TLS
	CaCert             string
	Cert               string
	Key                string
	InsecureTransport  bool
	InsecureSkipVerify bool

	// Auth
	Username string
	Password string
}

// New creates a new etcd client with the given configuration
func New(cfg *Config) (*clientv3.Client, error) {
	clientCfg := clientv3.Config{
		Endpoints:            cfg.Endpoints,
		DialTimeout:          cfg.DialTimeout,
		DialKeepAliveTime:    cfg.KeepaliveTime,
		DialKeepAliveTimeout: cfg.KeepaliveTimeout,
	}

	// Setup TLS
	var tlsConfig *tls.Config
	var err error

	if cfg.CaCert != "" || cfg.Cert != "" || cfg.Key != "" {
		tlsInfo := transport.TLSInfo{
			TrustedCAFile: cfg.CaCert,
			CertFile:      cfg.Cert,
			KeyFile:       cfg.Key,
		}
		tlsConfig, err = tlsInfo.ClientConfig()
		if err != nil {
			return nil, err
		}
	}

	if cfg.InsecureSkipVerify && tlsConfig != nil {
		tlsConfig.InsecureSkipVerify = true
	}

	if !cfg.InsecureTransport {
		clientCfg.TLS = tlsConfig
	}

	// Setup auth
	if cfg.Username != "" {
		clientCfg.Username = cfg.Username
		clientCfg.Password = cfg.Password
	}

	return clientv3.New(clientCfg)
}

// NewFromEndpoint creates a client for a single endpoint
func NewFromEndpoint(endpoint string, baseCfg *Config) (*clientv3.Client, error) {
	cfg := *baseCfg
	cfg.Endpoints = []string{endpoint}
	return New(&cfg)
}
