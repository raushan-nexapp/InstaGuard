package cmd

import (
	"fmt"
	"strconv"
	"text/tabwriter"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// ---- policy struct (matches server response) ----
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

// ---- parent "policy" command ----
var policyCmd = &cobra.Command{
	Use:   "policy",
	Short: "Manage firewall policies",
}

// ---- policy list ----
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
			fmt.Println("No policies configured.")
			return nil
		}

		tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		bold := color.New(color.Bold).SprintFunc()
		fmt.Fprintln(tw, bold("ID\tNAME\tSRC\tDST\tPROTO\tPORT\tACTION\tPRIO"))
		for _, p := range resp.Policies {
			action := p.Action
			switch p.Action {
			case "accept":
				action = color.GreenString(p.Action)
			case "drop", "reject":
				action = color.RedString(p.Action)
			}
			fmt.Fprintf(tw, "%d\t%s\t%s\t%s\t%s\t%s\t%s\t%d\n",
				p.ID, p.Name, p.SrcInterface, p.DstInterface,
				p.Protocol, p.DstPort, action, p.Priority)
		}
		tw.Flush()
		fmt.Printf("\n%d policies total\n", resp.Count)
		return nil
	},
}

// ---- policy show ----
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
		fmt.Printf("%s\n", bold("Policy #"+strconv.FormatInt(p.ID, 10)))
		fmt.Printf("  Name         : %s\n", p.Name)
		fmt.Printf("  Description  : %s\n", p.Description)
		fmt.Printf("  Source IF    : %s\n", p.SrcInterface)
		fmt.Printf("  Dest IF      : %s\n", p.DstInterface)
		fmt.Printf("  Source addr  : %s\n", p.SrcAddress)
		fmt.Printf("  Dest addr    : %s\n", p.DstAddress)
		fmt.Printf("  Protocol     : %s\n", p.Protocol)
		fmt.Printf("  Dest port    : %s\n", p.DstPort)
		fmt.Printf("  Action       : %s\n", p.Action)
		fmt.Printf("  Enabled      : %t\n", p.Enabled)
		fmt.Printf("  Priority     : %d\n", p.Priority)
		return nil
	},
}

// ---- policy add ----
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
		color.Green("✓ created policy #%d: %s", created.ID, created.Name)
		fmt.Println("  (run `instaguard apply --commit` to activate)")
		return nil
	},
}

// ---- policy delete ----
var policyDeleteCmd = &cobra.Command{
	Use:   "delete [id]",
	Short: "Delete a policy by ID",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := apiSend("DELETE", "/api/v1/policies/"+args[0], nil, nil); err != nil {
			return err
		}
		color.Red("✗ deleted policy #%s", args[0])
		fmt.Println("  (run `instaguard apply --commit` to activate)")
		return nil
	},
}

// ---- apply command ----
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
			color.Yellow("\n[dry-run] %d policies rendered. Add --commit to apply.", resp.Policies)
		} else {
			color.Green("✓ applied %d policies to kernel (status: %s)", resp.Policies, resp.Status)
		}
		return nil
	},
}

func init() {
	// flags for `policy add`
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

	// flags for `apply`
	applyCmd.Flags().BoolVar(&applyCommit, "commit", false, "actually write config and reload kernel (otherwise dry-run)")

	// wire up tree
	policyCmd.AddCommand(policyListCmd)
	policyCmd.AddCommand(policyShowCmd)
	policyCmd.AddCommand(policyAddCmd)
	policyCmd.AddCommand(policyDeleteCmd)
	rootCmd.AddCommand(policyCmd)
	rootCmd.AddCommand(applyCmd)
}
