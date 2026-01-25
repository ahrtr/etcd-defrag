package main

import (
	"fmt"
	"log"
)

// performAutoDisalarm performs automatic disalarm operation
func performAutoDisalarm(gcfg globalConfig, statusList []epStatus) error {
	alarms, err := noSpaceAlarms(gcfg)
	if err != nil {
		return fmt.Errorf("failed to get NOSPACE alarms: %w", err)
	}

	if len(alarms) == 0 {
		log.Println("No NOSPACE alarms found, skipping auto-disalarm")
		return nil
	}

	log.Println("Found NOSPACE alarms")

	// Check if all members' DB size is below threshold
	epsWithDBSize := checkAllMembersDBSize(gcfg, statusList)
	if len(epsWithDBSize) > 0 {
		log.Printf("Members %v DB size is still above threshold (%.2f), skipping auto-disalarm\n", epsWithDBSize, gcfg.disalarmThreshold)
		return nil
	}

	log.Println("Performing auto-disalarm operation...")
	if err := disAlarmNoSpaceAlarms(gcfg, alarms); err != nil {
		return fmt.Errorf("failed to disalarm NOSPACE alarms: %w", err)
	}

	log.Println("Auto-disalarm operation completed successfully")
	return nil
}

// checkAllMembersDBSize checks if all members' DB size is below the threshold
func checkAllMembersDBSize(gcfg globalConfig, statusList []epStatus) []string {
	var eps []string
	threshold := float64(gcfg.dbQuotaBytes) * gcfg.disalarmThreshold
	for _, status := range statusList {
		if float64(status.Resp.DbSize) > threshold {
			eps = append(eps, status.Ep)
		}
	}
	return eps
}
