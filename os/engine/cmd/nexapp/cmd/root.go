// Package cmd defines all instaroute CLI commands using Cobra.
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// apiURL is the base URL of the NexappOS engine. Override with
// --api flag or NEXAPP_API environment variable.
var apiURL string

var rootCmd = &cobra.Command{
	Use:   "nexapp",
	Short: "NexappOS — command-line client for the firewall engine",
	Long: `NexappOS CLI — the command-line interface to the NexappOS firewall OS.

This tool talks to the NexappOS engine over REST and lets you manage
firewall policies, interfaces, and runtime configuration without editing
config files by hand.

Examples:
  instaroute status
  instaroute policy list
  instaroute policy add --name allow-web --src enp2s0 --dst enp1s0 --port 443 --proto tcp --action accept
  instaroute apply --commit`,
	Version: "0.1.0",
}

// Execute runs the root command. Called from main().
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func init() {
	// Default to env var, fall back to localhost
	defaultAPI := os.Getenv("NEXAPP_API")
	if defaultAPI == "" {
		defaultAPI = "http://localhost:8080"
	}
	rootCmd.PersistentFlags().StringVar(&apiURL, "api", defaultAPI,
		"NexappOS engine URL (or set NEXAPP_API env var)")
}
