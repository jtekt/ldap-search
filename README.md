# LDAP API with OIDC Authentication

This is a simple LDAP API built with Go that provides a search functionality via the `/search?parameter=any` endpoint. The API requires authentication using OpenID Connect (OIDC).

## Features

- Authenticate requests with OIDC
- Search LDAP directory using query parameters
- Ignore specified fields in the search results

## Environment Variables

To configure the API, set the following environment variables:

```env
# OIDC Configuration
JWKS_URI=<OIDC JWKS URI>

# LDAP Configuration
LDAP_HOSTNAME=<LDAP server hostname>
LDAP_SEARCH_OU=<LDAP search base, e.g., ou=users,dc=example,dc=com>
LDAP_USERNAME=<LDAP bind username>
LDAP_PASSWORD=<LDAP bind password>

# Fields to ignore in the search results (comma-separated list)
IGNORED_FIELDS=<e.g., password,ssn,creditCard>
```

## Usage

### Run the API

```sh
go run .
```

### Search Endpoint

#### Request

```http
GET /search?email=example@jtekt.co.jp
Authorization: Bearer <OIDC Access Token>
```

#### Response

```json
[
  {
    "objectClass": "top,person,organizationalPerson,user",
    "cn": "00000000",
    "sn": "example",
    "title": "C",
    "givenName": "example",
    "mail": "example@jtekt.co.jp",
    "uid": "00000000"
  }
]
```

## Requirements

- Go 1.20+
- A running LDAP server
- An OIDC provider for authentication

## License

This project is licensed under the MIT License.
