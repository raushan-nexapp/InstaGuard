package cmd

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show engine health and DB statistics",
	RunE: func(cmd *cobra.Command, args []string) error {
		var status struct {
			Service   string `json:"service"`
			Version   string `json:"version"`
			Status    string `json:"status"`
			Hostname  string `json:"hostname"`
			Timestamp string `json:"timestamp"`
		}
		if err := apiGet("/api/v1/status", &status); err != nil {
			return err
		}
		var stats struct {
			Interfaces int `json:"interfaces"`
			Policies   int `json:"policies"`
		}
		if err := apiGet("/api/v1/stats", &stats); err != nil {
			return err
		}

		bold := color.New(color.Bold).SprintFunc()
		cyan := color.New(color.FgCyan, color.Bold).SprintFunc()
		green := color.New(color.FgGreen, color.Bold).SprintFunc()
		dim := color.New(color.Faint).SprintFunc()

		line := strings.Repeat("─", 50)

		fmt.Println()
		fmt.Println(cyan("  NexappOS Firewall"))
		fmt.Println(dim("  " + line))
		fmt.Println()
		fmt.Println(bold("  Engine"))
		fmt.Printf("    Service     %s\n", status.Service)
		fmt.Printf("    Version     %s\n", status.Version)
		fmt.Printf("    Status      %s\n", green(status.Status))
		fmt.Printf("    Hostname    %s\n", status.Hostname)
		fmt.Printf("    API URL     %s\n", apiURL)
		fmt.Println()
		fmt.Println(bold("  Database"))
		fmt.Printf("    Interfaces  %d\n", stats.Interfaces)
		fmt.Printf("    Policies    %d\n", stats.Policies)
		fmt.Println()
		fmt.Println(dim("  " + line))
		fmt.Println()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
