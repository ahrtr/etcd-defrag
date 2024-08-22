package main

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"strings"

	"go.etcd.io/etcd/client/pkg/v3/srv"
	"golang.org/x/exp/slices"
)

var errBadScheme = errors.New("url scheme must be http or https")

func endpointsWithLeaderAtEnd(gcfg globalConfig, statusList []epStatus) ([]string, error) {
	eps, err := endpoints(gcfg)
	if err != nil || len(eps) <= 1 {
		return eps, err
	}

	var sortedEps, leaderEps []string
	for _, status := range statusList {
		if status.Resp.Header.MemberId != status.Resp.Leader {
			sortedEps = append(sortedEps, status.Ep)
		} else {
			leaderEps = append(leaderEps, status.Ep)
		}
	}

	sortedEps = append(sortedEps, leaderEps...)
	return sortedEps, nil
}

func endpoints(gcfg globalConfig) ([]string, error) {
	if !gcfg.useClusterEndpoints {
		if len(gcfg.endpoints) == 0 {
			return nil, errors.New("no endpoints provided")
		}
		return gcfg.endpoints, nil
	}

	return endpointsFromCluster(gcfg)
}

func IsLocalEndpoint(ep string) (bool, error) {

	if strings.HasPrefix(ep, "unix:") || strings.HasPrefix(ep, "unixs:") {
		return true, nil
	}

	if strings.Contains(ep, "://") {
		url, err := url.Parse(ep)
		if err != nil {
			return false, err
		}
		if url.Scheme != "http" && url.Scheme != "https" {
			return false, errBadScheme
		}

		return isLocalEndpoint(url.Host)
	}

	return isLocalEndpoint(ep)
}

func isLocalEndpoint(ep string) (bool, error) {

	hostStr, _, err := net.SplitHostPort(ep)
	if err != nil {
		return false, err
	}

	// lookup localhost
	addrs, err := net.LookupHost(hostStr)
	if err != nil {
		return false, nil
	}

	for _, addr := range addrs {
		ip := net.ParseIP(addr)
		if ip != nil && ip.IsLoopback() {
			return true, nil
		}
	}
	return false, nil
}

func endpointsFromCluster(gcfg globalConfig) ([]string, error) {
	memberlistResp, err := memberList(gcfg)
	if err != nil {
		return nil, err
	}

	var eps []string
	for _, m := range memberlistResp.Members {
		// learner member only serves Status and SerializableRead requests, just ignore it
		if !m.GetIsLearner() {
			for _, ep := range m.ClientURLs {
				// not append local endpoint when set excludeLocalhost
				if gcfg.excludeLocalhost {
					ok, err := IsLocalEndpoint(ep)
					if err != nil {
						return nil, err
					}
					if ok {
						continue
					}
				}
				eps = append(eps, ep)
			}
		}
	}

	slices.Sort(eps)
	eps = slices.Compact(eps)

	return eps, nil
}

func endpointsFromCmd(gcfg globalConfig) ([]string, error) {
	eps, err := endpointsFromDNSDiscovery(gcfg)
	if err != nil {
		return nil, err
	}

	if len(eps) == 0 {
		eps = gcfg.endpoints
	}

	if len(eps) == 0 {
		return nil, errors.New("no endpoints provided")
	}

	return eps, nil
}

func endpointsFromDNSDiscovery(gcfg globalConfig) ([]string, error) {
	if gcfg.dnsDomain == "" {
		return nil, nil
	}

	srvs, err := srv.GetClient("etcd-client", gcfg.dnsDomain, gcfg.dnsService)
	if err != nil {
		return nil, err
	}

	eps := srvs.Endpoints
	if gcfg.insecureDiscovery {
		return eps, nil
	}

	// strip insecure connections
	var ret []string
	for _, ep := range eps {
		if strings.HasPrefix(ep, "http://") {
			fmt.Fprintf(os.Stderr, "ignoring discovered insecure endpoint %q\n", ep)
			continue
		}
		ret = append(ret, ep)
	}
	return ret, nil
}
