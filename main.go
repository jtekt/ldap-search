package main

import (
	"fmt"
	"ldap-search/utils"
	"ldap-search/controllers"
	"ldap-search/auth"

	"github.com/joho/godotenv"
	"github.com/gin-gonic/gin"



)




func main (){

	r := gin.Default()

	searchController := new(controllers.SearchController)

	searchroutes := r.Group("/search")
	searchroutes.Use(auth.Middleware)
	{
		searchroutes.GET("/", searchController.Search)
	}

	godotenv.Load()
	ldapUtils.Connect()
	ldapUtils.Bind()
	

	r.Run() 

	fmt.Println("Gin started")
	
	
}