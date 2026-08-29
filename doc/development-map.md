# 開発マップ

## Repositoryと実装Adapter

Repositoryは、Coreが必要とするデータ操作を表すinterfaceである。保存技術はCoreへ持ち込まない。

```text
internal/core                         # Repositoryの契約（Port）
├── cart.CartStore / PurchaseReader
├── discovery.IndexRepository
├── fulfillment.Repository
├── merchant.Registry / ManagementRepository
└── offer.Codec / VerifierRegistry
          │
          ├── internal/platform/postgres # 本番用PostgreSQL Adapter
          └── internal/testutil           # テスト用Memory Adapter
```

- 本番Serverは`internal/platform/postgres`を使用する。
- `internal/testutil`はテストからだけ使用し、本番コードからimportしない。
- 新しいRepository interfaceは、それを利用するCore packageに置く。
- 保存方式ごとのinterfaceを`internal/repository`等へ集約しない。

## 変更内容から編集場所を選ぶ

| 変更内容 | 最初に編集する場所 | 必要に応じて編集する場所 | 直接編集しない場所 |
|---|---|---|---|
| Domainモデル・API項目 | `api/domains/{api-domain}` | `internal/domain/{go-domain}`、`internal/platform/domainapi/{go-domain}` | `generated/` |
| Domain検索・Merchant選定 | `internal/domain/{go-domain}` | Domainのテスト | `internal/core` |
| 新Domain追加 | `api/domains`、`internal/codegen/domains`、`internal/domain` | `internal/platform/domainapi`、`cmd/server/main.go` | 既存Domainの生成物 |
| Cart・Purchase API | `api/protocol` | `internal/core/cart`、`internal/platform/protocolapi` | `internal/generated/protocol` |
| Merchant注文・予約 | `api/protocol`、`api/domains/{api-domain}` | `internal/core/fulfillment`、Domainの`fulfiller.go`、`internal/platform/protocolapi` | 生成コード |
| Merchant管理API | `api/management` | `internal/core/merchant`、`internal/platform/managementapi` | `internal/generated/management` |
| Offer ID／再確認 | `api/domains/{api-domain}` | `internal/core/offer`、Domainの`searcher.go`、`internal/platform/offertoken` | Domain生成物 |
| Coreユースケース | `internal/core/{feature}` | 対応するPlatform Adapterとテスト | Domain生成モデル |
| DB Schema | `doc/database-model.md`、`db/migrations` | `db/queries`、Coreモデル、PostgreSQL Adapter | sqlc生成物 |
| SQL Query | `db/queries` | `internal/platform/postgres`の変換処理 | Goコード内のSQL、sqlc生成物 |
| 共有テストFake | `internal/testutil` | 利用側の`*_test.go` | 本番Composition Root |

## ファイルを分ける基準

- Coreは`cart`、`merchant`、`offer`、`search`、`discovery`等のユースケース・Protocol概念単位でpackageを分ける。
- Domainの手書き実装は`extension.go`、`searcher.go`、`fulfiller.go`、`discovery.go`、`index_builder.go`を基本単位とする。
- HTTP、PostgreSQL等の技術実装は`internal/platform`へ置く。
- Unit Testは対象コードと同じpackageの`*_test.go`へ置く。
- 複数packageで共有するテスト実装だけを`internal/testutil`へ置く。
- 外部サービスやDBを起動する結合テストが増えた場合は、`test/integration`を追加する。

## 生成と確認

```bash
go generate ./...
go tool sqlc vet
go test ./...
go vet ./...
```

Domain追加の詳細は[Domain開発標準](domain-development.md)、Commerce／DB実装は[Commerce／DB開発標準](commerce-development.md)、DB変更の詳細は[マイグレーション](migrations.md)を参照する。
