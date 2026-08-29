# Offenro

## プロジェクト概要

Offenroは、成果報酬型のエージェントコマースを成立させるためのプロトコルである。Merchant Management APIでMerchant／Capability／Incentiveを登録し、Agentから構造化された検索条件を受け取り、Merchant Live APIから最新Offerを取得して統合する。CartからCheckoutでPurchaseを生成し、Purchase ItemごとのMerchant注文・予約成立後にPurchaseを`CONFIRMED`へ遷移させる。

現在のPrototypeは、次のDomain Extensionを持つ。

- `travel.hotel`
- `retail.shoes`

## 依存ルール

依存方向を次に固定する。

```text
core ← domain ← platform/cmd
```

- `internal/core`にはProtocolロジックだけを置く。
- `internal/core`から`net/http`や具体的なDomainを参照しない。
- `internal/domain/{domain}`はCoreが定義したInterfaceを実装する。
- `internal/platform`にはHTTP Server、Domain API Handler、Protocol API Handler、PostgreSQL、共通HTTP Client等の技術実装を置く。
- `cmd`はComposition RootとしてCore、Domain、Platformを組み立てる。
- Domain追加のためにSearchOffers、Merchant Registry、Discovery Indexer、Index Repositoryを変更しない。
- 具体的なDomainを削除しても`internal/core/...`がビルドできる状態を維持する。

Coreに次のような分岐を書かない。

```go
if domain == "travel.hotel" {
	// ...
}
```

## モデルの正

### Protocol内部モデル

Domain非依存の内部モデルは、次のGoコードを正とする。

```text
internal/core/model/
```

例：Domain、Merchant、MerchantCapability、DiscoveryIndexEntry。

### Domain APIモデル

Agent・Merchantとの通信に使うDomainモデルは、OpenAPIだけを正とする。

```text
api/common/schemas.yaml
api/domains/{domain}/schemas.yaml
api/domains/{domain}/agent.openapi.yaml
api/domains/{domain}/merchant.openapi.yaml
```

手書きGo構造体へDomainフィールドを重複定義しない。ConditionやOfferの手書き型は、Core Interfaceを実装するために生成モデルを包むだけとする。

生成依存は必ず次の一本にする。

```text
schemas.yaml
↓
generated/model
├── generated/agent
└── generated/merchant
```

`generated/agent`と`generated/merchant`は`generated/model`を参照またはaliasし、Domainモデル構造体を重複生成しない。生成されたGoコードは直接編集しない。

### Protocol共通APIモデル

Cart、Checkout、PurchaseのAPI Contractは次を正とする。

```text
api/protocol/schemas.yaml
→ internal/generated/protocol/model/openapi.gen.go

api/protocol/commerce.openapi.yaml
→ internal/generated/protocol/commerce/openapi.gen.go
```

生成コードは直接編集しない。

### Management APIモデル

Merchant、Capability、Incentive、Agentの管理API Contractは次を正とする。

```text
api/management/
→ internal/generated/management/
```

生成コードは直接編集しない。

## 開発者がDomain追加・変更時に触る場所

Domainの主要な手書き実装は、意図的に次の5ファイルへ絞る。

```text
internal/domain/{domain}/
├── generated/       # 自動生成。編集禁止
├── extension.go     # Core登録と薄いModel Adapter
├── searcher.go      # Merchant Live Search／Offer Revalidate
├── fulfiller.go     # Merchant注文・予約作成／結果照会
├── discovery.go     # Merchant選定
└── index_builder.go # CatalogからIndexを生成
```

HTTP transportはDomain実装から分離する。

```text
internal/platform/domainapi/{domain}/
```

コード生成設定もDomain実装から分離する。

```text
internal/codegen/domains/{domain}/
├── generate.go
├── model.yaml
├── agent.yaml
└── merchant.yaml
```

APIフィールドを追加・変更するときは、`extension.go`や生成Goコードではなく、`api/domains/{domain}/schemas.yaml`を編集する。

## Domain追加手順

1. `api/domains/{domain}/`へSchemas、Agent API、Merchant APIを追加する。
2. `internal/codegen/domains/{domain}/`へ生成設定と生成Directiveを追加する。
3. `model`、`agent`、`merchant`パッケージを生成する。
4. `extension.go`、`searcher.go`、`fulfiller.go`、`discovery.go`、`index_builder.go`を実装する。
5. `internal/platform/domainapi/{domain}/`へDomain API Handler Adapterを追加する。
6. `cmd/server/main.go`でDomain ExtensionとDomain API Handlerを登録する。

この作業でProtocol Coreの変更が必要になった場合は、Domain固有概念がCoreへ漏れていないかを先に確認する。

## Searchフロー

```text
Generated Agent Handler
↓
Domain Condition Adapter
↓
Core SearchOffers
↓
Merchant Registry
↓
Domain Merchant Discovery
↓
Domain Searcher
↓
Generated Merchant Client
↓
merchant_offer_refをOpaque offer_idへ変換
↓
Domain Offer Adapter
↓
Agent Response
```

## Merchant Fulfillmentフロー

```text
Protocol Confirm API
↓
Core Fulfillment Use Case
↓
Purchase ItemごとにPENDINGと冪等キーを保存
↓
Domain Fulfiller
↓
Generated Merchant Clientで注文・予約作成
↓
MerchantFulfillment状態を保存
↓
全件成立時だけPurchase CONFIRMED
```

Merchant APIが結果不明の場合は同じ冪等キーで状態照会し、未作成が確認できた場合だけ同じキーで再送する。外部HTTP通信中はDB Transactionを保持しない。

## Discovery Index

CoreはIndexerの実行基盤とGeneric Repositoryを持つ。

```text
internal/core/discovery/
```

Indexの意味は各Domainが持つ。

- `travel.hotel`: `prefecture_code → Merchant`
- `retail.shoes`: 正規化した`brand → Merchant`

Genericな保存形式は次のとおり。

```text
capability_id / dimension / value / supply_count / indexed_at
```

価格や在庫等の動的情報はIndexへ保存せず、検索時にMerchant Live APIから取得する。

## コード生成

全Domainを1コマンドで生成する。

```bash
go generate ./...
```

生成設定はDomain単位で`internal/codegen/domains/{domain}/`に集約する。各Domainの`generate.go`が自身の`model`、`agent`、`merchant`だけを生成する。出力先は次のとおり。

```text
api/common/schemas.yaml
→ internal/generated/common/openapi.gen.go

api/protocol/schemas.yaml
→ internal/generated/protocol/model/openapi.gen.go

api/protocol/commerce.openapi.yaml
→ internal/generated/protocol/commerce/openapi.gen.go

api/management/schemas.yaml + management.openapi.yaml
→ internal/generated/management/

api/domains/{domain}/schemas.yaml
→ internal/domain/{domain}/generated/model/openapi.gen.go

api/domains/{domain}/agent.openapi.yaml
→ internal/domain/{domain}/generated/agent/openapi.gen.go

api/domains/{domain}/merchant.openapi.yaml
→ internal/domain/{domain}/generated/merchant/openapi.gen.go
```

- Agent側はModel、`std-http-server`、`strict-server`を生成する。
- Merchant側はModelとHTTP Clientを生成する。
- Protocol共通APIはModel、`std-http-server`、`strict-server`を生成する。
- 手書きSearcherとFulfillerはGenerated Merchant Clientを使用する。
- OpenAPI変更と生成コードを同じ変更に含める。
- `db/migrations`と`db/queries`からsqlcコードも同じコマンドで生成する。

## ディレクトリ責務

```text
api/                 # OpenAPI Contract。外部モデルの正
cmd/                 # Composition Rootと実行プログラム
db/migrations/       # Protocol DB Schemaの正
db/queries/          # sqlc Queryの正
internal/codegen/    # コード生成Directiveと設定
internal/core/       # Transport非依存のProtocolロジック
internal/domain/     # Domain Extension
internal/generated/  # Domain横断の生成APIモデル
internal/platform/   # HTTP、PostgreSQL等の技術Adapter
internal/testutil/   # 複数packageで共有するテスト専用Fake
test/integration/    # PostgreSQLとHTTPを通す結合テスト
```

## Protocol責務外

次のものはProtocol Coreへ実装しない。

- Merchant自身のAPI Server
- Merchantの商品DB、在庫管理、価格計算
- リアルなHotel空室ロジック
- 大量のMockデータ

OffenroはContract、Merchant接続、Discovery、Offer統合、Merchant注文・予約結果の追跡までを責務とする。

## 技術方針

- HTTP AdapterではGo標準の`net/http`を使用する。
- 論理DBモデルと設計理由は`doc/database-model.md`、実行可能な物理Schemaの正は`db/migrations`とし、両方を同じ変更で更新する。Migrationは`goose`で管理する。
- SQL Queryは`db/queries`へ記述し、`sqlc`で`internal/platform/postgres/generated`へ生成する。
- PostgreSQL接続には`pgx/v5`の`pgxpool`を使用する。
- 手書きRepositoryはsqlc生成メソッドとCoreモデルの変換に限定し、SQLをGoコードへ記述しない。
- `internal/platform/postgres/types.go`は型変換補助に限定し、DomainモデルやDB Row型の第二の正にしない。
- CoreはRepository Interfaceだけを定義し、PostgreSQLの具体技術へ依存しない。
- Merchant Registry、Discovery Index、Cart、Purchase、Merchant FulfillmentはPostgreSQLへ保存する。
- InMemory Adapterは`internal/testutil`へ置き、テスト用途に限定する。
- DB接続情報は`DATABASE_URL`で受け取り、アプリケーション起動時にMigrationを実行しない。
- Merchant HTTPの共通Timeout設定は`internal/platform/httpclient`へ置く。
- Management APIは`MANAGEMENT_API_TOKEN`で保護し、Offer IDは32文字以上の`OFFER_TOKEN_SECRET`で暗号化・認証する。
- Agent向け`offer_id`とMerchant向け`merchant_offer_ref`を分離し、後者をAgent APIへ公開しない。
- Checkout時はMerchant Live APIでOfferを再確認し、Purchaseは`CREATED`、Cartは`CHECKED_OUT`として同一Transactionで保存する。
- `Purchase.CONFIRMED`はMerchant側の注文・予約成立後にのみ設定する。
- 金額は通貨の最小単位を表す`int64`／PostgreSQL `BIGINT`で保持する。割合を含む`reward_value`だけは`NUMERIC(19, 4)`とする。
- 自然言語解釈はAgent側の責務とし、Coreには構造化済みConditionを渡す。
- 必要のない抽象化や機能を先回りして追加しない。

## 開発コマンド

```bash
go generate ./...
go tool sqlc vet
gofmt -w <changed-go-files>
go mod tidy
go test ./...
go vet ./...
```

Server起動：

```bash
go run ./cmd/server
```

Health Check：

```bash
curl http://localhost:8080/health
```

実装変更後は、生成、整形、テスト、静的検査を実行する。
