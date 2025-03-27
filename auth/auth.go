package auth

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
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
			jwk.WithMaxInterval(6*time.Hour),    // Max 6 hours
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
		return nil, err // Return the error from cache initialization
	}

	keyset, err = cache.CachedSet(certsUrl)
	if err != nil {
		log.Printf("[ERROR] Failed to get cached JWKS for %s: %v", certsUrl, err)
	}

	return keyset, err
}

// Middleware creates an authentication middleware for OIDC token validation
func Middleware(ctx *gin.Context) {
	// Get JWKS URL from environment variable
	certsUrl, hasEnv := os.LookupEnv("JWKS_URI")
	if !hasEnv {
		log.Printf("[ERROR] JWKS_URI not set")
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": "JWKS_URI not set",
			"code":    "ENV_MISSING",
		})
		return
	}

	// Get the cached keyset
	keyset, err := getKeyset(certsUrl)
	if err != nil {
		log.Printf("[ERROR] Failed to get cached JWKS: %v", err)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": "Authentication service unavailable",
			"code":    "JWKS_CACHE_FAILED",
		})
		return
	}

	// Check for Authorization header
	authHeader := ctx.GetHeader("Authorization")
	if authHeader == "" {
		log.Printf("[WARNING] Authorization header not set")
		ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"message": "Authorization header not set",
			"code":    "AUTH_HEADER_MISSING",
		})
		return
	}

	// Extract Bearer token
	parts := strings.Split(authHeader, "Bearer ")
	if len(parts) != 2 {
		log.Printf("[WARNING] Invalid Authorization header format: %s", authHeader)
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"message": "Invalid Authorization header format",
			"code":    "AUTH_HEADER_INVALID",
		})
		return
	}

	jwtToken := parts[1]

	// Parse and verify token using the cached keyset
	_, err = jwt.Parse([]byte(jwtToken), jwt.WithKeySet(keyset))
	if err != nil {
		log.Printf("[WARNING] Token verification failed: %v", err)
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"message": "Invalid token",
			"code":    "TOKEN_INVALID",
		})
		return
	}

	ctx.Next()
}
