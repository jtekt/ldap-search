package main

import (
	"log"
	"net/http"

	"ldap-search/cache"
	"ldap-search/config"
	"ldap-search/docs"
	"ldap-search/handlers"
	"ldap-search/middleware"
)

// @title LDAP Search API
// @version 1.0.0
// @description This API provides endpoints to search an LDAP directory based on configurable filters and attributes. It supports dynamic endpoints defined in the configuration, allowing searches by path parameters (e.g., /api/{id}) and query parameters (e.g., ?search=value). All endpoints except / and /health require Bearer token authentication via the Authorization header.
// @contact.name Support Team
// @contact.email kevin_fabel@jtekt.co.jp
// @BasePath /
func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Initialize cache
	redisCache := cache.New(cfg.RedisAddr)

	// Create the base mux
	mux := http.NewServeMux()

	// Home endpoint
	mux.Handle("GET /{$}", middleware.ChainMiddleware(handlers.HomeHandler, middleware.Logging, middleware.Cors))
	mux.Handle("GET /health", middleware.ChainMiddleware(handlers.HealthHandler, middleware.Logging, middleware.Cors))

	// Register LDAP endpoints from config
	handlers.RegisterLdapEndpoints(mux, cfg, redisCache)

	// Docs
	docs.SetupSwagger(mux, cfg)

	// Start server with logging
	log.Printf(`[INFO] Starting server on port "%s"`, cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, mux))
}
