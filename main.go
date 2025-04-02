package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"ldap-search/auth"
	"ldap-search/controllers"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Printf("[ERROR] Failed to load .env file: %v", err)
		return
	}

	JWKS_URI, hasEnv := os.LookupEnv("JWKS_URI")
	if !hasEnv {
		log.Printf("[ERROR] JWKS_URI not set")
		return
	}

	ignoredFields := os.Getenv("IGNORED_FIELDS")

	// Create the base mux
	mux := http.NewServeMux()

	// Define root route (unprotected)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Handle the case when redirected to root
		if r.RequestURI != "/" {
			http.NotFound(w, r)
			return
		}

		// Handle method
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		response := map[string]string{
			"application":    "ldap-search",
			"JWKS_URI":       JWKS_URI,
			"ignored_fields": ignoredFields,
		}

		json.NewEncoder(w).Encode(response)
	})

	// Define search route
	searchController := new(controllers.SearchController)

	// Wrap the Search function in an http.HandlerFunc and apply auth middleware
	// Obs. if the request URL contains a in the end "/" like "/search/" it will not match and redirect to "/"
	mux.Handle("/search", auth.Middleware(http.HandlerFunc(searchController.Search)))

	// Apply CORS to the entire mux
	handler := cors(mux)

	// Start server with logging
	log.Printf("[INFO] Starting server on :8080")
	err = http.ListenAndServe(":8080", handler)
	if err != nil {
		log.Fatalf("[ERROR] Server failed to start: %v", err)
	}
}

// CORS middleware
func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}

		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		w.WriteHeader(http.StatusOK)
	})
}
