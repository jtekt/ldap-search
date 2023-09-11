package ldapUtils

import (
	"fmt"
	"os"

	"github.com/go-ldap/ldap/v3"
)


var connection *ldap.Conn

// Note the return value
func Connect()  {

	hostname := os.Getenv("LDAP_HOSTNAME")
	port := os.Getenv("LDAP_PORT")

	_connection, err := ldap.DialURL(fmt.Sprintf("ldap://%s:%s", hostname, port))
	connection = _connection

	if err != nil {
			panic(err)
	}

}

func Bind(){
	username := os.Getenv("LDAP_USERNAME")
	password := os.Getenv("LDAP_PASSWORD")
	connection.Bind(username,password)
}

func Search(filter string)(*ldap.SearchResult, error){

	baseDN :=  os.Getenv("LDAP_SEARCH_OU")
	searchReq := ldap.NewSearchRequest(
		baseDN,
		ldap.ScopeWholeSubtree, 
		ldap.NeverDerefAliases,
		0,
		0,
		false,
		filter,
		[]string{},
		nil,
	)
	result, err := connection.Search(searchReq)

	if err != nil {
		return nil, fmt.Errorf("Search Error: %s", err)
	}

	if len(result.Entries) > 0 {
		return result, nil
	} else {
		return nil, fmt.Errorf("Couldn't fetch search entries")
	}

	
}

