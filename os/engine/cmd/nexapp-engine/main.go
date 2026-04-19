// Package main is the entry point for the NexappOS firewall engine.
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/raushan-nexapp/nexappos/os/engine/internal/api"
	"github.com/raushan-nexapp/nexappos/os/engine/internal/db"
)

const Version = "0.1.0"

var database *db.DB

type StatusResponse struct {
	Service   string    `json:"service"`
	Version   string    `json:"version"`
	Status    string    `json:"status"`
	Hostname  string    `json:"hostname"`
	Timestamp time.Time `json:"timestamp"`
}

type StatsResponse struct {
	Interfaces int `json:"interfaces"`
	Policies   int `json:"policies"`
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	hostname, _ := os.Hostname()
	response := StatusResponse{
		Service:   "nexapp-engine",
		Version:   Version,
		Status:    "healthy",
		Hostname:  hostname,
		Timestamp: time.Now(),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func statsHandler(w http.ResponseWriter, r *http.Request) {
	var stats StatsResponse
	if err := database.QueryRow("SELECT COUNT(*) FROM interfaces").Scan(&stats.Interfaces); err != nil {
		http.Error(w, "db error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if err := database.QueryRow("SELECT COUNT(*) FROM policies").Scan(&stats.Policies); err != nil {
		http.Error(w, "db error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "🛡️  NexappOS Engine v"+Version)
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "System:")
	fmt.Fprintln(w, "  GET    /api/v1/status           → engine health")
	fmt.Fprintln(w, "  GET    /api/v1/stats            → row counts")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Policies:")
	fmt.Fprintln(w, "  GET    /api/v1/policies         → list all")
	fmt.Fprintln(w, "  POST   /api/v1/policies         → create new")
	fmt.Fprintln(w, "  GET    /api/v1/policies/{id}    → get one")
	fmt.Fprintln(w, "  PUT    /api/v1/policies/{id}    → update")
	fmt.Fprintln(w, "  DELETE /api/v1/policies/{id}    → remove")
}

func main() {
	dbPath := os.Getenv("NEXAPP_DB")
	if dbPath == "" {
		dbPath = "data/nexapp.db"
	}

	var err error
	database, err = db.Open(dbPath)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer database.Close()

	// Create API server + register routes
	apiServer := api.NewServer(database)

	mux := http.NewServeMux()
	mux.HandleFunc("/", rootHandler)
	mux.HandleFunc("/api/v1/status", statusHandler)
	mux.HandleFunc("/api/v1/stats", statsHandler)
	apiServer.Routes(mux)

	addr := ":8080"
	log.Println("╔════════════════════════════════════════╗")
	log.Println("║  🛡️  NexappOS Engine v" + Version + "           ║")
	log.Println("╚════════════════════════════════════════╝")
	log.Printf("Listening on http://0.0.0.0%s", addr)

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
