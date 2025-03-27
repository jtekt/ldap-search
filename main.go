package main

import (
	"net/http"
	"os"

	"ldap-search/auth"
	"ldap-search/controllers"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	JWKS_URI := os.Getenv("JWKS_URI")
	ignoredFields := os.Getenv("IGNORED_FIELDS")

	router := gin.Default()
	router.Use(cors.Default())

	// Root route
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"application":    "ldap-search",
			"JWKS_URI":       JWKS_URI,
			"ignored_fields": ignoredFields,
		})
	})

	protectedRoutes := router.Group("/")

	if JWKS_URI != "" {
		protectedRoutes.Use(auth.Middleware)
	}

	searchRoutes := protectedRoutes.Group("/search")
	searchController := new(controllers.SearchController)
	searchRoutes.GET("/", searchController.Search)

	router.Run()
}
