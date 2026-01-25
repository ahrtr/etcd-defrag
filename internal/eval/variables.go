package eval

// Variables holds the variables available for rule evaluation
type Variables struct {
	DbSize       int64
	DbSizeInUse  int64
	DbSizeFree   int64
	DbQuota      int64
	DbQuotaUsage float64
}

// ToMap converts variables to a map for expression evaluation
func (v *Variables) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"dbSize":       float64(v.DbSize),
		"dbSizeInUse":  float64(v.DbSizeInUse),
		"dbSizeFree":   float64(v.DbSizeFree),
		"dbQuota":      float64(v.DbQuota),
		"dbQuotaUsage": v.DbQuotaUsage,
	}
}
