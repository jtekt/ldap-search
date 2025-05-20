package config

import (
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

// EndpointConfig defines an API endpoint and its LDAP query parameters
type EndpointConfig struct {
	Path         string    `yaml:"path" json:"path"`                   // e.g., "/api/v1/groups"
	Filter       string    `yaml:"filter" json:"filter"`               // Base LDAP filter
	SearchParams *[]string `yaml:"search_params" json:"search_params"` // Allowed search attributes
	Attributes   *[]string `yaml:"attributes" json:"attributes"`       // Attributes to return
}

// Config holds the application configuration
type Config struct {
	Port string `json:"port"`

	// OIDC Configuration
	JWKSURI string `json:"jwks_uri"`

	// LDAP Configuration
	LDAPHostname string `json:"ldap_hostname"`
	LDAPPort     string `json:"ldap_port"`
	LDAPSearchOU string `json:"ldap_search_ou"`
	LDAPUsername string `json:"-"` // Don't expose username in JSON responses
	LDAPPassword string `json:"-"` // Don't expose passwords in JSON responses

	// Redis Configuration
	RedisAddr string `json:"-"` // Don't expose passwords in JSON responses

	// Endpoint configurations
	Endpoints []EndpointConfig `json:"endpoints"`
}

// LoadConfig loads the configuration from environment variables and YAML files
func LoadConfig() Config {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Printf("[WARN] No .env file found or failed to load: %v (continuing...)", err)
	}

	// Load base configuration from environment variables
	return GetConfig()
}

// GetConfig retrieves the configuration from environment variables and YAML
// It uses default values if the environment variables are not set
func GetConfig() Config {
	return Config{
		// Server
		Port: getEnv("PORT", "8080"),

		// OIDC
		JWKSURI: getEnv("JWKS_URI", ""),

		// LDAP
		LDAPHostname: getEnv("LDAP_HOSTNAME", ""),
		LDAPPort:     getEnv("LDAP_PORT", "389"),
		LDAPSearchOU: getEnv("LDAP_SEARCH_OU", ""),
		LDAPUsername: getEnv("LDAP_USERNAME", ""),
		LDAPPassword: getEnv("LDAP_PASSWORD", ""),

		// Redis
		RedisAddr: getEnv("REDIS_ADDR", "localhost:6379"),

		// Endpoints
		Endpoints: loadEndpointsConfig(),
	}
}

// Loads endpoint configurations from a YAML file
func loadEndpointsConfig() []EndpointConfig {
	var endpoints []EndpointConfig

	// Read endpoints.yaml
	data, err := os.ReadFile("config/endpoints.yaml")
	if err != nil {
		log.Printf("[WARN] Failed to read endpoints.yaml: %v (continuing with no endpoints)", err)
		return endpoints
	}

	// Unmarshal YAML
	if err := yaml.Unmarshal(data, &endpoints); err != nil {
		log.Fatalf("failed to unmarshal endpoints config: %v", err)
	}

	// Validate endpoints
	for _, ep := range endpoints {
		if ep.Path == "" || ep.Filter == "" {
			log.Fatalf("endpoint path and filter are required: %+v", ep)
		}
		// Validate placeholder count
		pathPlaceholders := strings.Count(ep.Path, "{}")
		filterPlaceholders := strings.Count(ep.Filter, "{}") + strings.Count(ep.Filter, "{LDAPSearchOU}")
		if pathPlaceholders != filterPlaceholders {
			log.Fatalf("mismatch in placeholders: path (%s) has %d '{}', filter (%s) has %d '{}'",
				ep.Path, pathPlaceholders, ep.Filter, filterPlaceholders)
		}
	}

	return endpoints
}

// getEnv fetches environment variables with a default value
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
