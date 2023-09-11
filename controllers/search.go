package controllers

import (
	"fmt"
	"net/http"
	"net/url"

	ldapUtils "ldap-search/utils"

	"github.com/gin-gonic/gin"
)


type SearchController struct{}

func formatFilter(fullQuery url.Values) string{
	filters := ""
	for key, values := range fullQuery {
		filters += fmt.Sprintf("(%v=%v)", key, values[0])
	}
	return fmt.Sprintf("(&%v)", filters)
}

func (p SearchController) Search(c *gin.Context) {


	fullQuery := c.Request.URL.Query()
	filter := formatFilter(fullQuery)
	

	result, err := ldapUtils.Search(filter)
	
	if err != nil {
		// TODO: don't panic
		fmt.Println(err)
		panic(err)
	}

	// TODO: formatting


	c.JSON(http.StatusOK, result)

}