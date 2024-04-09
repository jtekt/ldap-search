package main

import (
	"fmt"
	"net/http"
	"os"

	"ldap-search/auth"
	"ldap-search/controllers"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)




func main (){

	godotenv.Load()

	identification_url := os.Getenv("IDENTIFICATION_URL")
	ignoredFields := os.Getenv("IGNORED_FIELDS")

	router := gin.Default()
	router.Use(cors.Default())


	// Root route
	router.GET("/", func(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{
      "application": "ldap-search",
			"identification_url": identification_url,
			"ignored_fields": ignoredFields,
    })
  })


	searchController := new(controllers.SearchController)
	searchroutes := router.Group("/search")

	if identification_url != "" {
		fmt.Println("[Auth] Authentication is enabled")
		searchroutes.Use(auth.Middleware)
	} else {	
		fmt.Println("[Auth] Authentication is NOT enabled")
	}

	searchroutes.GET("/", searchController.Search)
	

	router.Run() 	
	
}