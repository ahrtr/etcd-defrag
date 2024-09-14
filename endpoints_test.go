package main

import (
	"context"

	"testing"

	"go.etcd.io/etcd/api/v3/etcdserverpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"golang.org/x/exp/slices"
)

func TestEndpointDedup(t *testing.T) {
	oldCreateClient := createClient
	t.Cleanup(func() {
		createClient = oldCreateClient
	})

	fakeClient := fakeClientURLClient{}
	createClient = func(cfgSpec *clientv3.ConfigSpec) (EtcdCluster, error) {
		return &fakeClient, nil
	}

	testcases := []struct {
		name               string
		returnedMemberList *clientv3.MemberListResponse
		expectedEndpoints  []string
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
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			fakeClient.memberListResp = testcase.returnedMemberList
			ep, err := endpointsFromCluster(globalConfig{endpoints: []string{"https://localhost:2379"}})
			if err != nil {
				t.Error(err)
			}

			if !slices.Equal(testcase.expectedEndpoints, ep) {
				t.Errorf("endpoints didn't match. Expected %v got %v", testcase.expectedEndpoints, ep)
			}
		})
	}
}

func TestEndpointExcludeLocalhost(t *testing.T) {
	oldCreateClient := createClient
	t.Cleanup(func() {
		createClient = oldCreateClient
	})

	fakeClient := fakeClientURLClient{}
	createClient = func(cfgSpec *clientv3.ConfigSpec) (EtcdCluster, error) {
		return &fakeClient, nil
	}

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
						ClientURLs: []string{"etcd2.example.com:2379", "127.0.0.1:2379"},
					},
					{
						ClientURLs: []string{"etcd3.example.com:2379"},
					},
				},
			},
			[]string{"127.0.0.1:2379", "etcd2.example.com:2379", "etcd3.example.com:2379"},
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
			fakeClient.memberListResp = testcase.returnedMemberList
			ep, err := endpointsFromCluster(globalConfig{endpoints: []string{"https://localhost:2379"}, excludeLocalhost: testcase.excludeLocalhost})
			if err != nil {
				t.Error(err)
			}

			if !slices.Equal(testcase.expectedEndpoints, ep) {
				t.Errorf("endpoints didn't match. Expected %v got %v", testcase.expectedEndpoints, ep)
			}
		})
	}
}

type fakeClientURLClient struct {
	*clientv3.Client
	memberListResp *clientv3.MemberListResponse
}

// MemberList lists the current cluster membership.
func (f fakeClientURLClient) MemberList(ctx context.Context, opts ...clientv3.OpOption) (*clientv3.MemberListResponse, error) {
	return f.memberListResp, nil
}

func (fakeClientURLClient) Close() error {
	return nil
}

func TestIsLocalEp(t *testing.T) {
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
			if ok, err := isLocalEndpoint(testcase.ep); err != testcase.err || ok != testcase.desire {
				t.Errorf("expected %v, got err: %v result: %v", testcase.desire, err, ok)
			}
		})
	}
}
