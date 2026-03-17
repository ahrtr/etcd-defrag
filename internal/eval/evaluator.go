package eval

import (
	"errors"

	"github.com/maja42/goval"
)

const (
	DBSize       = "dbSize"
	DBSizeInUse  = "dbSizeInUse"
	DBQuota      = "dbQuota"
	DBQuotaUsage = "dbQuotaUsage"
	DBSizeFree   = "dbSizeFree"
)

func defaultVariables() map[string]interface{} {
	variables := map[string]interface{}{
		DBQuota:     float64(2 * 1024 * 1024 * 1024), // 2GiB
		DBSize:      float64(100 * 1024 * 1024),      // 100MiB
		DBSizeInUse: float64(60 * 1024 * 1024),       // 60MiB
	}
	variables[DBQuotaUsage] = variables[DBSize].(float64) / variables[DBQuota].(float64)
	variables[DBSizeFree] = variables[DBSize].(float64) - variables[DBSizeInUse].(float64)
	return variables
}

func ValidateRule(rule string) error {
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

func Evaluate(rule string, dbQuota, dbSize, dbSizeInUse int64) (bool, error) {
	if len(rule) == 0 {
		return true, nil
	}

	variables := map[string]interface{}{
		DBQuota:      float64(dbQuota),
		DBSize:       float64(dbSize),
		DBSizeInUse:  float64(dbSizeInUse),
		DBQuotaUsage: float64(dbSize) / float64(dbQuota),
		DBSizeFree:   float64(dbSize - dbSizeInUse),
	}
	eval := goval.NewEvaluator()

	result, err := eval.Evaluate(rule, variables, nil)
	if err != nil {
		return false, err
	}
	if _, ok := result.(bool); !ok {
		return false, errors.New("the rule isn't a boolean expression")
	}

	return result.(bool), err
}
