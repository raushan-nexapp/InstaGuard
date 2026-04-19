package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

type policy struct {
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	SrcInterface string `json:"src_interface"`
	DstInterface string `json:"dst_interface"`
	SrcAddress   string `json:"src_address"`
	DstAddress   string `json:"dst_address"`
	Protocol     string `json:"protocol"`
	DstPort      string `json:"dst_port"`
	Action       string `json:"action"`
	Enabled      bool   `json:"enabled"`
	Priority     int    `json:"priority"`
	Description  string `json:"description"`
}

var policyCmd = &cobra.Command{
	Use:   "policy",
	Short: "Manage firewall policies",
}

var policyListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all firewall policies",
	RunE: func(cmd *cobra.Command, args []string) error {
		var resp struct {
			Count    int      `json:"count"`
			Policies []policy `json:"policies"`
		}
		if err := apiGet("/api/v1/policies", &resp); err != nil {
			return err
		}

		if resp.Count == 0 {
			fmt.Println("\n  No policies configured.\n")
			return nil
		}

		bold := color.New(color.Bold).SprintFunc()
		dim := color.New(color.Faint).SprintFunc()
		cyan := color.New(color.FgCyan, color.Bold).SprintFunc()

		line := strings.Repeat("─", 78)

		fmt.Println()
		fmt.Println(cyan("  Firewall Policies"))
		fmt.Println(dim("  " + line))
		fmt.Println()

		tw := tabwriter.NewWriter(os.Stdout, 4, 4, 3, ' ', 0)
		fmt.Fprintln(tw, "  "+bold("ID\tNAME\tSRC\tDST\tPROTO\tPORT\tACTION\tPRIO"))
		fmt.Fprintln(tw, "  "+dim("──\t────────────────────\t────────\t────────\t──────\t──────\t──────\t────"))

		for _, p := range resp.Policies {
			var action string
			switch p.Action {
			case "accept":
				action = color.GreenString(p.Action)
			case "drop", "reject":
				action = color.RedString(p.Action)
			default:
				action = p.Action
			}
			fmt.Fprintf(tw, "  %d\t%s\t%s\t%s\t%s\t%s\t%s\t%d\n",
				p.ID, p.Name, p.SrcInterface, p.DstInterface,
				p.Protocol, p.DstPort, action, p.Priority)
		}
		tw.Flush()

		fmt.Println()
		fmt.Printf("  %s %d total\n", dim("→"), resp.Count)
		fmt.Println(dim("  " + line))
		fmt.Println()
		return nil
	},
}

var policyShowCmd = &cobra.Command{
	Use:   "show [id]",
	Short: "Show one policy in detail",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var p policy
		if err := apiGet("/api/v1/policies/"+args[0], &p); err != nil {
			return err
		}

		bold := color.New(color.Bold).SprintFunc()
		dim := color.New(color.Faint).SprintFunc()
		cyan := color.New(color.FgCyan, color.Bold).SprintFunc()
		line := strings.Repeat("─", 50)

		fmt.Println()
		fmt.Println(cyan("  Policy #" + strconv.FormatInt(p.ID, 10)))
		fmt.Println(dim("  " + line))
		fmt.Printf("    %-13s %s\n", bold("Name"), p.Name)
		if p.Description != "" {
			fmt.Printf("    %-13s %s\n", bold("Description"), p.Description)
		}
		fmt.Printf("    %-13s %s → %s\n", bold("Interface"), p.SrcInterface, p.DstInterface)
		fmt.Printf("    %-13s %s → %s\n", bold("Address"), p.SrcAddress, p.DstAddress)
		fmt.Printf("    %-13s %s dport %s\n", bold("Match"), p.Protocol, p.DstPort)
		action := p.Action
		switch p.Action {
		case "accept":
			action = color.GreenString(p.Action)
		case "drop", "reject":
			action = color.RedString(p.Action)
		}
		fmt.Printf("    %-13s %s\n", bold("Action"), action)
		fmt.Printf("    %-13s %t\n", bold("Enabled"), p.Enabled)
		fmt.Printf("    %-13s %d\n", bold("Priority"), p.Priority)
		fmt.Println(dim("  " + line))
		fmt.Println()
		return nil
	},
}

var (
	addName, addSrc, addDst, addSrcAddr, addDstAddr, addProto, addPort, addAction, addDesc string
	addPriority                                                                            int
)

var policyAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Create a new firewall policy",
	RunE: func(cmd *cobra.Command, args []string) error {
		body := map[string]any{
			"name":          addName,
			"src_interface": addSrc,
			"dst_interface": addDst,
			"src_address":   addSrcAddr,
			"dst_address":   addDstAddr,
			"protocol":      addProto,
			"dst_port":      addPort,
			"action":        addAction,
			"priority":      addPriority,
			"description":   addDesc,
		}
		var created policy
		if err := apiSend("POST", "/api/v1/policies", body, &created); err != nil {
			return err
		}
		color.Green("\n  ✓ created policy #%d: %s", created.ID, created.Name)
		color.New(color.Faint).Println("    run 'instaroute apply --commit' to activate")
		fmt.Println()
		return nil
	},
}

var policyDeleteCmd = &cobra.Command{
	Use:   "delete [id]",
	Short: "Delete a policy by ID",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := apiSend("DELETE", "/api/v1/policies/"+args[0], nil, nil); err != nil {
			return err
		}
		color.Red("\n  ✗ deleted policy #%s", args[0])
		color.New(color.Faint).Println("    run 'instaroute apply --commit' to activate")
		fmt.Println()
		return nil
	},
}

var applyCommit bool

var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Regenerate nftables from DB (use --commit to actually apply)",
	RunE: func(cmd *cobra.Command, args []string) error {
		path := "/api/v1/apply"
		if applyCommit {
			path += "?commit=true"
		}
		var resp struct {
			Status   string `json:"status"`
			DryRun   bool   `json:"dry_run"`
			Policies int    `json:"policies"`
			Rendered string `json:"rendered"`
		}
		if err := apiSend("POST", path, nil, &resp); err != nil {
			return err
		}
		if resp.DryRun {
			fmt.Println(resp.Rendered)
			color.Yellow("\n  [dry-run] %d policies rendered. add --commit to apply.\n", resp.Policies)
		} else {
			color.Green("\n  ✓ applied %d policies to kernel (status: %s)\n", resp.Policies, resp.Status)
		}
		return nil
	},
}

func init() {
	policyAddCmd.Flags().StringVar(&addName, "name", "", "policy name (required)")
	policyAddCmd.Flags().StringVar(&addSrc, "src", "", "source interface, e.g. enp2s0 (required)")
	policyAddCmd.Flags().StringVar(&addDst, "dst", "", "destination interface, e.g. enp1s0 (required)")
	policyAddCmd.Flags().StringVar(&addSrcAddr, "src-addr", "any", "source address")
	policyAddCmd.Flags().StringVar(&addDstAddr, "dst-addr", "any", "destination address")
	policyAddCmd.Flags().StringVar(&addProto, "proto", "any", "protocol: any|tcp|udp|icmp")
	policyAddCmd.Flags().StringVar(&addPort, "port", "any", "destination port")
	policyAddCmd.Flags().StringVar(&addAction, "action", "", "accept|drop|reject (required)")
	policyAddCmd.Flags().IntVar(&addPriority, "priority", 100, "priority (lower = evaluated first)")
	policyAddCmd.Flags().StringVar(&addDesc, "desc", "", "human-readable description")
	policyAddCmd.MarkFlagRequired("name")
	policyAddCmd.MarkFlagRequired("src")
	policyAddCmd.MarkFlagRequired("dst")
	policyAddCmd.MarkFlagRequired("action")

	applyCmd.Flags().BoolVar(&applyCommit, "commit", false, "actually write config and reload kernel")

	policyCmd.AddCommand(policyListCmd)
	policyCmd.AddCommand(policyShowCmd)
	policyCmd.AddCommand(policyAddCmd)
	policyCmd.AddCommand(policyDeleteCmd)
	rootCmd.AddCommand(policyCmd)
	rootCmd.AddCommand(applyCmd)
}
