package main

import (
	"errors"

	"github.com/maja42/goval"
)

const (
	dbSize      = "dbSize"
	dbSizeInUse = "dbSizeInUse"
	dbQuota     = "dbQuota"
)

func defaultVariables() map[string]interface{} {
	return map[string]interface{}{
		dbQuota:     2 * 1024 * 1024 * 1024, // 2GiB
		dbSize:      100 * 1024 * 1024,      // 100MiB
		dbSizeInUse: 60 * 1024 * 1024,       // 60MiB
	}
}

func evaluate(gcfg globalConfig, es epStatus) (bool, error) {
	if len(gcfg.defragRule) == 0 {
		return true, nil
	}

	variables := map[string]interface{}{
		dbQuota:     gcfg.dbQuotaBytes,
		dbSize:      int(es.Resp.DbSize),
		dbSizeInUse: int(es.Resp.DbSizeInUse),
	}
	eval := goval.NewEvaluator()

	result, err := eval.Evaluate(gcfg.defragRule, variables, nil)
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
