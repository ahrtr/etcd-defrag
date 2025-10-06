package main

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.etcd.io/etcd/api/v3/etcdserverpb"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func TestEndpointsForHealthCheck_SkipHealthcheckClusterEndpoints(t *testing.T) {
	oldCreateClient := createClient
	t.Cleanup(func() {
		createClient = oldCreateClient
	})

	testCases := []struct {
		name                            string
		skipHealthcheckClusterEndpoints bool
		endpoints                       []string
		memberListResponse              *clientv3.MemberListResponse
		expectedEndpoints               []string
	}{
		{
			name:                            "flag disabled returns cluster endpoints",
			skipHealthcheckClusterEndpoints: false,
			endpoints:                       []string{"192.168.1.10:2379"},
			memberListResponse: &clientv3.MemberListResponse{
				Members: []*etcdserverpb.Member{
					{ClientURLs: []string{"192.168.1.10:2379"}},
					{ClientURLs: []string{"192.168.1.11:2379"}},
					{ClientURLs: []string{"192.168.1.12:2379"}},
				},
			},
			expectedEndpoints: []string{"192.168.1.10:2379", "192.168.1.11:2379", "192.168.1.12:2379"},
		},
		{
			name:                            "flag enabled returns provided endpoints only",
			skipHealthcheckClusterEndpoints: true,
			endpoints:                       []string{"192.168.1.10:2379"},
			memberListResponse: &clientv3.MemberListResponse{
				Members: []*etcdserverpb.Member{
					{ClientURLs: []string{"192.168.1.10:2379"}},
					{ClientURLs: []string{"192.168.1.11:2379"}},
					{ClientURLs: []string{"192.168.1.12:2379"}},
				},
			},
			expectedEndpoints: []string{"192.168.1.10:2379"},
		},
		{
			name:                            "flag enabled with multiple endpoints",
			skipHealthcheckClusterEndpoints: true,
			endpoints:                       []string{"192.168.1.10:2379", "192.168.1.11:2379"},
			memberListResponse: &clientv3.MemberListResponse{
				Members: []*etcdserverpb.Member{
					{ClientURLs: []string{"192.168.1.10:2379"}},
					{ClientURLs: []string{"192.168.1.11:2379"}},
					{ClientURLs: []string{"192.168.1.12:2379"}},
				},
			},
			expectedEndpoints: []string{"192.168.1.10:2379", "192.168.1.11:2379"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fakeClient := &fakeHealthCheckClient{
				memberListResp: tc.memberListResponse,
			}

			createClient = func(cfgSpec *clientv3.ConfigSpec) (EtcdCluster, error) {
				return fakeClient, nil
			}

			cfg := globalConfig{
				skipHealthcheckClusterEndpoints: tc.skipHealthcheckClusterEndpoints,
				endpoints:                       tc.endpoints,
			}

			var actualEndpoints []string
			var err error

			if cfg.skipHealthcheckClusterEndpoints {
				actualEndpoints, err = endpointsFromCmd(cfg)
			} else {
				actualEndpoints, err = endpointsFromCluster(cfg)
			}

			require.NoError(t, err)
			require.ElementsMatch(t, tc.expectedEndpoints, actualEndpoints)
		})
	}
}

func TestEndpointsFromCluster_ExcludesLearners(t *testing.T) {
	oldCreateClient := createClient
	t.Cleanup(func() {
		createClient = oldCreateClient
	})

	testCases := []struct {
		name               string
		memberListResponse *clientv3.MemberListResponse
		expectedEndpoints  []string
	}{
		{
			name: "includes only voting members",
			memberListResponse: &clientv3.MemberListResponse{
				Members: []*etcdserverpb.Member{
					{ClientURLs: []string{"192.168.1.10:2379"}, IsLearner: false},
					{ClientURLs: []string{"192.168.1.11:2379"}, IsLearner: true},
					{ClientURLs: []string{"192.168.1.12:2379"}, IsLearner: false},
				},
			},
			expectedEndpoints: []string{"192.168.1.10:2379", "192.168.1.12:2379"},
		},
		{
			name: "all learners excluded",
			memberListResponse: &clientv3.MemberListResponse{
				Members: []*etcdserverpb.Member{
					{ClientURLs: []string{"192.168.1.10:2379"}, IsLearner: true},
					{ClientURLs: []string{"192.168.1.11:2379"}, IsLearner: true},
				},
			},
			expectedEndpoints: []string{},
		},
		{
			name: "all voting members included",
			memberListResponse: &clientv3.MemberListResponse{
				Members: []*etcdserverpb.Member{
					{ClientURLs: []string{"192.168.1.10:2379"}, IsLearner: false},
					{ClientURLs: []string{"192.168.1.11:2379"}, IsLearner: false},
				},
			},
			expectedEndpoints: []string{"192.168.1.10:2379", "192.168.1.11:2379"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fakeClient := &fakeHealthCheckClient{
				memberListResp: tc.memberListResponse,
			}

			createClient = func(cfgSpec *clientv3.ConfigSpec) (EtcdCluster, error) {
				return fakeClient, nil
			}

			cfg := globalConfig{
				endpoints: []string{"192.168.1.10:2379"},
			}

			actualEndpoints, err := endpointsFromCluster(cfg)
			require.NoError(t, err)
			require.ElementsMatch(t, tc.expectedEndpoints, actualEndpoints)
		})
	}
}

func TestCheckAllMembersDBSize(t *testing.T) {
	testCases := []struct {
		name        string
		gcfg        globalConfig
		statusList  []epStatus
		expectedEps []string
	}{
		{
			name: "all members below threshold",
			gcfg: globalConfig{
				dbQuotaBytes:      1000,
				disalarmThreshold: 0.8,
			},
			statusList: []epStatus{
				{Ep: "ep1", Resp: &clientv3.StatusResponse{DbSize: 700}},
				{Ep: "ep2", Resp: &clientv3.StatusResponse{DbSize: 600}},
				{Ep: "ep3", Resp: &clientv3.StatusResponse{DbSize: 500}},
			},
			expectedEps: nil,
		},
		{
			name: "some members above threshold",
			gcfg: globalConfig{
				dbQuotaBytes:      1000,
				disalarmThreshold: 0.8,
			},
			statusList: []epStatus{
				{Ep: "ep1", Resp: &clientv3.StatusResponse{DbSize: 900}},
				{Ep: "ep2", Resp: &clientv3.StatusResponse{DbSize: 850}},
				{Ep: "ep3", Resp: &clientv3.StatusResponse{DbSize: 500}},
			},
			expectedEps: []string{"ep1", "ep2"},
		},
		{
			name: "all members above threshold",
			gcfg: globalConfig{
				dbQuotaBytes:      1000,
				disalarmThreshold: 0.5,
			},
			statusList: []epStatus{
				{Ep: "ep1", Resp: &clientv3.StatusResponse{DbSize: 600}},
				{Ep: "ep2", Resp: &clientv3.StatusResponse{DbSize: 700}},
				{Ep: "ep3", Resp: &clientv3.StatusResponse{DbSize: 800}},
			},
			expectedEps: []string{"ep1", "ep2", "ep3"},
		},
		{
			name: "empty status list",
			gcfg: globalConfig{
				dbQuotaBytes:      1000,
				disalarmThreshold: 0.8,
			},
			statusList:  []epStatus{},
			expectedEps: nil,
		},
		{
			name: "zero threshold",
			gcfg: globalConfig{
				dbQuotaBytes:      1000,
				disalarmThreshold: 0.0,
			},
			statusList: []epStatus{
				{Ep: "ep1", Resp: &clientv3.StatusResponse{DbSize: 100}},
				{Ep: "ep2", Resp: &clientv3.StatusResponse{DbSize: 50}},
			},
			expectedEps: []string{"ep1", "ep2"},
		},
		{
			name: "threshold at boundary",
			gcfg: globalConfig{
				dbQuotaBytes:      1000,
				disalarmThreshold: 0.8,
			},
			statusList: []epStatus{
				{Ep: "ep1", Resp: &clientv3.StatusResponse{DbSize: 800}}, // exactly at threshold
				{Ep: "ep2", Resp: &clientv3.StatusResponse{DbSize: 801}}, // just above threshold
			},
			expectedEps: []string{"ep2"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			eps := checkAllMembersDBSize(tc.gcfg, tc.statusList)

			require.ElementsMatch(t, tc.expectedEps, eps, "unexpected endpoints above threshold")
		})
	}
}

type fakeHealthCheckClient struct {
	*clientv3.Client
	memberListResp *clientv3.MemberListResponse
}

func (f *fakeHealthCheckClient) MemberList(ctx context.Context, opts ...clientv3.OpOption) (*clientv3.MemberListResponse, error) {
	return f.memberListResp, nil
}

func (f *fakeHealthCheckClient) Close() error {
	return nil
}
