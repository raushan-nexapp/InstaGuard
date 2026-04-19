package api

import (
	"net/http"

	"github.com/raushan-nexapp/nexappos/os/engine/internal/generators"
	"github.com/raushan-nexapp/nexappos/os/engine/internal/models"
)

// applyHandler regenerates all runtime config files from the database
// and reloads services.
//
// POST /api/v1/apply
func (s *Server) applyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// Pull current policies
	policies, err := models.ListPolicies(s.DB.DB)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "load policies: "+err.Error())
		return
	}

	// Dry-run by default unless ?commit=true
	dryRun := r.URL.Query().Get("commit") != "true"

	gen := &generators.NftablesGenerator{
		TemplatePath: "templates/nftables.conf.tmpl",
		OutputPath:   "/etc/nftables.conf",
		Version:      "0.1.0",
		DryRun:       dryRun,
	}

	rendered, err := gen.Apply(policies)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{
			"error":    err.Error(),
			"rendered": rendered,
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"status":   "ok",
		"dry_run":  dryRun,
		"policies": len(policies),
		"rendered": rendered,
	})
}
