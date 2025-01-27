package main

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/spf13/cobra"
)

var (
	globalCfg = globalConfig{}
)

func newDefragCommand() *cobra.Command {
	defragCmd := &cobra.Command{
		Use:   "etcd-defrag",
		Short: "A simple command line tool for etcd defragmentation",
		Run:   defragCommandFunc,
	}

	defragCmd.Flags().StringSliceVar(&globalCfg.endpoints, "endpoints", []string{"127.0.0.1:2379"}, "comma separated etcd endpoints")
	defragCmd.Flags().BoolVar(&globalCfg.useClusterEndpoints, "cluster", false, "use all endpoints from the cluster member list")
	defragCmd.Flags().BoolVar(&globalCfg.excludeLocalhost, "exclude-localhost", false, "whether to exclude localhost endpoints")
	defragCmd.Flags().BoolVar(&globalCfg.moveLeader, "move-leader", false, "whether to move the leader to a randomly picked non-leader ID and make it the new leader")

	defragCmd.Flags().DurationVar(&globalCfg.dialTimeout, "dial-timeout", 2*time.Second, "dial timeout for client connections")
	defragCmd.Flags().DurationVar(&globalCfg.commandTimeout, "command-timeout", 30*time.Second, "command timeout (excluding dial timeout)")
	defragCmd.Flags().DurationVar(&globalCfg.keepAliveTime, "keepalive-time", 2*time.Second, "keepalive time for client connections")
	defragCmd.Flags().DurationVar(&globalCfg.keepAliveTimeout, "keepalive-timeout", 6*time.Second, "keepalive timeout for client connections")

	defragCmd.Flags().BoolVar(&globalCfg.insecure, "insecure-transport", true, "disable transport security for client connections")

	defragCmd.Flags().BoolVar(&globalCfg.insecureSkepVerify, "insecure-skip-tls-verify", false, "skip server certificate verification (CAUTION: this option should be enabled only for testing purposes)")
	defragCmd.Flags().StringVar(&globalCfg.certFile, "cert", "", "identify secure client using this TLS certificate file")
	defragCmd.Flags().StringVar(&globalCfg.keyFile, "key", "", "identify secure client using this TLS key file")
	defragCmd.Flags().StringVar(&globalCfg.caFile, "cacert", "", "verify certificates of TLS-enabled secure servers using this CA bundle")

	defragCmd.Flags().StringVar(&globalCfg.username, "user", "", "username[:password] for authentication (prompt if password is not supplied)")
	defragCmd.Flags().StringVar(&globalCfg.password, "password", "", "password for authentication (if this option is used, --user option shouldn't include password)")

	defragCmd.Flags().StringVarP(&globalCfg.dnsDomain, "discovery-srv", "d", "", "domain name to query for SRV records describing cluster endpoints")
	defragCmd.Flags().StringVarP(&globalCfg.dnsService, "discovery-srv-name", "", "", "service name to query when using DNS discovery")
	defragCmd.Flags().BoolVar(&globalCfg.insecureDiscovery, "insecure-discovery", true, "accept insecure SRV records describing cluster endpoints")

	defragCmd.Flags().BoolVar(&globalCfg.compaction, "compaction", true, "whether execute compaction before the defragmentation (defaults to true)")

	defragCmd.Flags().BoolVar(&globalCfg.continueOnError, "continue-on-error", true, "whether continue to defragment next endpoint if current one fails")

	defragCmd.Flags().IntVar(&globalCfg.dbQuotaBytes, "etcd-storage-quota-bytes", 2*1024*1024*1024, "etcd storage quota in bytes (the value passed to etcd instance by flag --quota-backend-bytes)")
	defragCmd.Flags().StringVar(&globalCfg.defragRule, "defrag-rule", "", "defragmentation rule (etcd-defrag will run defragmentation if the rule is empty or it is evaluated to true)")

	defragCmd.Flags().BoolVar(&globalCfg.printVersion, "version", false, "print the version and exit")

	defragCmd.Flags().BoolVar(&globalCfg.dryRun, "dry-run", false, "evaluate whether or not endpoints require defragmentation, but don't actually perform it")
	return defragCmd
}

func main() {
	defragCmd := newDefragCommand()
	if err := defragCmd.Execute(); err != nil {
		if defragCmd.SilenceErrors {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		} else {
			os.Exit(1)
		}
	}
}

func printVersion(printVersion bool) {
	if printVersion {
		fmt.Printf("etcd-defrag Version: %s\n", Version)
		fmt.Printf("Git SHA: %s\n", GitSHA)
		fmt.Printf("Go Version: %s\n", runtime.Version())
		fmt.Printf("Go OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
		os.Exit(0)
	}
}

func defragCommandFunc(cmd *cobra.Command, args []string) {
	printVersion(globalCfg.printVersion)

	if globalCfg.dryRun {
		fmt.Println("Using dry run mode, will not perform defragmentation")
	}

	fmt.Println("Validating configuration.")
	if err := validateConfig(cmd, globalCfg); err != nil {
		fmt.Fprintf(os.Stderr, "Validating configuration failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Performing health check.")
	if !healthCheck(globalCfg) {
		os.Exit(1)
	}

	fmt.Println("Getting members status")
	statusList, err := getMembersStatus(globalCfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get members status: %v\n", err)
		os.Exit(1)
	}

	eps, err := endpointsWithLeaderAtEnd(globalCfg, statusList)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get endpoints: %v\n", err)
		os.Exit(1)
	}

	if globalCfg.compaction && !globalCfg.dryRun {
		fmt.Printf("Running compaction until revision: %d ... ", statusList[0].Resp.Header.Revision)
		if err := compact(globalCfg, statusList[0].Resp.Header.Revision, eps[0]); err != nil {
			fmt.Printf("failed, %v\n", err)
		} else {
			fmt.Println("successful")
		}
	} else {
		fmt.Println("Skip compaction.")
	}

	fmt.Printf("%d endpoint(s) need to be defragmented: %v\n", len(eps), eps)
	failures := 0
	for _, ep := range eps {
		fmt.Print("[Before defragmentation] ")
		status, err := getMemberStatus(globalCfg, ep)
		if err != nil {
			failures++
			fmt.Fprintf(os.Stderr, "Failed to get member (%q) status, error: %v\n", ep, err)
			if !globalCfg.continueOnError {
				break
			}
			continue
		}

		evalRet, err := evaluate(globalCfg, status)
		if !evalRet || err != nil {
			if err != nil {
				failures++
				fmt.Fprintf(os.Stderr, "Evaluation failed, endpoint: %s, error:%v\n", ep, err)
				if !globalCfg.continueOnError {
					break
				}
				continue
			}
			fmt.Printf("Evaluation result is false, so skipping endpoint: %s\n", ep)
			continue
		}

		if globalCfg.dryRun {
			fmt.Printf("[Dry run] skip defragmenting endpoint %q\n", ep)
			continue
		}

		// Check if the member is a leader and move the leader if necessary
		if globalCfg.moveLeader {
			if status.Resp.Leader == status.Resp.Header.MemberId {
				fmt.Printf("Member %q is the leader. Attempting to move the leader...\n", ep)

				// Identify a non-leader member to transfer the leadership
				newLeaderID := uint64(0)
				for _, memberStatus := range statusList {
					if memberStatus.Resp.Header.MemberId != status.Resp.Header.MemberId {
						newLeaderID = memberStatus.Resp.Header.MemberId
						break
					}
				}

				if newLeaderID == 0 {
					failures++
					fmt.Fprintf(os.Stderr, "Failed to find a non-leader member to transfer leadership from %q.\n", ep)
					if !globalCfg.continueOnError {
						break
					}
					continue
				}

				// Perform the leader transfer
				err = transferLeadership(globalCfg, status.Ep, newLeaderID)
				if err != nil {
					failures++
					fmt.Fprintf(os.Stderr, "Failed to move leader from %s to member ID %d: %v\n", status.Ep, newLeaderID, err)
					if !globalCfg.continueOnError {
						break
					}
					continue
				}
			}
		}

		fmt.Printf("Defragmenting endpoint %q\n", ep)
		startTs := time.Now()
		err = defragment(globalCfg, ep)
		d := time.Since(startTs)
		if err != nil {
			failures++
			fmt.Fprintf(os.Stderr, "Failed to defragment etcd member %q. took %s. (%v)\n", ep, d.String(), err)
			if !globalCfg.continueOnError {
				break
			}
			continue
		} else {
			fmt.Printf("Finished defragmenting etcd endpoint %q. took %s\n", ep, d.String())
		}

		fmt.Print("[Post defragmentation] ")
		_, err = getMemberStatus(globalCfg, ep)
		if err != nil {
			failures++
			fmt.Fprintf(os.Stderr, "Failed to get member (%q) status, error: %v\n", ep, err)
			if !globalCfg.continueOnError {
				break
			}
			continue
		}
	}
	if failures != 0 {
		fmt.Fprintf(os.Stderr, "%d (total %d) endpoint(s) failed to be defragmented.\n", failures, len(eps))
		os.Exit(1)
	}
	fmt.Println("The defragmentation is successful.")
}

func validateConfig(cmd *cobra.Command, gcfg globalConfig) error {
	if gcfg.certFile == "" && cmd.Flags().Changed("cert") {
		return errors.New("empty string is passed to --cert option")
	}

	if gcfg.keyFile == "" && cmd.Flags().Changed("key") {
		return errors.New("empty string is passed to --key option")
	}

	if gcfg.caFile == "" && cmd.Flags().Changed("cacert") {
		return errors.New("empty string is passed to --cacert option")
	}

	if len(gcfg.defragRule) > 0 {
		fmt.Printf("Validating the defragmentation rule: %v ... ", gcfg.defragRule)

		if err := validateRule(gcfg.defragRule); err != nil {
			fmt.Println("invalid")
			return fmt.Errorf("invalid rule %q, error: %w", gcfg.defragRule, err)
		}
		fmt.Println("valid")
	} else {
		fmt.Println("No defragmentation rule provided")
	}

	return nil
}

func healthCheck(gcfg globalConfig) bool {
	healthInfos, err := clusterHealth(gcfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get members' health info: %v\n", err)
		return false
	}

	unhealthyCount := 0
	for _, healthInfo := range healthInfos {
		if !healthInfo.Health {
			unhealthyCount++
		}

		fmt.Println(healthInfo.String())
	}

	return unhealthyCount == 0
}

func getMembersStatus(gcfg globalConfig) ([]epStatus, error) {
	statusList, err := membersStatus(gcfg)
	if err != nil {
		return nil, err
	}

	for _, status := range statusList {
		fmt.Println(status.String())
	}
	return statusList, nil
}

func getMemberStatus(gcfg globalConfig, ep string) (epStatus, error) {
	status, err := memberStatus(gcfg, ep)
	if err != nil {
		return epStatus{}, err
	}
	fmt.Println(status.String())
	return status, nil
}
