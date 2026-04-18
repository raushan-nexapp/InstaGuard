// Package api defines HTTP handlers for the InstaGuard REST API.
package api

import (
	"encoding/json"
	"net/http"

	"github.com/raushan-nexapp/instaguard/internal/db"
)

// Server bundles shared dependencies for all HTTP handlers.
type Server struct {
	DB *db.DB
}

// NewServer creates a new API server bound to the given database.
func NewServer(database *db.DB) *Server {
	return &Server{DB: database}
}

// ---- helpers ----

// writeJSON writes any value as JSON with the given status code.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

// writeError writes a standard error envelope as JSON.
func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// Routes registers every API handler on the given ServeMux.
func (s *Server) Routes(mux *http.ServeMux) {
	// Policy CRUD — collection
	mux.HandleFunc("/api/v1/policies", s.policiesHandler)
	// Policy CRUD — item (uses /api/v1/policies/{id})
	mux.HandleFunc("/api/v1/policies/", s.policyItemHandler)
	mux.HandleFunc("/api/v1/apply", s.applyHandler)
}
