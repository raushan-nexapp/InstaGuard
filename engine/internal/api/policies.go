package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/raushan-nexapp/instaguard/internal/models"
)

// policiesHandler handles:
//   GET  /api/v1/policies   → list all policies
//   POST /api/v1/policies   → create a new policy
func (s *Server) policiesHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.listPolicies(w, r)
	case http.MethodPost:
		s.createPolicy(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// policyItemHandler handles:
//   GET    /api/v1/policies/{id}  → get one policy
//   PUT    /api/v1/policies/{id}  → replace a policy
//   DELETE /api/v1/policies/{id}  → remove a policy
func (s *Server) policyItemHandler(w http.ResponseWriter, r *http.Request) {
	// Extract the id from the URL: /api/v1/policies/{id}
	idStr := strings.TrimPrefix(r.URL.Path, "/api/v1/policies/")
	if idStr == "" || strings.Contains(idStr, "/") {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "id must be a number")
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.getPolicy(w, id)
	case http.MethodPut:
		s.updatePolicy(w, r, id)
	case http.MethodDelete:
		s.deletePolicy(w, id)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (s *Server) listPolicies(w http.ResponseWriter, r *http.Request) {
	policies, err := models.ListPolicies(s.DB.DB)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"count":    len(policies),
		"policies": policies,
	})
}

func (s *Server) createPolicy(w http.ResponseWriter, r *http.Request) {
	var p models.Policy
	// Default to enabled — callers can set "enabled": false to disable
	p.Enabled = true

	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}

	if err := models.CreatePolicy(s.DB.DB, &p); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, p)
}

func (s *Server) getPolicy(w http.ResponseWriter, id int64) {
	p, err := models.GetPolicy(s.DB.DB, id)
	if err != nil {
		writeError(w, http.StatusNotFound, "policy not found")
		return
	}
	writeJSON(w, http.StatusOK, p)
}

func (s *Server) updatePolicy(w http.ResponseWriter, r *http.Request, id int64) {
	existing, err := models.GetPolicy(s.DB.DB, id)
	if err != nil {
		writeError(w, http.StatusNotFound, "policy not found")
		return
	}
	// Start with existing values, then overlay JSON body
	p := *existing
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	p.ID = id // prevent id change via body

	if err := models.UpdatePolicy(s.DB.DB, &p); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, p)
}

func (s *Server) deletePolicy(w http.ResponseWriter, id int64) {
	if err := models.DeletePolicy(s.DB.DB, id); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}
