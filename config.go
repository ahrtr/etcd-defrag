package main

import (
	"strings"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

type globalConfig struct {
	endpoints []string

	dialTimeout      time.Duration
	commandTimeout   time.Duration
	keepAliveTime    time.Duration
	keepAliveTimeout time.Duration

	insecure           bool
	insecureDiscovery  bool
	insecureSkepVerify bool

	certFile string
	keyFile  string
	caFile   string

	dnsDomain  string
	dnsService string

	username string
	password string

	useClusterEndpoints bool
	continueOnError     bool
}

func clientConfigWithoutEndpoints(gcfg globalConfig) *clientv3.ConfigSpec {
	cfg := &clientv3.ConfigSpec{
		DialTimeout:      gcfg.dialTimeout,
		KeepAliveTime:    gcfg.keepAliveTime,
		KeepAliveTimeout: gcfg.keepAliveTimeout,

		Secure: secureConfig(gcfg),
		Auth:   authConfig(gcfg),
	}

	return cfg
}

func secureConfig(gcfg globalConfig) *clientv3.SecureConfig {
	scfg := &clientv3.SecureConfig{
		Cert:       gcfg.certFile,
		Key:        gcfg.keyFile,
		Cacert:     gcfg.caFile,
		ServerName: gcfg.dnsDomain,

		InsecureTransport:  gcfg.insecure,
		InsecureSkipVerify: gcfg.insecureSkepVerify,
	}

	if gcfg.insecureDiscovery {
		scfg.ServerName = ""
	}

	return scfg
}

func authConfig(gcfg globalConfig) *clientv3.AuthConfig {
	if gcfg.username == "" {
		return nil
	}

	if gcfg.password == "" {
		userSecret := strings.SplitN(gcfg.username, ":", 2)
		if len(userSecret) < 2 {
			return nil
		}

		return &clientv3.AuthConfig{
			Username: userSecret[0],
			Password: userSecret[1],
		}
	}

	return &clientv3.AuthConfig{
		Username: gcfg.username,
		Password: gcfg.password,
	}
}
