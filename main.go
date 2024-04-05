package main

import (
	"fmt"
	"net/http"
	"os"

	"ldap-search/auth"
	"ldap-search/controllers"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)




func main (){

	godotenv.Load()

	r := gin.Default()

	identification_url := os.Getenv("IDENTIFICATION_URL")
	ignoredFields := os.Getenv("IGNORED_FIELDS")

	// Root route
	r.GET("/", func(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{
      "application": "ldap-search",
			"identification_url": identification_url,
			"ignored_fields": ignoredFields,
    })
  })

	searchController := new(controllers.SearchController)
	searchroutes := r.Group("/search")

	if identification_url != "" {
		fmt.Println("[Auth] Authentication is enabled")
		searchroutes.Use(auth.Middleware)
	} else {	
		fmt.Println("[Auth] Authentication is NOT enabled")
	}

	searchroutes.GET("/", searchController.Search)
	

	r.Run() 	
	
}