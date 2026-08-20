# Offenro

変更内容ごとの編集場所は[開発マップ](doc/development-map.md)、新しいDomainの追加手順は[Domain開発標準](doc/domain-development.md)を参照してください。

## コード生成

OpenAPI定義とSQL Queryから、APIコードとDBアクセスコードをまとめて生成します。

```bash
go generate ./...
```

主な生成先：

```text
api/common/schemas.yaml
→ internal/generated/common/openapi.gen.go

api/protocol/schemas.yaml
→ internal/generated/protocol/model/openapi.gen.go

api/protocol/commerce.openapi.yaml
→ internal/generated/protocol/commerce/openapi.gen.go

api/domains/{domain}/schemas.yaml
→ internal/domain/{domain}/generated/model/openapi.gen.go

api/domains/{domain}/agent.openapi.yaml
→ internal/domain/{domain}/generated/agent/openapi.gen.go

api/domains/{domain}/merchant.openapi.yaml
→ internal/domain/{domain}/generated/merchant/openapi.gen.go

db/migrations/ + db/queries/*.sql
→ internal/platform/postgres/generated/
```

生成された`openapi.gen.go`と`internal/platform/postgres/generated/`は直接編集しません。

DBアクセスコードだけを生成する場合：

```bash
go tool sqlc generate
```

SQL Queryを検査する場合：

```bash
go tool sqlc vet
```

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

## ローカルDB

初回のみ環境変数ファイルを作成します。

```bash
cp .env.example .env
```

起動：

```bash
docker compose up -d
```

マイグレーション：

```bash
set -a
source .env
set +a
go tool goose -dir db/migrations postgres "$DATABASE_URL" up
```

状態確認：

```bash
go tool goose -dir db/migrations postgres "$DATABASE_URL" status
```

停止：

```bash
docker compose down
```

詳細は[DB環境](doc/database-environment.md)と[マイグレーション](doc/migrations.md)を参照してください。

## Server起動

`travel.hotel`と`retail.shoes`を登録したOffenro Serverを起動します。

```bash
set -a
source .env
set +a
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
