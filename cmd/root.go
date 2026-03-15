package cmd

import (
	"github.com/spf13/cobra"
)

var version = "dev"

var rootCmd = &cobra.Command{
	Use:   "trail",
	Short: "A CLI planning tool for Claude Code",
	Long:  "trail keeps persistent plan files that bridge context between Claude Code sessions.",
}

func init() {
	rootCmd.Version = version
}

func Execute() error {
	return rootCmd.Execute()
}
