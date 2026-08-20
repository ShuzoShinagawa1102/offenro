// Package protocol contains OpenAPI generation settings for the Protocol API.
package protocol

//go:generate go tool oapi-codegen -config model.yaml ../../../api/protocol/schemas.yaml
//go:generate go tool oapi-codegen -config commerce.yaml ../../../api/protocol/commerce.openapi.yaml
