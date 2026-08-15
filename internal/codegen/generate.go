// Package codegen centralizes OpenAPI generation for all domains.
package codegen

//go:generate go tool oapi-codegen -config common.yaml ../../api/common/schemas.yaml

// 新しいDomainを追加するときは、model・agent・merchantの3つの生成Directiveを追加する。
//go:generate go tool oapi-codegen -config travelhotel.model.yaml ../../api/domains/travel.hotel/schemas.yaml
//go:generate go tool oapi-codegen -config travelhotel.agent.yaml ../../api/domains/travel.hotel/agent.openapi.yaml
//go:generate go tool oapi-codegen -config travelhotel.merchant.yaml ../../api/domains/travel.hotel/merchant.openapi.yaml
//go:generate go tool oapi-codegen -config retailshoes.model.yaml ../../api/domains/retail.shoes/schemas.yaml
//go:generate go tool oapi-codegen -config retailshoes.agent.yaml ../../api/domains/retail.shoes/agent.openapi.yaml
//go:generate go tool oapi-codegen -config retailshoes.merchant.yaml ../../api/domains/retail.shoes/merchant.openapi.yaml
