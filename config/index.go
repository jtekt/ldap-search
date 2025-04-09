package config

import (
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// Config holds the application configuration
type Config struct {
	Port string `json:"port"`

	// OIDC Configuration
	JWKSURI string `json:"jwks_uri"`

	// LDAP Configuration
	LDAPHostname string `json:"ldap_hostname"`
	LDAPSearchOU string `json:"ldap_search_ou"`
	LDAPUsername string `json:"-"` // Don't expose username in JSON responses
	LDAPPassword string `json:"-"` // Don't expose passwords in JSON responses

	// Fields to ignore in LDAP search results
	IgnoredFields []string `json:"ignored_fields"`
}

// GetConfig loads configuration from environment variables
func GetConfig() Config {
	return Config{
		// OIDC
		Port: getEnv("PORT", "8080"),

		// OIDC
		JWKSURI: getEnv("JWKS_URI", ""),

		// LDAP
		LDAPHostname: getEnv("LDAP_HOSTNAME", ""),
		LDAPSearchOU: getEnv("LDAP_SEARCH_OU", ""),
		LDAPUsername: getEnv("LDAP_USERNAME", ""),
		LDAPPassword: getEnv("LDAP_PASSWORD", ""),

		// Ignored fields (converted from CSV string to slice)
		IgnoredFields: parseCommaSeparatedEnv("IGNORED_FIELDS"),
	}
}

// LoadConfig initiate the dotenv and loads configuration from environment variables
func LoadConfig() Config {
	// Load env
	err := godotenv.Load()
	if err != nil {
		log.Printf("[WARN] No .env file found or failed to load: %v (continuing...)", err)
	}

	return GetConfig()
}

// getEnv fetches environment variables with a default value
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// parseCommaSeparatedEnv converts a comma-separated environment variable to a string slice
func parseCommaSeparatedEnv(key string) []string {
	value := getEnv(key, "")
	if value == "" {
		return []string{}
	}
	return strings.Split(value, ",")
}
