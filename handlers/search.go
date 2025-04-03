package handlers

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

// formatFilter constructs an LDAP filter from URL parameters.
func formatFilter(query url.Values) string {
	filters := ""
	for key, values := range query {
		filters += fmt.Sprintf("(%v=%v)", key, values[0])
	}
	return fmt.Sprintf("(&%v)", filters)
}

// stringInSlice checks if a string exists in a list.
func stringInSlice(s string, list []string) bool {
	return slices.Contains(list, s)
}

// formatEntries processes LDAP entries while ignoring specified fields.
func formatEntries(entries []*ldap.Entry) []map[string]interface{} {
	ignoredFields := strings.Split(os.Getenv("IGNORED_FIELDS"), ",")

	var output []map[string]interface{}
	for _, entry := range entries {
		entryMap := make(map[string]interface{})
		for _, attr := range entry.Attributes {
			if !stringInSlice(attr.Name, ignoredFields) {
				if len(attr.Values) == 1 {
					entryMap[attr.Name] = attr.Values[0] // Store as string
				} else {
					entryMap[attr.Name] = attr.Values // Store as []string
				}
			}
		}
		output = append(output, entryMap)
	}

	return output
}

// SearchHandler handles LDAP search requests.
func SearchHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	query := r.URL.Query()
	filter := formatFilter(query)

	entries, err := ldapUtils.Search(filter)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"message": "%s"}`, err.Error()), http.StatusBadRequest)
		return
	}

	response := map[string]interface{}{"items": formatEntries(entries)}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Encode response as JSON
	json.NewEncoder(w).Encode(response)
}
