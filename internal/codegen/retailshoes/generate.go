// Package retailshoes contains OpenAPI generation settings for retail.shoes.
package retailshoes

//go:generate go tool oapi-codegen -config model.yaml ../../../api/domains/retail.shoes/schemas.yaml
//go:generate go tool oapi-codegen -config agent.yaml ../../../api/domains/retail.shoes/agent.openapi.yaml
//go:generate go tool oapi-codegen -config merchant.yaml ../../../api/domains/retail.shoes/merchant.openapi.yaml
