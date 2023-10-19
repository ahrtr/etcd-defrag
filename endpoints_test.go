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
