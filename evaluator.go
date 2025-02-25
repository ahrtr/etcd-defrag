package main

import (
	"errors"

	"github.com/maja42/goval"
)

const (
	dbSize       = "dbSize"
	dbSizeInUse  = "dbSizeInUse"
	dbQuota      = "dbQuota"
	dbQuotaUsage = "dbQuotaUsage"
	dbSizeFree   = "dbSizeFree"
)

func defaultVariables() map[string]interface{} {
	variables := map[string]interface{}{
		dbQuota:     float64(2 * 1024 * 1024 * 1024), // 2GiB
		dbSize:      float64(100 * 1024 * 1024),      // 100MiB
		dbSizeInUse: float64(60 * 1024 * 1024),       // 60MiB
	}
	variables[dbQuotaUsage] = variables[dbSize].(float64) / variables[dbQuota].(float64)
	variables[dbSizeFree] = variables[dbSize].(float64) - variables[dbSizeInUse].(float64)
	return variables
}

func evaluate(gcfg globalConfig, es epStatus) (bool, error) {
	if len(gcfg.defragRule) == 0 {
		return true, nil
	}

	variables := map[string]interface{}{
		dbQuota:      float64(gcfg.dbQuotaBytes),
		dbSize:       float64(es.Resp.DbSize),
		dbSizeInUse:  float64(es.Resp.DbSizeInUse),
		dbQuotaUsage: float64(es.Resp.DbSize) / float64(gcfg.dbQuotaBytes),
		dbSizeFree:   float64(es.Resp.DbSize - es.Resp.DbSizeInUse),
	}
	eval := goval.NewEvaluator()

	result, err := eval.Evaluate(gcfg.defragRule, variables, nil)
	if err != nil {
		return false, err
	}
	if _, ok := result.(bool); !ok {
		return false, errors.New("the rule isn't a boolean expression")
	}

	return result.(bool), err
}

func validateRule(rule string) error {
	if len(rule) == 0 {
		return nil
	}
	eval := goval.NewEvaluator()
	result, err := eval.Evaluate(rule, defaultVariables(), nil)
	if err != nil {
		return err
	}
	if _, ok := result.(bool); !ok {
		return errors.New("the rule isn't a boolean expression")
	}
	return err
}
