# Domain開発標準

## 目的

新しいDomainを、Protocol Coreを変更せずに追加するための標準手順を定める。

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

internal/codegen/{go-domain}/
├── generate.go
├── model.yaml
├── agent.yaml
└── merchant.yaml

internal/domain/{go-domain}/
├── extension.go
├── searcher.go
├── discovery.go
└── index_builder.go

internal/platform/agentapi/{go-domain}/
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
```

これらはOpenAPIから生成する。変更が必要な場合は、元のOpenAPIまたはCodegen設定を修正して再生成する。

### Domain追加では原則変更しない

```text
internal/core/
internal/platform/httpclient/
internal/platform/httpserver/
api/common/schemas.yaml
```

- `internal/core`はDomain非依存のProtocolロジックである。
- 共通HTTP実装は、全Domainに影響する技術要件を変更するときだけ編集する。
- `api/common/schemas.yaml`は、MoneyやError等のDomain横断モデルを追加するときだけ編集する。

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
internal/codegen/{go-domain}/
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

### 4. Agent API Handlerを追加する

```text
internal/platform/agentapi/{go-domain}/handler.go
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

// Agent API Handler
agentAPIs := []httpserver.RouteMounter{
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

## 完了確認

```bash
go generate ./...
go mod tidy
go test ./...
go vet ./...
```

次も確認する。

- 生成コードを直接編集していない。
- `generated/agent`と`generated/merchant`にDomainモデル構造体が重複していない。
- `internal/core`に具体的なDomain名やHTTP依存を追加していない。
- 新Domain追加が、そのDomainの追加とComposition Rootへの登録だけで完結している。
