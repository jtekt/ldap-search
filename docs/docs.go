package docs

import (
	"encoding/json"
	"ldap-search/config"
	"log"
	"net/http"
	"strings"

	"github.com/go-openapi/spec"
	swaggerFiles "github.com/swaggo/files"
)

func generateSwaggerSpec(cfg config.Config) ([]byte, error) {
	swagger := spec.Swagger{
		SwaggerProps: spec.SwaggerProps{
			Swagger: "2.0",
			Info: &spec.Info{
				InfoProps: spec.InfoProps{
					Title:       "LDAP Search API",
					Description: "This API provides endpoints to search an LDAP based on configurable filters and attributes. It supports dynamic endpoints defined in the configuration, allowing searches by path parameters and query parameters. All endpoints except / and /health require Bearer token authentication via the Authorization header.",
					Version:     "1.0.0",
				},
			},
			Paths: &spec.Paths{
				Paths: make(map[string]spec.PathItem),
			},
			SecurityDefinitions: spec.SecurityDefinitions{
				"Bearer": &spec.SecurityScheme{
					SecuritySchemeProps: spec.SecuritySchemeProps{
						Type:        "apiKey",
						Name:        "Authorization",
						In:          "header",
						Description: "Bearer token authentication. Use format: Bearer ",
					},
				},
			},
		},
	}

	// Static endpoint: /
	swagger.Paths.Paths["/"] = spec.PathItem{
		PathItemProps: spec.PathItemProps{
			Get: &spec.Operation{
				OperationProps: spec.OperationProps{
					Summary:     "Get application metadata",
					Description: "Returns the application configuration, including endpoint details, excluding sensitive LDAP connection data",
					Tags:        []string{"Static"},
					Consumes:    []string{"application/json"},
					Produces:    []string{"application/json"},
					Responses: &spec.Responses{
						ResponsesProps: spec.ResponsesProps{
							StatusCodeResponses: map[int]spec.Response{
								200: {
									ResponseProps: spec.ResponseProps{
										Description: "OK",
										Schema: &spec.Schema{
											SchemaProps: spec.SchemaProps{
												Type: []string{"object"},
												Properties: map[string]spec.Schema{
													"port": {
														SchemaProps: spec.SchemaProps{
															Type:        []string{"string"},
															Description: "OIDC Configuration",
														},
													},
													"jwks_uri": {
														SchemaProps: spec.SchemaProps{
															Type:        []string{"string"},
															Description: "OIDC Configuration",
														},
													},
													"ldap_hostname": {
														SchemaProps: spec.SchemaProps{
															Type:        []string{"string"},
															Description: "OIDC Configuration",
														},
													},
													"ldap_port": {
														SchemaProps: spec.SchemaProps{
															Type:        []string{"string"},
															Description: "OIDC Configuration",
														},
													},
													"ldap_search_ou": {
														SchemaProps: spec.SchemaProps{
															Type:        []string{"string"},
															Description: "OIDC Configuration",
														},
													},
													"endpoints": {
														SchemaProps: spec.SchemaProps{
															Type: []string{"array"},
															Items: &spec.SchemaOrArray{
																Schema: &spec.Schema{
																	SchemaProps: spec.SchemaProps{
																		Type: []string{"object"},
																		Properties: map[string]spec.Schema{
																			"path": {
																				SchemaProps: spec.SchemaProps{
																					Type:        []string{"string"},
																					Description: "Endpoint path",
																				},
																			},
																			"filter": {
																				SchemaProps: spec.SchemaProps{
																					Type:        []string{"string"},
																					Description: "Base LDAP filter",
																				},
																			},
																			"search_params": {
																				SchemaProps: spec.SchemaProps{
																					Type:        []string{"array"},
																					Description: "Allowed search attributes",
																					Items: &spec.SchemaOrArray{
																						Schema: &spec.Schema{
																							SchemaProps: spec.SchemaProps{
																								Type:        []string{"string"},
																								Description: "Search parameter",
																							},
																						},
																					},
																				},
																			},
																			"attributes": {
																				SchemaProps: spec.SchemaProps{
																					Type:        []string{"array"},
																					Description: "Attributes to return",
																					Items: &spec.SchemaOrArray{
																						Schema: &spec.Schema{
																							SchemaProps: spec.SchemaProps{
																								Type:        []string{"string"},
																								Description: "Attribute",
																							},
																						},
																					},
																				},
																			},
																		},
																	},
																},
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	// Static endpoint: /health
	swagger.Paths.Paths["/health"] = spec.PathItem{
		PathItemProps: spec.PathItemProps{
			Get: &spec.Operation{
				OperationProps: spec.OperationProps{
					Summary:     "Health check",
					Description: "Returns 200 if healthy, else 500",
					Tags:        []string{"Static"},
					Consumes:    []string{"application/json"},
					Produces:    []string{"application/json"},
					Responses: &spec.Responses{
						ResponsesProps: spec.ResponsesProps{
							StatusCodeResponses: map[int]spec.Response{
								200: {
									ResponseProps: spec.ResponseProps{
										Description: "Healthy response",
										Schema: &spec.Schema{
											SchemaProps: spec.SchemaProps{
												Properties: map[string]spec.Schema{
													"status": {
														SchemaProps: spec.SchemaProps{
															Type:        []string{"string"},
															Description: "Should be 'ok' when healthy",
														},
													},
												},
												Required: []string{"status"},
											},
										},
									},
								},
								500: {
									ResponseProps: spec.ResponseProps{
										Description: "Unhealthy response",
									},
								},
							},
						},
					},
				},
			},
		},
	}

	for _, ep := range cfg.Endpoints {
		path := ep.Path
		operation := &spec.Operation{
			OperationProps: spec.OperationProps{
				Summary:     "Generated Endpoint: " + path,
				Description: "Dynamically generated from config. Filter: " + ep.Filter,
				Tags:        []string{"Generated"},
				Produces:    []string{"application/json"},
			},
		}

		if ep.SearchParams != nil {
			for _, param := range *ep.SearchParams {
				operation.Parameters = append(operation.Parameters, spec.Parameter{
					ParamProps: spec.ParamProps{
						Name:        param,
						In:          "query",
						Description: "Query parameter for " + param,
						Required:    false,
					},
				})
			}
		}

		for _, param := range extractPathKeys(ep.Path) {
			operation.Parameters = append(operation.Parameters, spec.Parameter{
				ParamProps: spec.ParamProps{
					Name:        param,
					In:          "path",
					Description: "Path parameter for " + param,
					Required:    true,
				},
			})
		}

		dataSchema := spec.Schema{
			SchemaProps: spec.SchemaProps{
				Type: []string{"array"},
			},
		}

		if ep.Attributes == nil || len(*ep.Attributes) == 0 {
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
						Properties: buildAttributesSchema(*ep.Attributes),
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
							Description: "Bad request",
							Schema: &spec.Schema{
								SchemaProps: spec.SchemaProps{
									Type:     []string{"object"},
									Required: []string{"error"},
									Properties: map[string]spec.Schema{
										"error": {
											SchemaProps: spec.SchemaProps{
												Type:     []string{"object"},
												Required: []string{"message", "code"},
												Properties: map[string]spec.Schema{
													"message": {
														SchemaProps: spec.SchemaProps{
															Type:        []string{"string"},
															Description: "Error message",
														},
													},
													"code": {
														SchemaProps: spec.SchemaProps{
															Type:        []string{"string"},
															Description: "Error code",
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		}

		swagger.Paths.Paths[path] = spec.PathItem{
			PathItemProps: spec.PathItemProps{
				Get: operation,
			},
		}
	}

	swagger.Schemes = []string{"http"}

	return json.MarshalIndent(swagger, "", " ")
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
	// Serve Swagger JSON
	mux.HandleFunc("/swagger.json", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
		log.Printf("Serving dynamic swagger.json")
		data, err := generateSwaggerSpec(config)
		if err != nil {
			http.Error(w, "Failed to generate Swagger spec", http.StatusInternalServerError)
			log.Printf("Swagger spec generation error: %v", err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(data)
	})

	// Serve Swagger UI
	mux.Handle("/swagger/", http.StripPrefix("/swagger/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Log the request for debugging
		log.Printf("Serving Swagger UI: %s %s", r.Method, r.URL.Path)

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
		} else if r.URL.Path == "" || r.URL.Path == "/" {
			// Redirect to the Swagger UI index
			http.Redirect(w, r, "/swagger/index.html", http.StatusFound)
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
