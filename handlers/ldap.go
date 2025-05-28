package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"ldap-search/cache"
	"ldap-search/config"
	"ldap-search/middleware"
	ldapUtils "ldap-search/utils"

	"github.com/go-ldap/ldap/v3"

	gocache "github.com/go-redis/cache/v9"
)

// LdapHandler handles LDAP-related requests
type LdapHandler struct {
	config config.Config
}

func NewLdapHandler(cfg config.Config) *LdapHandler {
	return &LdapHandler{config: cfg}
}

// RegisterLdapEndpoints registers all LDAP endpoints from the config
func RegisterLdapEndpoints(mux *http.ServeMux, cfg config.Config, redisCache *cache.Cache) {
	handler := NewLdapHandler(cfg)

	for _, ep := range cfg.Endpoints {
		mux.Handle("GET "+ep.Path, middleware.ChainMiddleware(handler.handleLdapSearch(ep, redisCache), middleware.Logging, middleware.Cors, middleware.Auth))
		log.Printf("Registered endpoint: %s with filter: %s", ep.Path, ep.Filter)
	}
}

// formatResponse standardizes the API response format
type apiResponse struct {
	Data  interface{} `json:"data,omitempty"`
	Error *apiError   `json:"error,omitempty"`
}

type apiError struct {
	Message string `json:"message"`
	Code    string `json:"code"`
}

// formatResponse creates a standardized response with data or error
func formatResponse(w http.ResponseWriter, status int, data interface{}, err error, errCode string) {
	response := apiResponse{}

	if err != nil {
		response.Error = &apiError{
			Message: err.Error(),
			Code:    errCode,
		}
	} else {
		response.Data = data
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if encodeErr := json.NewEncoder(w).Encode(response); encodeErr != nil {
		// Fallback if encoding fails
		http.Error(w, `{"error":{"message":"Failed to encode response","code":"ENCODING_ERROR"}}`, http.StatusInternalServerError)
	}
}

// handleLdapSearch handles LDAP searches with search parameters and pagination
func (h *LdapHandler) handleLdapSearch(endpoint config.EndpointConfig, redisCache *cache.Cache) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract attributes and base filter
		attributes := endpoint.Attributes
		searchParams := endpoint.SearchParams
		filter := endpoint.Filter

		query := r.URL.Query()
		pathParams := extractPathParams(endpoint.Path, r.URL.Path)

		// Create a map of valid search parameters
		validMap := make(map[string]bool)
		if searchParams != nil {
			for _, key := range *searchParams {
				validMap[key] = true
			}
		}

		// Validate: check if any query key is not in validKeys
		invalidQueryParams := []string{}
		for key := range query {
			if !validMap[key] {
				invalidQueryParams = append(invalidQueryParams, key)
			}
		}

		// If there are invalid query parameters, return an error
		if len(invalidQueryParams) > 0 {
			formatResponse(w, http.StatusBadRequest, nil, fmt.Errorf("invalid query parameters: %v", strings.Join(invalidQueryParams, ", ")), "INVALID_QUERY_PARAMS")
			return
		}

		// First: Replace path parameters like {cn} in the filter
		for _, param := range getPathParams(endpoint.Path) {
			if value, ok := pathParams[param]; ok {
				filter = strings.ReplaceAll(filter, "{"+param+"}", sanitize(value))
			}
		}

		searchFilters := []string{}
		// Second: Add filters from SearchParams (from ?query=...)
		if searchParams != nil {
			for _, param := range *searchParams {
				if value := query.Get(param); value != "" {
					searchFilters = append(searchFilters, fmt.Sprintf("(%s=%s)", param, sanitize(value)))
				}
			}
		}

		// Append additional search filters if any
		if len(searchFilters) > 0 {
			if strings.HasPrefix(filter, "(&") {
				// Inject additional filters directly
				filter = strings.TrimSuffix(filter, ")") + strings.Join(searchFilters, "") + ")"
			} else {
				// Wrap original filter with additional filters
				filter = fmt.Sprintf("(&%s%s)", filter, strings.Join(searchFilters, ""))
			}
		}

		var entries []*ldap.Entry
		var err error

		// Only use caching if RedisAddr is configured
		if h.config.RedisAddr != "" && redisCache != nil {
			// Try to get from cache
			cacheKey := cache.HashKey(filter)
			err = redisCache.Get(r.Context(), cacheKey, &entries)
			if err != nil {
				// Cache miss or error — perform LDAP search
				entries, err = ldapUtils.Search(h.config, filter, attributes)
				if err != nil {
					formatResponse(w, http.StatusInternalServerError, nil, fmt.Errorf("LDAP search failed: %v", err), "LDAP_SEARCH_ERROR")
					return
				}
				// Cache the result
				cacheErr := redisCache.Set(&gocache.Item{
					Ctx:   r.Context(),
					Key:   cacheKey,
					Value: entries,
					TTL:   24 * time.Hour, // Cache for 24 hours
				})
				if cacheErr != nil {
					log.Printf("[WARN] Failed to cache LDAP result: %v", cacheErr)
				}
			}
		} else {
			// No caching - direct LDAP search
			entries, err = ldapUtils.Search(h.config, filter, attributes)
			if err != nil {
				formatResponse(w, http.StatusInternalServerError, nil, fmt.Errorf("LDAP search failed: %v", err), "LDAP_SEARCH_ERROR")
				return
			}
		}

		// Format entries
		data := formatEntries(entries)

		formatResponse(w, http.StatusOK, data, nil, "")
	}
}

// formatEntries converts LDAP entries to a configurable JSON format
func formatEntries(entries []*ldap.Entry) []map[string]interface{} {
	var result []map[string]interface{}

	// Get parsed properties from the config
	cfg := config.GetConfig()

	for _, entry := range entries {
		entryMap := make(map[string]interface{})

		for _, attr := range entry.Attributes {
			attrNameLower := strings.ToLower(attr.Name)

			if cfg.ParsedProperties[attrNameLower] {
				// Parse each DN string into a map
				var parsedDNs []map[string]interface{}
				for _, dn := range attr.Values {
					parsedDNs = append(parsedDNs, parseDNToMap(dn))
				}
				entryMap[attr.Name] = parsedDNs
			} else {
				if len(attr.Values) == 1 {
					entryMap[attr.Name] = attr.Values[0]
				} else {
					entryMap[attr.Name] = strings.Join(attr.Values, ",")
				}
			}
		}

		result = append(result, entryMap)
	}
	return result
}

func parseDNToMap(dn string) map[string]interface{} {
	result := make(map[string]interface{})
	pairs := strings.Split(dn, ",")

	for _, pair := range pairs {
		kv := strings.SplitN(strings.TrimSpace(pair), "=", 2)
		if len(kv) != 2 {
			continue
		}
		key := kv[0]
		value := kv[1]

		// Handle multiple values for same key (e.g., DC=example,DC=com)
		if existing, ok := result[key]; ok {
			switch v := existing.(type) {
			case []string:
				result[key] = append(v, value)
			case string:
				result[key] = []string{v, value}
			}
		} else {
			result[key] = value
		}
	}
	return result
}

// Allow letters, digits, space, dash, underscore, dot, and @
var allowedChars = regexp.MustCompile(`[^a-zA-Z0-9 ._@-]`)

// sanitize removes disallowed characters from a string
func sanitize(value string) string {
	// Remove characters not in the allowed set
	return allowedChars.ReplaceAllString(value, "")
}

// getPathParams extracts keys from a path pattern like /groups/{cn}
func getPathParams(path string) []string {
	var params []string
	parts := strings.Split(path, "/")
	for _, part := range parts {
		if strings.HasPrefix(part, "{") && strings.HasSuffix(part, "}") {
			params = append(params, strings.Trim(part, "{}"))
		}
	}
	return params
}

// extractPathParams extracts actual values from a request path based on a path pattern
func extractPathParams(pattern, actualPath string) map[string]string {
	patternParts := strings.Split(strings.Trim(pattern, "/"), "/")
	actualParts := strings.Split(strings.Trim(actualPath, "/"), "/")

	params := make(map[string]string)
	if len(patternParts) != len(actualParts) {
		return params
	}

	for i, part := range patternParts {
		if strings.HasPrefix(part, "{") && strings.HasSuffix(part, "}") {
			key := strings.Trim(part, "{}")
			params[key] = actualParts[i]
		}
	}
	return params
}
