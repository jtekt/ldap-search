package controllers

import (
	"fmt"
	"net/http"
	"net/url"

	ldapUtils "ldap-search/utils"

	"github.com/gin-gonic/gin"
	"github.com/go-ldap/ldap/v3"
)


type SearchController struct{}

func formatFilter(fullQuery url.Values) string{
	filters := ""
	for key, values := range fullQuery {
		filters += fmt.Sprintf("(%v=%v)", key, values[0])
	}
	return fmt.Sprintf("(&%v)", filters)
}

func formatEntries(entries []*ldap.Entry) []map[string]interface{} {
	
	var outputSlice []map[string]interface{}
	for _, entry := range entries {
		entryMap := make(map[string]interface{})
		for _, attribute := range entry.Attributes {
			if(len(attribute.Values) == 1) {
				entryMap[attribute.Name] = attribute.Values[0]
			} else {
				entryMap[attribute.Name] = attribute.Values
			}
			
		}
		outputSlice = append(outputSlice, entryMap)
	}

	return outputSlice
	
}

func (p SearchController) Search(c *gin.Context) {


	fullQuery := c.Request.URL.Query()
	
	filter := formatFilter(fullQuery)
	

	entries, err := ldapUtils.Search(filter)
	
	if err != nil {
		c.AbortWithStatusJSON(400, gin.H{"message": err })
		return
	}

	// TODO: formatting
	formattedEntries := formatEntries(entries)

	c.JSON(http.StatusOK, formattedEntries)

}