# Offenro

## コード生成

`api/`配下のOpenAPI定義をもとに、Domainモデル、Agent API Handler、Merchant API Clientを生成します。

```bash
go generate ./...
```

主な生成先：

```text
api/common/schemas.yaml
→ internal/generated/common/openapi.gen.go

api/domains/{domain}/schemas.yaml
→ internal/domain/{domain}/generated/model/openapi.gen.go

api/domains/{domain}/agent.openapi.yaml
→ internal/domain/{domain}/generated/agent/openapi.gen.go

api/domains/{domain}/merchant.openapi.yaml
→ internal/domain/{domain}/generated/merchant/openapi.gen.go
```

生成された`openapi.gen.go`は直接編集しません。

## Module整理

不要な依存を削除し、`go.mod`と`go.sum`を整理します。

```bash
go mod tidy
```

## テスト

全パッケージのテストとコンパイルを実行します。

```bash
go test ./...
```

## 静的検査

Goコードの静的な問題を検査します。

```bash
go vet ./...
```

## Server起動

`travel.hotel`と`retail.shoes`を登録したOffenro Serverを起動します。

```bash
go run ./cmd/server
```

起動先は`http://localhost:8080`です。

## Health Check

```bash
curl http://localhost:8080/health
```

## Hotel検索

`travel.hotel`のAgent APIを呼び出します。

```bash
curl -X POST http://localhost:8080/v1/travel/hotels/search \
  -H "Content-Type: application/json" \
  -d '{
    "destination": {"prefecture_code": "14"},
    "stay": {
      "check_in": "2026-09-10",
      "check_out": "2026-09-12"
    },
    "guests": {"adults": 2, "rooms": 1},
    "filters": {"max_price": 60000}
  }'
```

## Shoes検索

`retail.shoes`のAgent APIを呼び出します。

```bash
curl -X POST http://localhost:8080/v1/retail/shoes/search \
  -H "Content-Type: application/json" \
  -d '{
    "criteria": {
      "brand": "example-brand",
      "size": "27.0",
      "category": "sneakers"
    },
    "quantity": 1,
    "filters": {"max_price": 30000}
  }'
```
