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
