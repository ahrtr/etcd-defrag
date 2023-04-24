package main

import (
	"testing"

	clientv3 "go.etcd.io/etcd/client/v3"
)

func TestValidateRule(t *testing.T) {
	testCases := []struct {
		name        string
		rule        string
		expectError bool
	}{
		{
			name:        "valid rule with comparison",
			rule:        "dbSize > dbQuota*80/100",
			expectError: false,
		},
		{
			name:        "valid rule with arithmetic operator",
			rule:        "dbSize - dbSizeInUse > 200*1024*1024",
			expectError: false,
		},
		{
			name:        "valid rule with logic OR",
			rule:        "dbSize > dbQuota*80/100 || dbSize - dbSizeInUse > 200*1024*1024",
			expectError: false,
		},
		{
			name:        "valid rule with logic AND",
			rule:        "dbSize > dbQuota*80/100 && dbSize - dbSizeInUse > 200*1024*1024",
			expectError: false,
		},
		{
			name:        "valid rule with parenthesis",
			rule:        "(dbSize > dbQuota*80/100) || (dbSize - dbSizeInUse > 200*1024*1024)",
			expectError: false,
		},
		{
			name:        "not a boolean expression",
			rule:        "dbSize - dbSizeInUse",
			expectError: true,
		},
		{
			name:        "invalid variable",
			rule:        "dbSizE > 100",
			expectError: true,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			err := validateRule(tc.rule)
			if tc.expectError != (err != nil) {
				t.Errorf("Unexpected result, expected error: %t, got %v", tc.expectError, err)
			}
		})
	}
}

func TestEvaluate(t *testing.T) {
	testCases := []struct {
		name             string
		rule             string
		expectError      bool
		dbQuota          int
		dbSize           int
		dbSizeInUse      int
		evaluationResult bool
	}{
		{
			name:        "not a boolean expression",
			rule:        "dbSize - dbSizeInUse",
			expectError: true,
		},
		{
			name:        "invalid variable",
			rule:        "dbSizE > 100",
			expectError: true,
		},
		{
			name:             "dbSize is greater than dbQuota*80/100",
			rule:             "dbSize > dbQuota*80/100",
			dbQuota:          100,
			dbSize:           81,
			evaluationResult: true,
		},
		{
			name:    "dbSize is less than dbQuota*80/100",
			rule:    "dbSize > dbQuota*80/100",
			dbQuota: 100,
			dbSize:  79,
		},
		{
			name:             "free space is greater than 200",
			rule:             "dbSize - dbSizeInUse > 200",
			dbSize:           301,
			dbSizeInUse:      100,
			evaluationResult: true,
		},
		{
			name:        "free space is less than 200",
			rule:        "dbSize - dbSizeInUse > 200",
			dbSize:      299,
			dbSizeInUse: 100,
		},
		{
			name:             "dbSize is greater than dbQuota*80/100 AND free space is greater than 200",
			rule:             "dbSize > dbQuota*80/100 && dbSize - dbSizeInUse > 200",
			dbQuota:          1000,
			dbSize:           801,
			dbSizeInUse:      600,
			evaluationResult: true,
		},
		{
			name:        "dbSize is greater than dbQuota*80/100 AND free space is less than 200",
			rule:        "dbSize > dbQuota*80/100 && dbSize - dbSizeInUse > 200",
			dbQuota:     100,
			dbSize:      81,
			dbSizeInUse: 60,
		},
		{
			name:             "dbSize is greater than dbQuota*80/100 OR free space is greater than 200",
			rule:             "dbSize > dbQuota*80/100 || dbSize - dbSizeInUse > 200",
			dbQuota:          1000,
			dbSize:           801,
			dbSizeInUse:      600,
			evaluationResult: true,
		},
		{
			name:             "dbSize is greater than dbQuota*80/100 AND free space is less than 200",
			rule:             "dbSize > dbQuota*80/100 || dbSize - dbSizeInUse > 200",
			dbQuota:          100,
			dbSize:           81,
			dbSizeInUse:      60,
			evaluationResult: true,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			gcfg := globalConfig{
				defragRule:   tc.rule,
				dbQuotaBytes: tc.dbQuota,
			}
			es := epStatus{
				Resp: &clientv3.StatusResponse{
					DbSize:      int64(tc.dbSize),
					DbSizeInUse: int64(tc.dbSizeInUse),
				},
			}
			ret, err := evaluate(gcfg, es)
			if tc.expectError != (err != nil) {
				t.Fatalf("Unexpected result, expected error: %t, got %v", tc.expectError, err)
			}
			if ret != tc.evaluationResult {
				t.Fatalf("Unexpected evaluation result, expected %t, got %t", tc.evaluationResult, ret)
			}
		})
	}
}
