package ldapUtils

import (
	"fmt"

	"ldap-search/config"

	"github.com/go-ldap/ldap/v3"
)

// Connect establishes an LDAP connection using config
func Connect(cfg config.Config) (*ldap.Conn, error) {
	connection, err := ldap.DialURL(fmt.Sprintf("ldap://%s:%s", cfg.LDAPHostname, cfg.LDAPPort))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to LDAP: %v", err)
	}
	return connection, nil
}

// Search performs an LDAP search with pagination
func Search(cfg config.Config, filter string, attributes []string) ([]*ldap.Entry, error) {
	// TODO: Add caching logic here, maybe 20 minutes
	// Depending on the query it can take a while to get results
	// For example, if the query is too broad, it can take a long time to get results

	connection, err := Connect(cfg)
	if err != nil {
		return nil, fmt.Errorf("connect error: %v", err)
	}
	defer connection.Unbind()
	defer connection.Close()

	err = connection.Bind(cfg.LDAPUsername, cfg.LDAPPassword)
	if err != nil {
		return nil, fmt.Errorf("bind error: %v", err)
	}

	// Perform search with paging
	searchReq := ldap.NewSearchRequest(
		cfg.LDAPSearchOU,
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		0,
		0,
		false,
		filter,
		attributes,
		nil,
	)

	result, err := connection.SearchWithPaging(searchReq, 1000)
	if err != nil {
		return nil, fmt.Errorf("search error: %v", err)
	}

	return result.Entries, nil
}
