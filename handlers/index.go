package handlers

import (
	"encoding/json"
	"ldap-search/config"
	"net/http"
)

// HomeHandler serves app metadata, excluding sensitive LDAP data
// @Summary Get application metadata
// @Description Returns the application configuration, including endpoint details, excluding sensitive LDAP connection data
// @Tags metadata
// @Accept json
// @Produce json
// @Success 200 {object} config.Config
// @Router / [get]
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	cfg := config.GetConfig()

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Encode response as JSON
	json.NewEncoder(w).Encode(cfg)
}

// HealthHandler provides a health check endpoint for monitoring
// @Summary Health check
// @Description Returns a simple status to indicate the application is running
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string "Example response: {\"status\": \"ok\"}"
// @Router /health [get]
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Encode response as JSON
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
