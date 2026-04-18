// Package cmd defines all instaguard CLI commands using Cobra.
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// apiURL is the base URL of the InstaGuard engine. Override with
// --api flag or INSTAGUARD_API environment variable.
var apiURL string

var rootCmd = &cobra.Command{
	Use:   "instaguard",
	Short: "InstaGuard — command-line client for the firewall engine",
	Long: `InstaGuard CLI — the command-line interface to the InstaGuard firewall OS.

This tool talks to the InstaGuard engine over REST and lets you manage
firewall policies, interfaces, and runtime configuration without editing
config files by hand.

Examples:
  instaguard status
  instaguard policy list
  instaguard policy add --name allow-web --src enp2s0 --dst enp1s0 --port 443 --proto tcp --action accept
  instaguard apply --commit`,
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
	defaultAPI := os.Getenv("INSTAGUARD_API")
	if defaultAPI == "" {
		defaultAPI = "http://localhost:8080"
	}
	rootCmd.PersistentFlags().StringVar(&apiURL, "api", defaultAPI,
		"InstaGuard engine URL (or set INSTAGUARD_API env var)")
}
