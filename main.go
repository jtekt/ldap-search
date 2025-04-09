package main

import (
	"log"
	"net/http"

	"ldap-search/config"
	"ldap-search/handlers"
	"ldap-search/middleware"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Create the base mux
	mux := http.NewServeMux()

	// Home endpoint
	mux.Handle("GET /{$}", middleware.ChainMiddleware(handlers.HomeHandler, middleware.Logging, middleware.Cors))

	// Ldap endpoints
	mux.Handle("GET /search", middleware.ChainMiddleware(handlers.SearchHandler, middleware.Logging, middleware.Cors, middleware.Auth))

	// Start server with logging
	log.Printf(`[INFO] Starting server on port "%s"`, cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, mux))
}
