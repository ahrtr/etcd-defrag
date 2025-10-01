package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	globalCfg = globalConfig{}
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

	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.SetEnvPrefix("ETCD_DEFRAG")
	viper.AutomaticEnv()
	setDefaults()

	// Manually splitting, because GetStringSlice has inconsistent behavior for splitting command line flags and environment variables
	// https://github.com/spf13/viper/issues/380
	defragCmd.Flags().StringSliceVar(&globalCfg.endpoints, "endpoints", strings.Split(viper.GetString("endpoints"), ","), "comma separated etcd endpoints")

	defragCmd.Flags().BoolVar(&globalCfg.useClusterEndpoints, "cluster", viper.GetBool("cluster"), "use all endpoints from the cluster member list")
	defragCmd.Flags().BoolVar(&globalCfg.excludeLocalhost, "exclude-localhost", viper.GetBool("exclude-localhost"), "whether to exclude localhost endpoints")
	defragCmd.Flags().BoolVar(&globalCfg.moveLeader, "move-leader", viper.GetBool("move-leader"), "whether to move the leadership before performing defragmentation on the leader")

	defragCmd.Flags().DurationVar(&globalCfg.dialTimeout, "dial-timeout", viper.GetDuration("dial-timeout"), "dial timeout for client connections")
	defragCmd.Flags().DurationVar(&globalCfg.commandTimeout, "command-timeout", viper.GetDuration("command-timeout"), "command timeout (excluding dial timeout)")
	defragCmd.Flags().DurationVar(&globalCfg.keepAliveTime, "keepalive-time", viper.GetDuration("keepalive-time"), "keepalive time for client connections")
	defragCmd.Flags().DurationVar(&globalCfg.keepAliveTimeout, "keepalive-timeout", viper.GetDuration("keepalive-timeout"), "keepalive timeout for client connections")

	defragCmd.Flags().BoolVar(&globalCfg.insecure, "insecure-transport", viper.GetBool("insecure-transport"), "disable transport security for client connections")

	defragCmd.Flags().BoolVar(&globalCfg.insecureSkepVerify, "insecure-skip-tls-verify", viper.GetBool("insecure-skip-tls-verify"), "skip server certificate verification (CAUTION: this option should be enabled only for testing purposes)")
	defragCmd.Flags().StringVar(&globalCfg.certFile, "cert", viper.GetString("cert"), "identify secure client using this TLS certificate file")
	defragCmd.Flags().StringVar(&globalCfg.keyFile, "key", viper.GetString("key"), "identify secure client using this TLS key file")
	defragCmd.Flags().StringVar(&globalCfg.caFile, "cacert", viper.GetString("cacert"), "verify certificates of TLS-enabled secure servers using this CA bundle")

	defragCmd.Flags().StringVar(&globalCfg.username, "user", viper.GetString("user"), "username[:password] for authentication (prompt if password is not supplied)")
	defragCmd.Flags().StringVar(&globalCfg.password, "password", viper.GetString("password"), "password for authentication (if this option is used, --user option shouldn't include password)")

	defragCmd.Flags().StringVarP(&globalCfg.dnsDomain, "discovery-srv", "d", viper.GetString("discovery-srv"), "domain name to query for SRV records describing cluster endpoints")
	defragCmd.Flags().StringVarP(&globalCfg.dnsService, "discovery-srv-name", "", viper.GetString("discovery-srv-name"), "service name to query when using DNS discovery")
	defragCmd.Flags().BoolVar(&globalCfg.insecureDiscovery, "insecure-discovery", viper.GetBool("insecure-discovery"), "accept insecure SRV records describing cluster endpoints")

	defragCmd.Flags().BoolVar(&globalCfg.compaction, "compaction", viper.GetBool("compaction"), "whether execute compaction before the defragmentation (defaults to true)")

	defragCmd.Flags().BoolVar(&globalCfg.continueOnError, "continue-on-error", viper.GetBool("continue-on-error"), "whether continue to defragment next endpoint if current one fails")

	defragCmd.Flags().IntVar(&globalCfg.dbQuotaBytes, "etcd-storage-quota-bytes", viper.GetInt("etcd-storage-quota-bytes"), "etcd storage quota in bytes (the value passed to etcd instance by flag --quota-backend-bytes)")
	defragCmd.Flags().StringVar(&globalCfg.defragRule, "defrag-rule", viper.GetString("defrag-rule"), "defragmentation rule (etcd-defrag will run defragmentation if the rule is empty or it is evaluated to true)")

	defragCmd.Flags().BoolVar(&globalCfg.printVersion, "version", viper.GetBool("version"), "print the version and exit")

	defragCmd.Flags().BoolVar(&globalCfg.dryRun, "dry-run", viper.GetBool("dry-run"), "evaluate whether or not endpoints require defragmentation, but don't actually perform it")

	defragCmd.Flags().BoolVar(&globalCfg.skipHealthcheckClusterEndpoints, "skip-healthcheck-cluster-endpoints", viper.GetBool("skip-healthcheck-cluster-endpoints"), "skip cluster endpoint discovery during health check and only check the endpoints provided via --endpoints")

	return defragCmd
}

func setDefaults() {
	viper.SetDefault("endpoints", "127.0.0.1:2379")
	viper.SetDefault("cluster", false)
	viper.SetDefault("exclude-localhost", false)
	viper.SetDefault("move-leader", false)
	viper.SetDefault("dial-timeout", 2*time.Second)
	viper.SetDefault("command-timeout", 30*time.Second)
	viper.SetDefault("keepalive-time", 2*time.Second)
	viper.SetDefault("keepalive-timeout", 6*time.Second)
	viper.SetDefault("insecure-transport", true)
	viper.SetDefault("insecure-skip-tls-verify", false)
	viper.SetDefault("cert", "")
	viper.SetDefault("key", "")
	viper.SetDefault("cacert", "")
	viper.SetDefault("user", "")
	viper.SetDefault("password", "")
	viper.SetDefault("discovery-srv", "")
	viper.SetDefault("discovery-srv-name", "")
	viper.SetDefault("insecure-discovery", true)
	viper.SetDefault("compaction", true)
	viper.SetDefault("continue-on-error", true)
	viper.SetDefault("etcd-storage-quota-bytes", 2*1024*1024*1024)
	viper.SetDefault("defrag-rule", "")
	viper.SetDefault("version", false)
	viper.SetDefault("dry-run", false)
	viper.SetDefault("skip-healthcheck-cluster-endpoints", false)
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
		log.Println("Using dry run mode, will not perform defragmentation")
	}

	log.Println("Validating configuration.")
	if err := validateConfig(cmd, globalCfg); err != nil {
		log.Printf("Validating configuration failed: %v\n", err)
		os.Exit(1)
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

	if globalCfg.compaction && !globalCfg.dryRun {
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
	for _, ep := range eps {
		log.Print("[Before defragmentation] ")
		status, err := getMemberStatus(globalCfg, ep)
		if err != nil {
			failures++
			log.Printf("Failed to get member (%q) status, error: %v\n", ep, err)
			if !globalCfg.continueOnError {
				break
			}
			continue
		}

		evalRet, err := evaluate(globalCfg, status)
		if !evalRet || err != nil {
			if err != nil {
				failures++
				log.Printf("Evaluation failed, endpoint: %s, error:%v\n", ep, err)
				if !globalCfg.continueOnError {
					break
				}
				continue
			}
			log.Printf("Evaluation result is false, so skipping endpoint: %s\n", ep)
			continue
		}

		if globalCfg.dryRun {
			log.Printf("[Dry run] skip defragmenting endpoint %q\n", ep)
			continue
		}

		// Check if the member is a leader and move the leader if necessary
		if globalCfg.moveLeader {
			if status.Resp.Leader == status.Resp.Header.MemberId {
				log.Println("Transferring the leadership from the current leader")
				if err = moveLeader(globalCfg, status.Resp.Leader, ep); err != nil {
					log.Printf("Failed to transfer the leadership from %x to a follower, error: %v\n", status.Resp.Leader, err)
					if !globalCfg.continueOnError {
						break
					}
					continue
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
			if !globalCfg.continueOnError {
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
			if !globalCfg.continueOnError {
				break
			}
			continue
		}
	}
	if failures != 0 {
		log.Printf("%d (total %d) endpoint(s) failed to be defragmented.\n", failures, len(eps))
		os.Exit(1)
	}
	log.Println("The defragmentation is successful.")
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
		log.Printf("Validating the defragmentation rule: %v ... ", gcfg.defragRule)

		if err := validateRule(gcfg.defragRule); err != nil {
			log.Println("invalid")
			return fmt.Errorf("invalid rule %q, error: %w", gcfg.defragRule, err)
		}
		log.Println("valid")
	} else {
		log.Println("No defragmentation rule provided")
	}

	if gcfg.skipHealthcheckClusterEndpoints && len(gcfg.endpoints) == 0 {
		return errors.New("--skip-healthcheck-cluster-endpoints requires explicit endpoints to be provided via --endpoints flag")
	}

	if gcfg.skipHealthcheckClusterEndpoints && gcfg.useClusterEndpoints {
		return errors.New("--skip-healthcheck-cluster-endpoints and --cluster flags are mutually exclusive")
	}

	if gcfg.skipHealthcheckClusterEndpoints && gcfg.dnsDomain != "" {
		return errors.New("--skip-healthcheck-cluster-endpoints and --discovery-srv flags are mutually exclusive")
	}

	return nil
}

func healthCheck(gcfg globalConfig) bool {
	if gcfg.skipHealthcheckClusterEndpoints {
		log.Printf("Health check will be performed only on the explicitly provided endpoints: %v\n", gcfg.endpoints)
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

func getMembersStatus(gcfg globalConfig) ([]epStatus, error) {
	statusList, err := membersStatus(gcfg)
	if err != nil {
		return nil, err
	}

	for _, status := range statusList {
		log.Println(status.String())
	}
	return statusList, nil
}

func getMemberStatus(gcfg globalConfig, ep string) (epStatus, error) {
	status, err := memberStatus(gcfg, ep)
	if err != nil {
		return epStatus{}, err
	}
	log.Println(status.String())
	return status, nil
}

func moveLeader(gcfg globalConfig, leaderID uint64, leaderEndpoint string) error {
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
