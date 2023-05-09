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
		dbQuota:     2 * 1024 * 1024 * 1024, // 2GiB
		dbSize:      100 * 1024 * 1024,      // 100MiB
		dbSizeInUse: 60 * 1024 * 1024,       // 60MiB
	}
	variables[dbQuotaUsage] = float64(variables[dbSize].(int)) / float64(variables[dbQuota].(int))
	variables[dbSizeFree] = variables[dbSize].(int) - variables[dbSizeInUse].(int)
	return variables
}

func evaluate(gcfg globalConfig, es epStatus) (bool, error) {
	if len(gcfg.defragRule) == 0 {
		return true, nil
	}

	variables := map[string]interface{}{
		dbQuota:      gcfg.dbQuotaBytes,
		dbSize:       int(es.Resp.DbSize),
		dbSizeInUse:  int(es.Resp.DbSizeInUse),
		dbQuotaUsage: float64(es.Resp.DbSize) / float64(gcfg.dbQuotaBytes),
		dbSizeFree:   int(es.Resp.DbSize) - int(es.Resp.DbSizeInUse),
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
