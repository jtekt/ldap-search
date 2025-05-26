package handlers

import (
	"encoding/json"
	"ldap-search/config"
	"net/http"
)

// HomeHandler serves app metadata, excluding sensitive LDAP data
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	cfg := config.GetConfig()

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Encode response as JSON
	json.NewEncoder(w).Encode(cfg)
}

// HealthHandler provides a health check endpoint for monitoring
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Encode response as JSON
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
