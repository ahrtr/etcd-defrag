package main

import (
	"errors"
	"fmt"
	"os"
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

	defragCmd.Flags().DurationVar(&globalCfg.dialTimeout, "dial-timeout", 2*time.Second, "dial timeout for client connections")
	defragCmd.Flags().DurationVar(&globalCfg.commandTimeout, "command-timeout", 60*time.Second, "command timeout (excluding dial timeout)")
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

	defragCmd.Flags().BoolVar(&globalCfg.useClusterEndpoints, "cluster", false, "use all endpoints from the cluster member list")
	defragCmd.Flags().BoolVar(&globalCfg.continueOnError, "continue-on-error", true, "whether continue to defragment next endpoint if current one fails")

	defragCmd.Flags().IntVar(&globalCfg.dbQuotaBytes, "etcd-storage-quota-bytes", 2*1024*1024*1024, "etcd storage quota in bytes (the value passed to etcd instance by flag --quota-backend-bytes)")
	defragCmd.Flags().StringSliceVar(&globalCfg.defragRules, "defrag-rules", []string{}, "comma separated rules (etcd-defrag will run defragmentation if the rule is empty or any rule is evaluated to true)")

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

func defragCommandFunc(cmd *cobra.Command, args []string) {
	fmt.Println("Validating configuration.")
	if err := validateConfig(cmd); err != nil {
		fmt.Fprintf(os.Stderr, "Validating configuration failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Performing health check.")
	if !healthCheck(globalCfg) {
		os.Exit(1)
	}

	fmt.Println("Getting members' status")
	statusList, err := getMemberStatus(globalCfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get members' status: %v\n", err)
		os.Exit(1)
	}

	eps, err := endpointsWithLeaderAtEnd(globalCfg, statusList)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get endpoints: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("%d endpoints need to be defragmented: %v\n", len(eps), eps)
	cfg := clientConfigWithoutEndpoints(globalCfg)
	failures := 0
	for i, ep := range eps {
		cfg.Endpoints = []string{ep}

		fmt.Printf("Defragmenting endpoint: %s\n", ep)

		evalRet, err := evaluate(globalCfg, statusList[i])
		if err != nil {

		}
		if !evalRet || err != nil {
			if err != nil {
				failures++
				fmt.Fprintf(os.Stderr, "Evaluation failed, endpoint: %s, error:%v\n", ep, err)
				if !globalCfg.continueOnError {
					break
				}
			}
			if !evalRet {
				fmt.Fprintf(os.Stderr, "Evaluation result is false, so skipping endpoint: %s\n", ep)
			}

			continue
		}

		c, err := createClient(cfg)
		if err != nil {
			failures++
			fmt.Fprintf(os.Stderr, "Failed to connect to member[%s]: %v\n", ep, err)
			if !globalCfg.continueOnError {
				break
			}
			continue
		}

		ctx, cancel := commandCtx(globalCfg.commandTimeout)
		startTs := time.Now()
		_, err = c.Defragment(ctx, ep)
		d := time.Now().Sub(startTs)
		cancel()

		if err != nil {
			failures++
			fmt.Fprintf(os.Stderr, "Failed to defragment etcd member [%s]. took %s. (%v)\n", ep, d.String(), err)
			if !globalCfg.continueOnError {
				break
			}
		} else {
			fmt.Printf("Finished defragmenting etcd member[%s]. took %s\n", ep, d.String())
		}
	}
	if failures != 0 {
		os.Exit(1)
	}
	fmt.Println("The defragmentation is successful.")
}

func validateConfig(cmd *cobra.Command) error {
	if globalCfg.certFile == "" && cmd.Flags().Changed("cert") {
		return errors.New("empty string is passed to --cert option")
	}

	if globalCfg.keyFile == "" && cmd.Flags().Changed("key") {
		return errors.New("empty string is passed to --key option")
	}

	if globalCfg.caFile == "" && cmd.Flags().Changed("cacert") {
		return errors.New("empty string is passed to --cacert option")
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

func getMemberStatus(gcfg globalConfig) ([]epStatus, error) {
	statusList, err := memberStatus(gcfg)
	if err != nil {
		return nil, err
	}

	for _, status := range statusList {
		fmt.Println(status.String())
	}
	return statusList, nil
}
