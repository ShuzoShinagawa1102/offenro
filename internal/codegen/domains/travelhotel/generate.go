// Package travelhotel contains OpenAPI generation settings for travel.hotel.
package travelhotel

//go:generate go tool oapi-codegen -config model.yaml ../../../../api/domains/travel.hotel/schemas.yaml
//go:generate go tool oapi-codegen -config agent.yaml ../../../../api/domains/travel.hotel/agent.openapi.yaml
//go:generate go tool oapi-codegen -config merchant.yaml ../../../../api/domains/travel.hotel/merchant.openapi.yaml
