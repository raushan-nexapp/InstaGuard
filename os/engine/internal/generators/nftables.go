// Package generators renders runtime config files from the NexappOS
// database. Each generator produces one file (nftables.conf, dhcpd.conf,
// dnsmasq.conf, etc.) from the current state of SQLite.
package generators

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"
	"time"

	"github.com/raushan-nexapp/nexappos/os/engine/internal/models"
)

// NftablesGenerator renders /etc/nftables.conf from DB policies.
type NftablesGenerator struct {
	TemplatePath string // path to .tmpl file
	OutputPath   string // where to write the generated file
	Version      string // engine version (shown in the header)
	DryRun       bool   // if true, don't write or reload — return rendered text only
}

// tmplData is everything the template can reference.
type tmplData struct {
	Version     string
	GeneratedAt string
	Policies    []models.Policy
}

// Render produces the nftables.conf content as a string.
func (g *NftablesGenerator) Render(policies []models.Policy) (string, error) {
	tmpl, err := template.ParseFiles(g.TemplatePath)
	if err != nil {
		return "", fmt.Errorf("parse template: %w", err)
	}

	data := tmplData{
		Version:     g.Version,
		GeneratedAt: time.Now().Format(time.RFC3339),
		Policies:    policies,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("execute template: %w", err)
	}
	return buf.String(), nil
}

// Apply writes the rendered config and reloads nftables.
// Returns the rendered text on success.
func (g *NftablesGenerator) Apply(policies []models.Policy) (string, error) {
	rendered, err := g.Render(policies)
	if err != nil {
		return "", err
	}
	if g.DryRun {
		return rendered, nil
	}

	// Validate with nftables before writing (catch syntax errors)
	if err := g.validate(rendered); err != nil {
		return rendered, fmt.Errorf("ruleset validation failed: %w", err)
	}

	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(g.OutputPath), 0755); err != nil {
		return rendered, fmt.Errorf("create output dir: %w", err)
	}

	// Atomic write: temp file + rename
	tmpPath := g.OutputPath + ".tmp"
	if err := os.WriteFile(tmpPath, []byte(rendered), 0644); err != nil {
		return rendered, fmt.Errorf("write temp file: %w", err)
	}
	if err := os.Rename(tmpPath, g.OutputPath); err != nil {
		return rendered, fmt.Errorf("rename temp file: %w", err)
	}

	// Reload live ruleset
	cmd := exec.Command("nft", "-f", g.OutputPath)
	if out, err := cmd.CombinedOutput(); err != nil {
		return rendered, fmt.Errorf("nft reload failed: %w (output: %s)", err, string(out))
	}

	return rendered, nil
}

// validate runs the ruleset through `nft -c` (check only, no apply).
func (g *NftablesGenerator) validate(rendered string) error {
	tmpFile, err := os.CreateTemp("", "nexapp-check-*.nft")
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(rendered); err != nil {
		return err
	}
	tmpFile.Close()

	cmd := exec.Command("nft", "-c", "-f", tmpFile.Name())
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("%w — %s", err, string(out))
	}
	return nil
}
