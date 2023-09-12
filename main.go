package main

import (
	"net/http"

	"ldap-search/auth"
	"ldap-search/controllers"
	ldapUtils "ldap-search/utils"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)




func main (){

	godotenv.Load()
	ldapUtils.Connect()
	ldapUtils.Bind()

	r := gin.Default()

	// Root route
	r.GET("/", func(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{
      "application": "ldap-search",
    })
  })

	searchController := new(controllers.SearchController)
	searchroutes := r.Group("/search")
	searchroutes.Use(auth.Middleware)
	{
		searchroutes.GET("/", searchController.Search)
	}

	r.Run() 	
	
}