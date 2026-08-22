// Package management contains OpenAPI generation settings for Management APIs.
package management

//go:generate go tool oapi-codegen -config model.yaml ../../../api/management/schemas.yaml
//go:generate go tool oapi-codegen -config api.yaml ../../../api/management/management.openapi.yaml
