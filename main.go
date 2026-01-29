package main

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/ahrtr/etcd-defrag/internal/config"
	"github.com/ahrtr/etcd-defrag/pkg/version"
)

var (
	globalCfg = config.GlobalConfig{}
)

func newDefragCommand() *cobra.Command {
	defragCmd := &cobra.Command{
		Use:   "etcd-defrag",
		Short: "A simple command line tool for etcd defragmentation",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return viper.BindPFlags(cmd.Flags())
		},
		Run: defragCommandFunc,
	}

	config.SetupViper()
	config.RegisterFlags(defragCmd, &globalCfg)

	return defragCmd
}

func main() {
	defragCmd := newDefragCommand()
	if err := defragCmd.Execute(); err != nil {
		if defragCmd.SilenceErrors {
			log.Println("Error:", err)
			os.Exit(1)
		} else {
			os.Exit(1)
		}
	}
}

func printVersion(printVersion bool) {
	if printVersion {
		fmt.Printf("etcd-defrag Version: %s\n", version.Version)
		fmt.Printf("Git SHA: %s\n", version.GitSHA)
		fmt.Printf("Go Version: %s\n", runtime.Version())
		fmt.Printf("Go OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
		os.Exit(0)
	}
}

func defragCommandFunc(cmd *cobra.Command, args []string) {
	printVersion(globalCfg.PrintVersion)

	if globalCfg.DryRun {
		log.Println("Using dry run mode, will not perform defragmentation")
	}

	log.Println("Validating configuration.")
	if err := globalCfg.Validate(cmd); err != nil {
		log.Printf("Validating configuration failed: %v\n", err)
		os.Exit(1)
	}

	if len(globalCfg.DefragRule) > 0 {
		log.Printf("Validating the defragmentation rule: %v ... ", globalCfg.DefragRule)

		if err := validateRule(globalCfg.DefragRule); err != nil {
			log.Println("invalid")
			log.Printf("Validating configuration failed: invalid rule %q, error: %v\n", globalCfg.DefragRule, err)
			os.Exit(1)
		}
		log.Println("valid")
	} else {
		log.Println("No defragmentation rule provided")
	}

	log.Println("Performing health check.")
	if !healthCheck(globalCfg) {
		os.Exit(1)
	}

	log.Println("Getting members status")
	statusList, err := getMembersStatus(globalCfg)
	if err != nil {
		log.Printf("Failed to get members status: %v\n", err)
		os.Exit(1)
	}

	eps, err := endpointsWithLeaderAtEnd(globalCfg, statusList)
	if err != nil {
		log.Printf("Failed to get endpoints: %v\n", err)
		os.Exit(1)
	}

	if globalCfg.Compaction && !globalCfg.DryRun {
		log.Printf("Running compaction until revision: %d ... ", statusList[0].Resp.Header.Revision)
		if err := compact(globalCfg, statusList[0].Resp.Header.Revision, eps[0]); err != nil {
			log.Printf("failed, %v\n", err)
		} else {
			log.Println("successful")
		}
	} else {
		log.Println("Skip compaction.")
	}

	log.Printf("%d endpoint(s) need to be defragmented: %v\n", len(eps), eps)
	failures := 0
	for index, ep := range eps {
		log.Print("[Before defragmentation] ")
		status, err := getMemberStatus(globalCfg, ep)
		if err != nil {
			failures++
			log.Printf("Failed to get member (%q) status, error: %v\n", ep, err)
			if !globalCfg.ContinueOnError {
				break
			}
			continue
		}

		evalRet, err := evaluate(globalCfg, status)
		if !evalRet || err != nil {
			if err != nil {
				failures++
				log.Printf("Evaluation failed, endpoint: %s, error:%v\n", ep, err)
				if !globalCfg.ContinueOnError {
					break
				}
				continue
			}
			log.Printf("Evaluation result is false, so skipping endpoint: %s\n", ep)
			continue
		}

		if globalCfg.DryRun {
			log.Printf("[Dry run] skip defragmenting endpoint %q\n", ep)
			continue
		}

		// Check if the member is a leader and move the leader if necessary
		if globalCfg.MoveLeader {
			if status.Resp.Leader == status.Resp.Header.MemberId {
				log.Println("Transferring the leadership from the current leader")
				if err = moveLeader(globalCfg, status.Resp.Leader, ep); err != nil {
					log.Printf("Failed to transfer the leadership from %x to a follower, error: %v\n", status.Resp.Leader, err)
					if !globalCfg.ContinueOnError {
						break
					}
					continue
				}
				if globalCfg.WaitBetweenDefrags > 0 {
					log.Printf("Transferred the leadership successfully! Waiting for %s for next operation\n", globalCfg.WaitBetweenDefrags.String())
					time.Sleep(globalCfg.WaitBetweenDefrags)
				}
			}
		}

		log.Printf("Defragmenting endpoint %q\n", ep)
		startTS := time.Now()
		err = defragment(globalCfg, ep)
		d := time.Since(startTS)
		if err != nil {
			failures++
			log.Printf("Failed to defragment etcd member %q. took %s. (%v)\n", ep, d.String(), err)
			if !globalCfg.ContinueOnError {
				break
			}
			continue
		} else {
			log.Printf("Finished defragmenting etcd endpoint %q. took %s\n", ep, d.String())
		}

		log.Print("[Post defragmentation] ")
		_, err = getMemberStatus(globalCfg, ep)
		if err != nil {
			failures++
			log.Printf("Failed to get member (%q) status, error: %v\n", ep, err)
			if !globalCfg.ContinueOnError {
				break
			}
			continue
		}

		if globalCfg.WaitBetweenDefrags > 0 && index < len(eps)-1 {
			log.Printf("Waiting for %s for next operation\n", globalCfg.WaitBetweenDefrags.String())
			time.Sleep(globalCfg.WaitBetweenDefrags)
		}
	}
	if failures != 0 {
		log.Printf("%d (total %d) endpoint(s) failed to be defragmented.\n", failures, len(eps))
		os.Exit(1)
	}
	log.Println("The defragmentation is successful.")

	// Perform auto-disalarm if enabled and not in dry-run mode
	if !globalCfg.DryRun && globalCfg.AutoDisalarm {
		log.Println("Start auto-disalarm")
		// Get updated status after defragmentation
		updatedStatusList, err := getMembersStatus(globalCfg)
		if err != nil {
			log.Printf("Failed to get updated members status for auto-disalarm: %v\n", err)
		} else {
			if err := performAutoDisalarm(globalCfg, updatedStatusList); err != nil {
				log.Printf("Auto-disalarm failed: %v\n", err)
			}
		}
	}
}

func healthCheck(gcfg config.GlobalConfig) bool {
	if gcfg.SkipHealthcheckClusterEndpoints {
		log.Printf("Health check will be performed only on the explicitly provided endpoints: %v\n", gcfg.Endpoints)
	} else {
		log.Println("Health check will be performed on all cluster member endpoints")
	}

	healthInfos, err := clusterHealth(gcfg)
	if err != nil {
		log.Printf("Failed to get members' health info: %v\n", err)
		return false
	}

	unhealthyCount := 0
	for _, healthInfo := range healthInfos {
		if !healthInfo.Health {
			unhealthyCount++
		}

		log.Println(healthInfo.String())
	}

	return unhealthyCount == 0
}

func getMembersStatus(gcfg config.GlobalConfig) ([]epStatus, error) {
	statusList, err := membersStatus(gcfg)
	if err != nil {
		return nil, err
	}

	for _, status := range statusList {
		log.Println(status.String())
	}
	return statusList, nil
}

func getMemberStatus(gcfg config.GlobalConfig, ep string) (epStatus, error) {
	status, err := memberStatus(gcfg, ep)
	if err != nil {
		return epStatus{}, err
	}
	log.Println(status.String())
	return status, nil
}

func moveLeader(gcfg config.GlobalConfig, leaderID uint64, leaderEndpoint string) error {
	memberlistResp, err := memberList(gcfg)
	if err != nil {
		return fmt.Errorf("failed to get member list: %w", err)
	}

	if len(memberlistResp.Members) == 1 {
		log.Println("Skip moving leader as there is only one member in the cluster")
		return nil
	}

	// pick up a follower to transfer the leadership to
	newLeaderID := uint64(0)
	for _, m := range memberlistResp.Members {
		if m.ID != leaderID {
			newLeaderID = m.ID
			break
		}
	}

	if newLeaderID == 0 {
		return fmt.Errorf("coundn't find a follower in the %d member cluster", len(memberlistResp.Members))
	}

	log.Printf("Transferring the leadership from %x to %x\n", leaderID, newLeaderID)
	return transferLeadership(globalCfg, leaderEndpoint, newLeaderID)
}
