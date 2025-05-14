package docs

import (
	"encoding/json"
	"ldap-search/config"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/go-openapi/spec"
	swaggerFiles "github.com/swaggo/files"
)

// UpdateSwaggerWithDynamicEndpoints updates the Swagger JSON with dynamic endpoints
func updateSwaggerWithDynamicEndpoints(swaggerFile string, config config.Config) error {
	// Read the existing swagger.json
	data, err := os.ReadFile(swaggerFile)
	if err != nil {
		return err
	}

	var swagger spec.Swagger
	err = json.Unmarshal(data, &swagger)
	if err != nil {
		return err
	}

	// Ensure Paths is initialized
	if swagger.Paths == nil {
		swagger.Paths = &spec.Paths{Paths: make(map[string]spec.PathItem)}
	}

	// Add security definitions for Bearer token
	swagger.SecurityDefinitions = spec.SecurityDefinitions{
		"Bearer": &spec.SecurityScheme{
			SecuritySchemeProps: spec.SecuritySchemeProps{
				Type:        "apiKey",
				Name:        "Authorization",
				In:          "header",
				Description: "Bearer token authentication. Use format: Bearer <token>",
			},
		},
	}

	// Add dynamic endpoints to the Swagger paths
	for _, config := range config.Endpoints {
		path := config.Path

		// Initialize the operation
		operation := &spec.Operation{
			OperationProps: spec.OperationProps{
				Summary:     "Dynamic Endpoint: " + path,
				Description: "Dynamic endpoint generated from config. Filter: " + config.Filter,
				Tags:        []string{"dynamic"},
				Produces:    []string{"application/json"},
			},
		}

		// Add search parameters to the operation
		for _, param := range config.SearchParams {
			operation.Parameters = append(operation.Parameters, spec.Parameter{
				ParamProps: spec.ParamProps{
					Name:        param,
					In:          "query",
					Description: "Query parameter for " + param,
				},
			})
		}

		// Add path parameters to the operation
		for _, param := range extractPathKeys(config.Path) {
			operation.Parameters = append(operation.Parameters, spec.Parameter{
				ParamProps: spec.ParamProps{
					Name:        param,
					In:          "path",
					Description: "Path parameter for " + param,
					Required:    true,
				},
			})
		}

		// Define the success response schema
		dataSchema := spec.Schema{
			SchemaProps: spec.SchemaProps{
				Type: []string{"array"},
			},
		}
		if len(config.Attributes) == 0 {
			// Use empty schema for 'any' type
			dataSchema.Items = &spec.SchemaOrArray{
				Schema: &spec.Schema{
					SchemaProps: spec.SchemaProps{
						Type: []string{"object"},
					},
				},
			}
		} else {
			dataSchema.Items = &spec.SchemaOrArray{
				Schema: &spec.Schema{
					SchemaProps: spec.SchemaProps{
						Type:       []string{"object"},
						Properties: buildAttributesSchema(config.Attributes),
					},
				},
			}
		}

		operation.Responses = &spec.Responses{
			ResponsesProps: spec.ResponsesProps{
				StatusCodeResponses: map[int]spec.Response{
					200: {
						ResponseProps: spec.ResponseProps{
							Description: "Successful response",
							Schema: &spec.Schema{
								SchemaProps: spec.SchemaProps{
									Type: []string{"object"},
									Properties: map[string]spec.Schema{
										"data": {
											SchemaProps: spec.SchemaProps{
												OneOf: []spec.Schema{
													dataSchema,
													{SchemaProps: spec.SchemaProps{Type: []string{"null"}}},
												},
											},
										},
									},
								},
							},
						},
					},
					400: {
						ResponseProps: spec.ResponseProps{
							Description: "Bad request due to invalid parameters or server error",
							Schema: &spec.Schema{
								SchemaProps: spec.SchemaProps{
									Type: []string{"object"},
									Properties: map[string]spec.Schema{
										"error": {
											SchemaProps: spec.SchemaProps{
												Type: []string{"object"},
												Properties: map[string]spec.Schema{
													"message": {
														SchemaProps: spec.SchemaProps{
															Type:        []string{"string"},
															Description: "A human-readable error message",
														},
													},
													"code": {
														SchemaProps: spec.SchemaProps{
															Type:        []string{"string"},
															Description: "A unique error code for programmatic handling",
														},
													},
												},
												Required: []string{"message", "code"},
											},
										},
									},
									Required: []string{"error"},
								},
							},
						},
					},
				},
			},
		}

		// Add to Swagger paths
		swagger.Paths.Paths[path] = spec.PathItem{
			PathItemProps: spec.PathItemProps{
				Get: operation,
			},
		}
	}

	// Write the updated swagger.json
	updatedData, err := json.MarshalIndent(swagger, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(swaggerFile, updatedData, 0644)
}

// buildAttributesSchema creates a schema for the attributes
func buildAttributesSchema(attributes []string) map[string]spec.Schema {
	props := make(map[string]spec.Schema)
	for _, attr := range attributes {
		props[attr] = spec.Schema{
			SchemaProps: spec.SchemaProps{
				Type:     []string{"string"},
				Nullable: true,
			},
		}
	}
	return props
}

// SetupSwagger configures the Swagger JSON and UI routes
func SetupSwagger(mux *http.ServeMux, config config.Config) error {
	// Update Swagger JSON with dynamic endpoints
	err := updateSwaggerWithDynamicEndpoints("./docs/swagger.json", config)
	if err != nil {
		return err
	}

	// Serve Swagger JSON
	mux.HandleFunc("/swagger.json", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
		log.Printf("Serving Swagger JSON: %s %s", r.Method, r.URL.Path)
		http.ServeFile(w, r, "./docs/swagger.json")
	})

	// Serve Swagger UI
	mux.Handle("/swagger/", http.StripPrefix("/swagger/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Log the request for debugging
		log.Printf("Swagger request: %s %s (stripped path: %s)", r.Method, r.URL.Path, r.URL.Path)

		// Ensure GET method
		if r.Method != http.MethodGet {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		// Handle swagger-initializer.js with hardcoded content
		if r.URL.Path == "swagger-initializer.js" {
			content := `
window.onload = function() {
  window.ui = SwaggerUIBundle({
    url: "/swagger.json",
    dom_id: '#swagger-ui',
    deepLinking: true,
    presets: [
      SwaggerUIBundle.presets.apis,
      SwaggerUIStandalonePreset
    ],
    plugins: [
      SwaggerUIBundle.plugins.DownloadUrl
    ],
    layout: "StandaloneLayout"
  });
};
`
			w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
			w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
			w.WriteHeader(http.StatusOK)
			_, err := w.Write([]byte(content))
			if err != nil {
				log.Printf("Error writing swagger-initializer.js: %v", err)
			}
			log.Printf("Served hardcoded swagger-initializer.js")
			return
		}

		// Serve other Swagger UI files (index.html, CSS, JS, etc.)
		swaggerFiles.Handler.ServeHTTP(w, r)
	})))

	return nil
}

func extractPathKeys(path string) []string {
	// Extract path keys from the path string
	// For example, "/api/v1/{cn}/details" -> ["cn"]
	keys := []string{}
	for _, part := range strings.Split(path, "/") {
		if strings.HasPrefix(part, "{") && strings.HasSuffix(part, "}") {
			keys = append(keys, strings.Trim(part, "{}"))
		}
	}
	return keys
}
