package main

import (
	"github.com/maja42/goval"
)

const (
	dbSize      = "dbSize"
	dbSizeInUse = "dbSizeInUse"
	dbQuota     = "dbQuota"
)

func evaluate(gcfg globalConfig, es epStatus) (bool, error) {
	if len(gcfg.defragRules) == 0 {
		return true, nil
	}

	variables := map[string]interface{}{
		dbQuota:     gcfg.dbQuotaBytes,
		dbSize:      int(es.Resp.DbSize),
		dbSizeInUse: int(es.Resp.DbSizeInUse),
	}
	eval := goval.NewEvaluator()
	for _, rule := range gcfg.defragRules {
		resultRaw, err := eval.Evaluate(rule, variables, nil)
		result := resultRaw.(bool)
		if result || err != nil {
			return result, err
		}
	}

	return false, nil
}
