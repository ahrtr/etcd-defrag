package endpoint

import (
	"testing"

	"go.etcd.io/etcd/api/v3/etcdserverpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"golang.org/x/exp/slices"

	"github.com/ahrtr/etcd-defrag/internal/config"
)

func TestExtractEndpoints(t *testing.T) {
	testcases := []struct {
		name               string
		returnedMemberList *clientv3.MemberListResponse
		expectedEndpoints  []string
		excludeLocalhost   bool
	}{
		{
			"normal",
			&clientv3.MemberListResponse{
				Members: []*etcdserverpb.Member{
					{
						ClientURLs: []string{"etcd1.example.com:2379"},
					},
					{
						ClientURLs: []string{"etcd2.example.com:2379"},
					},
					{
						ClientURLs: []string{"etcd3.example.com:2379"},
					},
				},
			},
			[]string{"etcd1.example.com:2379", "etcd2.example.com:2379", "etcd3.example.com:2379"},
			false,
		},
		{
			"sort",
			&clientv3.MemberListResponse{
				Members: []*etcdserverpb.Member{
					{
						ClientURLs: []string{"etcd1.example.com:2379"},
					},
					{
						ClientURLs: []string{"etcd3.example.com:2379"},
					},
					{
						ClientURLs: []string{"etcd2.example.com:2379"},
					},
				},
			},
			[]string{"etcd1.example.com:2379", "etcd2.example.com:2379", "etcd3.example.com:2379"},
			false,
		},
		{
			"uniq",
			&clientv3.MemberListResponse{
				Members: []*etcdserverpb.Member{
					{
						ClientURLs: []string{"etcd1.example.com:2379", "etcd.example.com:2379"},
					},
					{
						ClientURLs: []string{"etcd3.example.com:2379", "etcd.example.com:2379"},
					},
					{
						ClientURLs: []string{"etcd2.example.com:2379", "etcd.example.com:2379"},
					},
				},
			},
			[]string{"etcd.example.com:2379", "etcd1.example.com:2379", "etcd2.example.com:2379", "etcd3.example.com:2379"},
			false,
		},
		{
			"ignore learner",
			&clientv3.MemberListResponse{
				Members: []*etcdserverpb.Member{
					{
						ClientURLs: []string{"etcd1.example.com:2379"},
						IsLearner:  true,
					},
					{
						ClientURLs: []string{"etcd3.example.com:2379"},
					},
					{
						ClientURLs: []string{"etcd2.example.com:2379"},
					},
				},
			},
			[]string{"etcd2.example.com:2379", "etcd3.example.com:2379"},
			false,
		},
		{
			"excludeLocalhost",
			&clientv3.MemberListResponse{
				Members: []*etcdserverpb.Member{
					{
						ClientURLs: []string{"etcd2.example.com:2379", "127.0.0.1:2379"},
					},
					{
						ClientURLs: []string{"etcd3.example.com:2379"},
					},
				},
			},
			[]string{"etcd2.example.com:2379", "etcd3.example.com:2379"},
			true,
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			mgr := NewManager(&config.GlobalConfig{
				Endpoints:        []string{"https://localhost:2379"},
				ExcludeLocalhost: testcase.excludeLocalhost,
			})

			ep, err := mgr.extractEndpoints(testcase.returnedMemberList)
			if err != nil {
				t.Error(err)
			}

			if !slices.Equal(testcase.expectedEndpoints, ep) {
				t.Errorf("endpoints didn't match. Expected %v got %v", testcase.expectedEndpoints, ep)
			}
		})
	}
}

func TestIsLocalEndpoint(t *testing.T) {
	testcases := []struct {
		name   string
		ep     string
		desire bool
		err    error
	}{
		{
			"ipv4 loopback address",
			"127.0.0.1:2379",
			true,
			nil,
		},
		{
			"ipv4 non-loopback address",
			"10.7.7.7:2379",
			false,
			nil,
		},
		{
			"http url with ipv4 loopback address",
			"http://127.0.0.1:2379",
			true,
			nil,
		},
		{
			"http url with ipv4 non-loopback address",
			"http://10.7.7.7:2379",
			false,
			nil,
		},
		{
			"https url with hostname",
			"https://abc-0.ns1-etcd.ns1.svc.cluster.local.:2379",
			false,
			nil,
		},
		{
			"ipv6 abbreviated loopback address",
			"[::1]:2379",
			true,
			nil,
		},
		{
			"ipv6 loopback address",
			"[0:0:0:0:0:0:0:1]:2379",
			true,
			nil,
		},
		{
			"ipv6 non-loopback address",
			"[2007:0db8:3c4d:0015:0000:0000:1a2f:1a2b]:2379",
			false,
			nil,
		},
		{
			"localhost hostname",
			"localhost:2379",
			true,
			nil,
		},
		{
			"https url with localhost hostname",
			"https://localhost:2379",
			true,
			nil,
		},
		{
			"url with bad scheme",
			"abc://localhost:2379",
			false,
			errBadScheme,
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			if ok, err := IsLocalEndpoint(testcase.ep); err != testcase.err || ok != testcase.desire {
				t.Errorf("expected %v, got err: %v result: %v", testcase.desire, err, ok)
			}
		})
	}
}
