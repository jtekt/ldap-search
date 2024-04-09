package ldapUtils

import (
	"fmt"
	"os"

	"github.com/go-ldap/ldap/v3"
)

// Note the return value
func Connect() (*ldap.Conn, error) {

	hostname := os.Getenv("LDAP_HOSTNAME")
	port := os.Getenv("LDAP_PORT")

	connection, err := ldap.DialURL(fmt.Sprintf("ldap://%s:%s", hostname, port))

	if err != nil {
		return nil, err
	}

	return connection, nil

}


func Search(filter string)([]*ldap.Entry, error){

	connection, err := Connect()

	if err != nil {
		return nil, fmt.Errorf("Connect error: %s", err)
	}

	username := os.Getenv("LDAP_USERNAME")
	password := os.Getenv("LDAP_PASSWORD")
	connection.Bind(username,password)


	defer connection.Unbind()
	defer connection.Close()

	// controls := []ldap.Control{
	// 	ldap.NewControlPaging(3),
	// }

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
		nil, // controls,
	)

	// result, err := connection.Search(searchReq)
	result, err := connection.SearchWithPaging(searchReq,100)

	// updatedControl := ldap.FindControl(result.Controls, ldap.ControlTypePaging)


	if err != nil {
		return nil, err
	}

	if len(result.Entries) > 0 {
		return result.Entries, nil
	} else {
		return nil, fmt.Errorf("couldn't fetch search entries")
	}

	
}

