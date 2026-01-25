package eval

import (
	"errors"
	"fmt"

	"github.com/maja42/goval"
)

// Evaluator evaluates defrag rules
type Evaluator struct {
	rule       string
	expression *goval.Evaluator
}

// New creates a new Evaluator with the given rule
func New(rule string) (*Evaluator, error) {
	e := &Evaluator{rule: rule}

	if rule == "" {
		return e, nil // No rule = always defrag
	}

	expr := goval.NewEvaluator()

	// Validate the rule
	defaultVars := map[string]interface{}{
		"dbQuota":      float64(2 * 1024 * 1024 * 1024), // 2GiB
		"dbSize":       float64(100 * 1024 * 1024),      // 100MiB
		"dbSizeInUse":  float64(60 * 1024 * 1024),       // 60MiB
		"dbQuotaUsage": 0.05,
		"dbSizeFree":   float64(40 * 1024 * 1024),
	}

	result, err := expr.Evaluate(rule, defaultVars, nil)
	if err != nil {
		return nil, fmt.Errorf("invalid rule %q: %w", rule, err)
	}
	if _, ok := result.(bool); !ok {
		return nil, errors.New("the rule isn't a boolean expression")
	}

	e.expression = expr
	return e, nil
}

// Evaluate evaluates the rule with the given variables
func (e *Evaluator) Evaluate(vars *Variables) (bool, error) {
	if e.expression == nil {
		return true, nil // No rule = always defrag
	}

	result, err := e.expression.Evaluate(e.rule, vars.ToMap(), nil)
	if err != nil {
		return false, fmt.Errorf("evaluation failed: %w", err)
	}

	boolResult, ok := result.(bool)
	if !ok {
		return false, fmt.Errorf("rule must evaluate to boolean, got %T", result)
	}

	return boolResult, nil
}
