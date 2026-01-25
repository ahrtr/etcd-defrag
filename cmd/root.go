// Package cmd provides the command-line interface for etcd-defrag
package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"runtime"

	"github.com/spf13/cobra"

	"github.com/ahrtr/etcd-defrag/internal/agent"
	"github.com/ahrtr/etcd-defrag/internal/config"
	"github.com/ahrtr/etcd-defrag/pkg/version"
)

var (
	globalCfg = config.NewGlobalConfig()
)

// NewRootCommand creates the root cobra command
func NewRootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "etcd-defrag",
		Short: "A simple command line tool for etcd defragmentation",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Handle version flag
			if globalCfg.PrintVersion {
				printVersion()
				os.Exit(0)
			}
			return nil
		},
		RunE:          runDefragCommand,
		SilenceUsage:  false,
		SilenceErrors: false,
	}

	// Register all flags
	config.RegisterFlags(rootCmd, globalCfg)

	return rootCmd
}

// Execute runs the root command
func Execute() error {
	return NewRootCommand().Execute()
}

// printVersion prints version information
func printVersion() {
	fmt.Printf("etcd-defrag Version: %s\n", version.Version)
	fmt.Printf("Git SHA: %s\n", version.GitSHA)
	fmt.Printf("Go Version: %s\n", runtime.Version())
	fmt.Printf("Go OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
}

// runDefragCommand is called when the command is executed
func runDefragCommand(cmd *cobra.Command, args []string) error {
	return runDefrag(globalCfg)
}

// runDefrag executes the defragmentation
func runDefrag(cfg *config.GlobalConfig) error {
	a, err := agent.New(cfg)
	if err != nil {
		return err
	}

	ctx := context.Background()
	if err := a.Run(ctx); err != nil {
		log.Printf("Defragmentation failed: %v", err)
		return err
	}

	return nil
}
