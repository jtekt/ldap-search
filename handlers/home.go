package handlers

import (
	"encoding/json"
	"ldap-search/config"
	"net/http"
)

// HomeHandler serves app metadata, excluding sensitive LDAP data
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	cfg := config.LoadConfig()

	// Create a sanitized response struct
	response := struct {
		JWKSURI       string   `json:"jwks_uri"`
		LDAPHostname  string   `json:"ldap_hostname"`
		LDAPSearchOU  string   `json:"ldap_search_ou"`
		IgnoredFields []string `json:"ignored_fields"`
	}{
		JWKSURI:       cfg.JWKSURI,
		LDAPHostname:  cfg.LDAPHostname,
		LDAPSearchOU:  cfg.LDAPSearchOU,
		IgnoredFields: cfg.IgnoredFields,
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Encode response as JSON
	json.NewEncoder(w).Encode(response)
}
