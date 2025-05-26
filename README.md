# LDAP Search API

This is a Go-based API for searching an LDAP directory using dynamically configured endpoints. It supports OpenID Connect (OIDC) authentication via Bearer tokens for secure access to LDAP search endpoints, with a standardized response format for success and error cases. The API includes a health check endpoint and interactive Swagger UI for documentation.

## Features

- **Dynamic Endpoints**: Configurable LDAP search endpoints (e.g., `/api/users`, `/api/{id}`) defined in a YAML configuration.
- **OIDC Authentication**: Requires Bearer token authentication for dynamic endpoints, validated using an OIDC provider’s JWKS URI.
- **Swagger UI**: Interactive API documentation at `/swagger/` (publicly accessible).

## Setup

### Prerequisites

- Go 1.23 or higher
- A running LDAP server
- An OIDC provider (e.g., Keycloak, Auth0) with a JWKS URI

### Installation

1. Clone the repository:

   ```sh
   git clone <repository-url>
   cd <repository-directory>
   ```

2. Install dependencies:

   ```bash
   go mod tidy
   ```

3. Configure environment variables (see [Environment Variables](#environment-variables) below).

### Run the API

Run the API locally:

```bash
go run .
```

The API will start on `http://localhost:8080` (or the port specified in your configuration).

## API Documentation

Access the interactive Swagger UI at:

```
http://localhost:8080/swagger/
```

The Swagger UI provides detailed documentation for all endpoints, including schemas, parameters, and authentication requirements. The raw OpenAPI specification is available at `http://localhost:8080/swagger.json`.

## Endpoints

### Health Check

- **URL**: `GET /health`
- **Authentication**: None (public)
- **Description**: Verifies the API is running.
- **Response**:
  ```json
  {
    "status": "ok"
  }
  ```

### Dynamic LDAP Search Endpoints

- **URL**: Configured in `config.yaml` (e.g., `GET /api/v1/groups`, `GET /api/v1/groups/{id}`)
- **Authentication**: Required (`Authorization: Bearer <OIDC Access Token>`)
- **Description**: Searches the LDAP directory using path parameters (e.g., `{id}`) and query parameters (e.g., `?email=value`). The response includes LDAP entries with specified attributes, excluding ignored fields.
- **Example Request**:
  ```http
  GET /api/v1/groups?cn=example_cn
  Authorization: Bearer <OIDC Access Token>
  ```
- **Success Response** (200):
  ```json
  {
    "data": [
      {
        "cn": "string",
        "department": "string",
        "displayName": "string",
        "distinguishedName": "string",
        "info": "string",
        "mail": "string",
        "memberOf": "string",
        "msDS-PhoneticDisplayName": "string",
        "whenChanged": "string",
        "whenCreated": "string"
      }
    ]
  }
  ```
- **Empty Response** (200):
  ```json
  {
    "data": null
  }
  ```
- **Error Response** (400, 500, etc.):
  ```json
  {
    "error": {
      "message": "LDAP search failed: invalid filter",
      "code": "LDAP_SEARCH_ERROR"
    }
  }
  ```

## Environment Variables

Configure the API using the following environment variables (e.g., in a `.env` file or via your environment):

Example `.env` file:

```env
JWKS_URI=https://your-oidc-provider/.well-known/jwks.json

LDAP_HOSTNAME=ldap.example.com
LDAP_SEARCH_OU=ou=users,dc=example,dc=com
LDAP_PORT=389
LDAP_USERNAME=admin
LDAP_PASSWORD=secret

PORT=8080

REDIS_ADDR=localhost:30379

PARSED_PROPERTIES="memberOf,owner,manager"
```

## Configuration

Dynamic endpoints are defined in a `config.yaml` file (or equivalent). Example:

```yaml
- path: "/api/v1/groups"
  filter: "(&(objectClass=group)(displayName=部署G_*))"
  search_params: ["cn", "displayName", "memberOf", "department"]
  attributes:
    [
      "cn",
      "distinguishedName",
      "displayName",
      "memberOf",
      "department",
      "info",
      "msDS-PhoneticDisplayName",
      "mail",
      "whenChanged",
      "whenCreated",
    ]
- path: "/api/v1/groups/{cn}"
  filter: "(&(objectClass=group)(CN={cn}))"
  search_params: []
  # All data
  attributes: []
- path: "/api/v1/groups/{cn}/groups"
  filter: "(&(objectClass=group)(memberOf=CN={cn},OU=Groups,OU=JTEKT,DC=jtekt,DC=local))"
  search_params: []
  # All data
  attributes: []
```

## License

This project is licensed under the MIT License.
