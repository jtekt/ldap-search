package auth

import (
	"os"
	"net/http"
	"fmt"

	"github.com/gin-gonic/gin"




)

func Middleware(c *gin.Context){

	authHeaders := c.Request.Header["Authorization"]

	if(len(authHeaders) < 1) {
		c.AbortWithStatusJSON(403, gin.H{"message": "Authorization header not set" })
		return
	}

	authHeader := authHeaders[0]

	// TODO: check if URL is actually set
	identificationUrl := os.Getenv("IDENTIFICATION_URL")

	client := &http.Client{}


	req, err := http.NewRequest("GET", identificationUrl, nil)
	if err != nil {
			fmt.Println("Error creating HTTP request:", err)
			c.AbortWithStatusJSON(500, gin.H{"message": "Authentication failed"})
			return
	}

	req.Header.Add("Authorization", authHeader)
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
			fmt.Println("Error sending HTTP request:", err)
			c.AbortWithStatusJSON(500, gin.H{"message": "Authentication failed"})
			return
	}

	if(resp.StatusCode != 200) {
		c.AbortWithStatusJSON(resp.StatusCode, gin.H{"message": "Authentication failed"})
		return
	}

	c.Next()
}