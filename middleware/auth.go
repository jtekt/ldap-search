package middleware

import (
	"context"
	"fmt"
	"ldap-search/config"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/lestrrat-go/httprc/v3"
	"github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/lestrrat-go/jwx/v3/jwt"
)

var cache *jwk.Cache
var cacheInit sync.Once

// getKeyset retrieves the JWKS keyset, initializing the cache if necessary.
func getKeyset(certsUrl string) (jwk.Set, error) {
	var keyset jwk.Set
	var err error

	cacheInit.Do(func() {
		// Create a background context for the cache
		bgCtx := context.Background()

		// Create the cache with an HTTP client
		cache, err = jwk.NewCache(bgCtx, httprc.NewClient())
		if err != nil {
			log.Printf("[ERROR] Failed to create JWKS cache: %v", err)
			return
		}

		// Register the JWKS URL with refresh intervals
		err = cache.Register(bgCtx, certsUrl,
			jwk.WithMaxInterval(24*time.Hour),   // Max 24 hours
			jwk.WithMinInterval(15*time.Minute), // Min 15 minutes
		)
		if err != nil {
			log.Printf("[ERROR] Failed to register JWKS URL %s: %v", certsUrl, err)
			return
		}

		// Pre-fetch once to ensure we have keys at startup
		_, err = cache.CachedSet(certsUrl)
		if err != nil {
			log.Printf("[ERROR] Failed to pre-fetch JWKS from %s: %v", certsUrl, err)
			return
		}
	})

	if cache == nil {
		log.Printf("[ERROR] Cache is nil, initialization failed")
		return nil, err
	}

	keyset, err = cache.CachedSet(certsUrl)
	if err != nil {
		log.Printf("[ERROR] Failed to get cached JWKS for %s: %v", certsUrl, err)
	}

	return keyset, err
}

// Middleware creates an authentication middleware for OIDC token validation
func Auth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get JWKS URL from environment variable
		config := config.GetConfig()
		if len(config.JWKSURI) == 0 {
			log.Printf("[ERROR] JWKS_URI not set")
			http.Error(w, fmt.Sprintf(`{"message": "%s", "code": "%v"}`, "JWKS_URI not set", "ENV_MISSING"), http.StatusInternalServerError)
			return
		}

		// Get the cached keyset
		keyset, err := getKeyset(config.JWKSURI)
		if err != nil {
			log.Printf("[ERROR] Failed to get cached JWKS: %v", err)
			http.Error(w, fmt.Sprintf(`{"message": "%s", "code": "%v"}`, err.Error(), "JWKS_CACHE_FAILED"), http.StatusInternalServerError)
			return
		}

		// Check for Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			log.Printf("[WARNING] Authorization header not set")
			http.Error(w, fmt.Sprintf(`{"message": "%s", "code": "%v"}`, "Authorization header not set", "AUTH_HEADER_MISSING"), http.StatusForbidden)
			return
		}

		// Extract Bearer token
		parts := strings.Split(authHeader, "Bearer ")
		if len(parts) != 2 {
			log.Printf("[WARNING] Invalid Authorization header format: %s", authHeader)
			http.Error(w, fmt.Sprintf(`{"message": "%s", "code": "%v"}`, "Invalid Authorization header format", "AUTH_HEADER_INVALID"), http.StatusUnauthorized)
			return
		}

		jwtToken := parts[1]

		// Parse and verify token using the cached keyset
		_, err = jwt.Parse([]byte(jwtToken), jwt.WithKeySet(keyset))
		if err != nil {
			log.Printf("[WARNING] Token verification failed: %v", err)
			http.Error(w, fmt.Sprintf(`{"message": "%s", "code": "%v"}`, "Invalid token", "TOKEN_INVALID"), http.StatusUnauthorized)
			return
		}

		// Token is valid, proceed to next handler
		next.ServeHTTP(w, r)
	}
}
