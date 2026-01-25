package endpoint

import (
	"errors"
	"log"
	"net"
	"net/url"
	"strings"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/pkg/v3/srv"
	"golang.org/x/exp/slices"
)

var (
	errBadScheme   = errors.New("url scheme must be http or https")
	ErrNoEndpoints = errors.New("no endpoints provided")
)

// extractEndpoints extracts and filters endpoints from member list response
func (m *Manager) extractEndpoints(memberlistResp *clientv3.MemberListResponse) ([]string, error) {
	var eps []string
	for _, member := range memberlistResp.Members {
		// learner member only serves Status and SerializableRead requests, just ignore it
		if !member.GetIsLearner() {
			for _, ep := range member.ClientURLs {
				// Do not append loopback endpoints when `--exclude-localhost` is set.
				if m.cfg.ExcludeLocalhost {
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

// endpointsFromDNSDiscovery discovers endpoints via DNS SRV records
func (m *Manager) endpointsFromDNSDiscovery() ([]string, error) {
	if m.cfg.DiscoverySrv == "" {
		return nil, nil
	}

	srvs, err := srv.GetClient("etcd-client", m.cfg.DiscoverySrv, m.cfg.DiscoverySrvName)
	if err != nil {
		return nil, err
	}

	eps := srvs.Endpoints
	if m.cfg.InsecureDiscovery {
		return eps, nil
	}

	// strip insecure connections
	var ret []string
	for _, ep := range eps {
		if strings.HasPrefix(ep, "http://") {
			log.Printf("ignoring discovered insecure endpoint %q\n", ep)
			continue
		}
		ret = append(ret, ep)
	}
	return ret, nil
}

// IsLocalEndpoint checks if an endpoint is a local/loopback address
func IsLocalEndpoint(ep string) (bool, error) {
	if strings.HasPrefix(ep, "unix:") || strings.HasPrefix(ep, "unixs:") {
		return true, nil
	}

	hostPort := ep
	if strings.Contains(ep, "://") {
		url, err := url.Parse(ep)
		if err != nil {
			return false, err
		}
		if url.Scheme != "http" && url.Scheme != "https" {
			return false, errBadScheme
		}

		hostPort = url.Host
	}

	hostname, _, err := net.SplitHostPort(hostPort)
	if err != nil {
		return false, err
	}

	if strings.EqualFold(hostname, "localhost") {
		return true, nil
	}

	ip := net.ParseIP(hostname)
	if ip != nil && ip.IsLoopback() {
		return true, nil
	}

	return false, nil
}
