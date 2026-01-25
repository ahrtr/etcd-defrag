// Package cmd provides the command-line interface for etcd-defrag
package cmd

import (
	"github.com/spf13/cobra"

	"github.com/ahrtr/etcd-defrag/internal/config"
)

var (
	globalCfg = config.NewGlobalConfig()
)

// NewRootCommand creates the root cobra command
func NewRootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "etcd-defrag",
		Short: "A simple command line tool for etcd defragmentation",
		RunE:  runDefragCommand,
	}

	// Register all flags
	config.RegisterFlags(rootCmd, globalCfg)

	return rootCmd
}

// Execute runs the root command
func Execute() error {
	return NewRootCommand().Execute()
}

// runDefragCommand is called when the command is executed
func runDefragCommand(cmd *cobra.Command, args []string) error {
	// This will be implemented to call the agent
	// For now, delegate to the original main logic
	return runDefrag(globalCfg)
}

// runDefrag executes the defragmentation (bridge to existing logic)
func runDefrag(cfg *config.GlobalConfig) error {
	// TODO: This will call internal/agent.New(cfg).Run()
	// For now, return nil as placeholder
	return nil
}
