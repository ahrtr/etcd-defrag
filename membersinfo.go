package main

import (
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"

	"go.etcd.io/etcd/api/v3/etcdserverpb"
	"go.etcd.io/etcd/api/v3/v3rpc/rpctypes"
	"go.etcd.io/etcd/client/pkg/v3/logutil"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func memberList(gcfg globalConfig) (*clientv3.MemberListResponse, error) {
	cfgSpec := clientConfigWithoutEndpoints(gcfg)
	eps, err := endpointsFromCmd(gcfg)
	if err != nil {
		return nil, err
	}
	cfgSpec.Endpoints = eps

	c, err := createClient(cfgSpec)
	if err != nil {
		return nil, err
	}

	ctx, cancel := commandCtx(gcfg.commandTimeout)
	defer func() {
		c.Close()
		cancel()
	}()

	members, err := c.MemberList(ctx)
	if err != nil {
		return nil, err
	}

	return members, nil
}

type epHealth struct {
	Ep     string `json:"endpoint"`
	Health bool   `json:"health"`
	Took   string `json:"took"`
	Error  string `json:"error,omitempty"`
}

func (eh epHealth) String() string {
	return fmt.Sprintf("endpoint: %s, health: %t, took: %s, error: %s", eh.Ep, eh.Health, eh.Took, eh.Error)
}

func clusterHealth(gcfg globalConfig) ([]epHealth, error) {
	eps, err := endpointsFromCluster(gcfg)
	if err != nil {
		return nil, err
	}
	lg, err := logutil.CreateDefaultZapLogger(zap.InfoLevel)
	if err != nil {
		return nil, err
	}

	cfgSpec := clientConfigWithoutEndpoints(gcfg)
	var cfgs []*clientv3.Config
	for _, ep := range eps {
		cfg, err := clientv3.NewClientConfig(&clientv3.ConfigSpec{
			Endpoints:        []string{ep},
			DialTimeout:      cfgSpec.DialTimeout,
			KeepAliveTime:    cfgSpec.KeepAliveTime,
			KeepAliveTimeout: cfgSpec.KeepAliveTimeout,
			Secure:           cfgSpec.Secure,
			Auth:             cfgSpec.Auth,
		}, lg)
		if err != nil {
			return nil, err
		}
		cfgs = append(cfgs, cfg)
	}

	healthCh := make(chan epHealth, len(eps))

	var wg sync.WaitGroup
	for _, cfg := range cfgs {
		wg.Add(1)
		go func(cfg *clientv3.Config) {
			defer wg.Done()

			ep := cfg.Endpoints[0]
			cfg.Logger = lg.Named("client")
			cli, err := clientv3.New(*cfg)
			if err != nil {
				healthCh <- epHealth{Ep: ep, Health: false, Error: err.Error()}
				return
			}
			startTs := time.Now()
			// get a random key. As long as we can get the response
			// without an error, the endpoint is health.
			ctx, cancel := commandCtx(gcfg.commandTimeout)
			_, err = cli.Get(ctx, "health")
			eh := epHealth{Ep: ep, Health: false, Took: time.Since(startTs).String()}
			if err == nil || err == rpctypes.ErrPermissionDenied {
				eh.Health = true
			} else {
				eh.Error = err.Error()
			}

			if eh.Health {
				resp, err := cli.AlarmList(ctx)
				if err == nil && len(resp.Alarms) > 0 {
					eh.Health = false
					eh.Error = "Active Alarm(s): "
					for _, v := range resp.Alarms {
						switch v.Alarm {
						case etcdserverpb.AlarmType_NOSPACE:
							// We ignore AlarmType_NOSPACE, and we need to
							// continue to perform defragmentation.
							eh.Health = true
							eh.Error = eh.Error + "NOSPACE "
						case etcdserverpb.AlarmType_CORRUPT:
							eh.Error = eh.Error + "CORRUPT "
						default:
							eh.Error = eh.Error + "UNKNOWN "
						}
					}
				} else if err != nil {
					eh.Health = false
					eh.Error = "Unable to fetch the alarm list"
				}
			}
			cancel()
			healthCh <- eh
		}(cfg)
	}
	wg.Wait()
	close(healthCh)

	var healthList []epHealth
	for h := range healthCh {
		healthList = append(healthList, h)
	}

	return healthList, nil
}

type epStatus struct {
	Ep   string                   `json:"Endpoint"`
	Resp *clientv3.StatusResponse `json:"Status"`
}

func memberStatus(gcfg globalConfig) ([]epStatus, error) {
	eps, err := endpoints(gcfg)
	if err != nil {
		return nil, err
	}

	cfgSpec := clientConfigWithoutEndpoints(gcfg)

	var statusList []epStatus
	for _, ep := range eps {
		cfgSpec.Endpoints = []string{ep}
		c, err := createClient(cfgSpec)
		if err != nil {
			return nil, err
		}

		ctx, cancel := commandCtx(gcfg.commandTimeout)
		resp, err := c.Status(ctx, ep)
		cancel()
		c.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to get member(%q) status: %w", ep, err)
		}
		statusList = append(statusList, epStatus{Ep: ep, Resp: resp})
	}

	return statusList, nil
}
