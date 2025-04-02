package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"slices"
	"strings"

	ldapUtils "ldap-search/utils"

	"github.com/go-ldap/ldap/v3"
)

type SearchController struct{}

func formatFilter(fullQuery url.Values) string {
	filters := ""
	for key, values := range fullQuery {
		filters += fmt.Sprintf("(%v=%v)", key, values[0])
	}
	return fmt.Sprintf("(&%v)", filters)
}

func stringInSlice(a string, list []string) bool {
	return slices.Contains(list, a)
}

func formatEntries(entries []*ldap.Entry) []map[string]interface{} {
	ignoredFieldsEnvString := os.Getenv("IGNORED_FIELDS")
	ignoredFields := strings.Split(ignoredFieldsEnvString, ",")

	var outputSlice []map[string]interface{}
	for _, entry := range entries {
		entryMap := make(map[string]interface{})
		for _, attribute := range entry.Attributes {
			if !stringInSlice(attribute.Name, ignoredFields) {
				if len(attribute.Values) == 1 {
					entryMap[attribute.Name] = attribute.Values[0] // Store as string
				} else {
					entryMap[attribute.Name] = attribute.Values // Store as []string
				}
			}
		}
		outputSlice = append(outputSlice, entryMap)
	}

	return outputSlice
}

func (p *SearchController) Search(w http.ResponseWriter, r *http.Request) {
	// Only allow GET method
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get query parameters
	fullQuery := r.URL.Query()

	filter := formatFilter(fullQuery)

	entries, err := ldapUtils.Search(filter)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest) // 400
		json.NewEncoder(w).Encode(map[string]string{"message": err.Error()})
		return
	}

	formattedEntries := formatEntries(entries)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{"items": formattedEntries})
}
