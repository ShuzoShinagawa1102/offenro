# Domain開発標準

## 目的

新しいDomainを、Protocol Coreを変更せずに追加するための標準手順を定める。

変更内容から編集場所を探す場合は、先に[開発マップ](development-map.md)を参照する。

依存方向は次に固定する。

```text
core ← domain ← platform/cmd
```

Domain追加のために`internal/core`を変更する必要がある場合は、Domain固有概念がCoreへ漏れていないかを先に確認する。

## 編集場所

### 開発者が編集する

```text
api/domains/{api-domain}/
├── schemas.yaml
├── agent.openapi.yaml
└── merchant.openapi.yaml

internal/codegen/domains/{go-domain}/
├── generate.go
├── model.yaml
├── agent.yaml
└── merchant.yaml

internal/domain/{go-domain}/
├── extension.go
├── searcher.go
├── discovery.go
└── index_builder.go

internal/platform/domainapi/{go-domain}/
└── handler.go

cmd/server/main.go
```

- `{api-domain}`は`travel.hotel`、`retail.shoes`のようなProtocol上のDomain名とする。
- `{go-domain}`は`travelhotel`、`retailshoes`のようなGo package名とする。

### 直接編集しない

```text
internal/domain/{go-domain}/generated/
├── model/openapi.gen.go
├── agent/openapi.gen.go
└── merchant/openapi.gen.go

internal/generated/common/openapi.gen.go
internal/generated/protocol/
internal/platform/postgres/generated/
```

これらはOpenAPIまたはSQLから生成する。変更が必要な場合は、生成元のOpenAPI、Migration、Query、Codegen設定を修正して再生成する。

### Domain追加では原則変更しない

```text
internal/core/
internal/platform/httpclient/
internal/platform/httpserver/
api/common/schemas.yaml
api/protocol/
internal/platform/protocolapi/
```

- `internal/core`はDomain非依存のProtocolロジックである。
- 共通HTTP実装は、全Domainに影響する技術要件を変更するときだけ編集する。
- `api/common/schemas.yaml`は、MoneyやError等のDomain横断モデルを追加するときだけ編集する。
- `api/protocol`と`internal/platform/protocolapi`は、CartやPurchase等のProtocol共通APIを変更するときだけ編集する。

API ContractとHTTP Adapterは同じ分類名で対応させる。

```text
api/protocol/         → internal/platform/protocolapi/
api/domains/{domain}/ → internal/platform/domainapi/{go-domain}/
```

## モデル生成ルール

Domainモデルの正は`api/domains/{api-domain}/schemas.yaml`とする。

生成依存は必ず次の一本にする。

```text
schemas.yaml
↓
generated/model
├── generated/agent
└── generated/merchant
```

- `generated/model`だけがDomainモデル構造体を持つ。
- `generated/agent`と`generated/merchant`は`generated/model`を参照する。
- 手書きGo構造体へDomainフィールドを再定義しない。
- `extension.go`のConditionやOfferは、生成モデルを包んでCore Interfaceを実装するだけとする。

## 新Domain追加手順

### 1. OpenAPI Contractを定義する

```text
api/domains/{api-domain}/schemas.yaml
api/domains/{api-domain}/agent.openapi.yaml
api/domains/{api-domain}/merchant.openapi.yaml
```

- `schemas.yaml`: Domainモデルの正
- `agent.openapi.yaml`: AgentからOffenroへ公開するAPI
- `merchant.openapi.yaml`: OffenroからMerchantへ接続するAPI

Agent向けOfferにだけ必要な`merchant_id`や`domain`は、Merchant向けOfferへ要求しない。Domain Adapterで付与する。

### 2. Codegen設定を追加する

既存Domainのフォルダを参考に、次を追加する。

```text
internal/codegen/domains/{go-domain}/
├── generate.go
├── model.yaml
├── agent.yaml
└── merchant.yaml
```

`generate.go`は自身のDomainのModel、Agent、Merchantだけを生成する。

```bash
go generate ./...
```

生成後、AgentとMerchantの生成コードが`generated/model`を参照していることを確認する。

### 3. Domain Extensionを実装する

```text
internal/domain/{go-domain}/
├── extension.go
├── searcher.go
├── discovery.go
└── index_builder.go
```

#### extension.go

- Domain識別子を定義する。
- 生成Requestを包むCondition Adapterを実装する。
- Domain固有Validationを実装する。
- 生成Offerを包むOffer Adapterを実装する。
- Searcher、Discovery、IndexBuilderをExtensionとしてまとめる。

#### searcher.go

- Generated Merchant Clientを利用する。
- ConditionをMerchant Requestへ渡す。
- Merchant OfferをAgent向けOfferへ変換する。
- `merchant_id`と`domain`をOffenro側で付与する。

手書きでHTTP RequestやJSONモデルを再実装しない。

#### discovery.go

- 検索条件からDiscovery Indexの検索軸を決定する。
- Index結果とMerchant候補を照合する。
- CoreのGeneric Repositoryを利用する。

#### index_builder.go

- Generated Merchant ClientでCatalogを取得する。
- Domain固有の軸でCatalogを集計する。
- Genericな`DiscoveryIndexEntry`へ変換する。

価格や在庫等の動的情報はIndexへ保存しない。

### 4. Domain API Handlerを追加する

```text
internal/platform/domainapi/{go-domain}/handler.go
```

- Generated Agent Handler Interfaceを実装する。
- Agent RequestをDomain Condition Adapterへ渡す。
- Coreから返されたOfferをGenerated Agent Responseへ変換する。
- 共通HTTP ServerへMountできるようにする。

### 5. Composition Rootへ登録する

`cmd/server/main.go`の次の2か所へ追加する。

```go
// Domain Extension
domainExtensions := []extension.Extension{
	// 新Domainを追加
}

// Domain API Handler
routeMounters := []httpserver.RouteMounter{
	// 新Domainを追加
}
```

Domain追加時にCore側へ登録分岐を追加しない。

### 6. テストを追加する

最低限、次を確認する。

- Condition Validation
- Agent APIが共通HTTP ServerへMountされること
- Merchantが未登録でも空のOffer一覧を返せること
- SearcherがGenerated Merchant Clientを利用していること
- DiscoveryとIndexBuilderが同じDimension／Value形式を使うこと
- 複数packageで共有するFake Repositoryは`internal/testutil`を利用すること

## 既存Domain変更手順

APIフィールドの追加・変更では、通常は次の手順だけでよい。

```text
schemas.yamlを変更
↓
必要ならagent.openapi.yaml／merchant.openapi.yamlを変更
↓
go generate ./...
↓
Domain Adapterとテストを更新
```

既存Domainの通常変更では、Codegen設定や`cmd/server/main.go`を変更しない。

## DB変更手順

DB Schemaの正はMigration SQL、Queryの正は`db/queries`のSQLとする。GoコードへSQLを書かない。

```text
db/migrations/                         # goose Migration。Schemaの正
db/queries/                            # sqlc Query。開発者がSQLを記述する
sqlc.yaml                              # 全Query共通の生成設定
internal/platform/postgres/generated/ # sqlc生成物。直接編集しない
internal/platform/postgres/            # Coreとの薄い変換Adapter
```

DB SchemaまたはQueryを追加・変更するときは、次の手順に固定する。

1. Schema変更がある場合は`db/migrations`へgoose Migrationを追加する。
2. `db/queries/*.sql`へ`-- name: QueryName :one`等のsqlc Queryを記述する。
3. `go generate ./...`でDBアクセスコードを生成する。
4. `internal/platform/postgres`のRepositoryから生成メソッドを呼び、Coreモデルとの変換だけを実装する。
5. Migration、Query、Repositoryのテストを実行する。

```bash
go tool goose -dir db/migrations create <変更名> sql
go generate ./...
go tool sqlc vet
go test ./...
```

`internal/platform/postgres/types.go`は`pgtype`とCore型の変換補助だけに使う。DomainモデルやDB Row構造体を定義せず、第二のSource of Truthにしない。

## 完了確認

```bash
go generate ./...
go tool sqlc vet
go mod tidy
go test ./...
go vet ./...
```

次も確認する。

- 生成コードを直接編集していない。
- `generated/agent`と`generated/merchant`にDomainモデル構造体が重複していない。
- `internal/core`に具体的なDomain名やHTTP依存を追加していない。
- 新Domain追加が、そのDomainの追加とComposition Rootへの登録だけで完結している。
- SQLをGoコードへ直接記述していない。
- `internal/platform/postgres/generated`を直接編集していない。
