# Offenro

## プロジェクト概要

Offenroは、成果報酬型のエージェントコマースを成立させるためのプロトコルである。Agentから構造化された検索条件を受け取り、候補Merchantを選び、Merchant Live APIから最新Offerを取得して統合する。

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
- `internal/platform`にはHTTP Server、Agent API Handler、共通HTTP Client等の技術実装を置く。
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

## 開発者がDomain追加・変更時に触る場所

Domainの主要な手書き実装は、意図的に次の4ファイルへ絞る。

```text
internal/domain/{domain}/
├── generated/       # 自動生成。編集禁止
├── extension.go     # Core登録と薄いModel Adapter
├── searcher.go      # Merchant Live Search
├── discovery.go     # Merchant選定
└── index_builder.go # CatalogからIndexを生成
```

HTTP transportはDomain実装から分離する。

```text
internal/platform/agentapi/{domain}/
```

コード生成設定もDomain実装から分離する。

```text
internal/codegen/
```

APIフィールドを追加・変更するときは、`extension.go`や生成Goコードではなく、`api/domains/{domain}/schemas.yaml`を編集する。

## Domain追加手順

1. `api/domains/{domain}/`へSchemas、Agent API、Merchant APIを追加する。
2. `internal/codegen/`へ生成設定と生成Directiveを追加する。
3. `model`、`agent`、`merchant`パッケージを生成する。
4. `extension.go`、`searcher.go`、`discovery.go`、`index_builder.go`を実装する。
5. `internal/platform/agentapi/{domain}/`へAgent Handler Adapterを追加する。
6. `cmd/server/main.go`でDomain ExtensionとAgent Handlerを登録する。

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
Domain Offer Adapter
↓
Agent Response
```

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
domain / dimension / value / merchant_id / supply_count / indexed_at
```

価格や在庫等の動的情報はIndexへ保存せず、検索時にMerchant Live APIから取得する。

## OpenAPIコード生成

全Domainを1コマンドで生成する。

```bash
go generate ./...
```

生成設定は`internal/codegen/`に集約する。出力先は次のとおり。

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

- Agent側はModel、`std-http-server`、`strict-server`を生成する。
- Merchant側はModelとHTTP Clientを生成する。
- 手書きSearcherはGenerated Merchant Clientを使用する。
- OpenAPI変更と生成コードを同じ変更に含める。

## ディレクトリ責務

```text
api/                 # OpenAPI Contract。外部モデルの正
cmd/                 # Composition Rootと実行プログラム
internal/codegen/    # コード生成Directiveと設定
internal/core/       # Transport非依存のProtocolロジック
internal/domain/     # Domain Extension
internal/generated/  # Domain横断の生成APIモデル
internal/platform/   # HTTP等の技術Adapter
```

## Protocol責務外

次のものはProtocol Coreへ実装しない。

- Merchant自身のAPI Server
- Merchantの商品DB、在庫管理、価格計算
- リアルなHotel空室ロジック
- 大量のMockデータ

OffenroはContract、Merchant接続、Discovery、Offer統合までを責務とする。

## 技術方針

- HTTP AdapterではGo標準の`net/http`を使用する。
- 必要になるまでWeb FrameworkやDBを導入しない。
- PrototypeのMerchant RegistryとDiscovery IndexはInMemoryとする。
- Merchant HTTPの共通Timeout設定は`internal/platform/httpclient`へ置く。
- 自然言語解釈はAgent側の責務とし、Coreには構造化済みConditionを渡す。
- 必要のない抽象化や機能を先回りして追加しない。

## 開発コマンド

```bash
go generate ./...
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
