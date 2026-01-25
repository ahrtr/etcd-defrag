package agent

import (
	"context"
	"log"

	"go.etcd.io/etcd/api/v3/etcdserverpb"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// autoDisalarm clears NOSPACE alarms if threshold conditions are met
func (a *Agent) autoDisalarm(ctx context.Context, endpoints []string, statuses map[string]*MemberStatus) error {
	// Check if all members are below threshold
	threshold := int64(float64(a.cfg.EtcdStorageQuotaBytes) * a.cfg.DisalarmThreshold)

	for ep, status := range statuses {
		if status.DbSize >= threshold {
			log.Printf("Endpoint %s db size %d >= threshold %d, skipping disalarm",
				ep, status.DbSize, threshold)
			return nil
		}
	}

	// Get client
	cli, err := a.getClient(endpoints[0])
	if err != nil {
		return err
	}
	defer cli.Close()

	ctx, cancel := context.WithTimeout(ctx, a.cfg.CommandTimeout)
	defer cancel()

	// Check for alarms
	alarmResp, err := cli.AlarmList(ctx)
	if err != nil {
		return err
	}

	// Disarm NOSPACE alarms
	hasNospace := false
	for _, alarm := range alarmResp.Alarms {
		if alarm.Alarm == etcdserverpb.AlarmType_NOSPACE {
			hasNospace = true
			_, err := cli.AlarmDisarm(ctx, &clientv3.AlarmMember{
				MemberID: alarm.MemberID,
				Alarm:    alarm.Alarm,
			})
			if err != nil {
				return err
			}
			log.Printf("Disarmed NOSPACE alarm for member %x", alarm.MemberID)
		}
	}

	if hasNospace {
		log.Println("Successfully disarmed NOSPACE alarms")
	}

	return nil
}
