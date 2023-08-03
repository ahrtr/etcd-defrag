package main

import (
	"context"
	"io"
	"time"

	"go.uber.org/zap"

	"go.etcd.io/etcd/client/pkg/v3/logutil"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func commandCtx(timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), timeout)
}

// EtcdCluster composes the interfaces needed to interact with the cluster. It mainly exist to enable testing.
type EtcdCluster interface {
	clientv3.Maintenance
	clientv3.Cluster
	clientv3.KV
	io.Closer
}

var createClient = func(cfgSpec *clientv3.ConfigSpec) (EtcdCluster, error) {
	lg, _ := logutil.CreateDefaultZapLogger(zap.InfoLevel)
	cfg, err := clientv3.NewClientConfig(cfgSpec, lg)
	if err != nil {
		return nil, err
	}
	return clientv3.New(*cfg)
}
