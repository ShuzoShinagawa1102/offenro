# マイグレーション

## 管理ルール

- ツールは`goose`を使用する
- SQLは`db/migrations`へ全ドメイン共通の時系列で置く
- 適用済みのSQLは変更せず、変更内容を新しいSQLとして追加する
- 各SQLに`-- +goose Up`と`-- +goose Down`を記述する
- スキーマ設計は[Database Model](database-model.md)を参照する
- Schema変更ではDatabase Model、Migration、Query、Coreモデルを同じ変更で更新する
- `db/migrations`をsqlcのSchema入力としても使用する

新規SQLの作成：

```bash
go tool goose -dir db/migrations create <変更名> sql
```

## ローカルでの実行

Git Bashの場合：

```bash
set -a
source .env
set +a
go tool goose -dir db/migrations postgres "$DATABASE_URL" up
go tool goose -dir db/migrations postgres "$DATABASE_URL" down
go tool goose -dir db/migrations postgres "$DATABASE_URL" status
```

PowerShellの場合：

```powershell
$env:DATABASE_URL = 'postgres://offenro:offenro@localhost:5432/offenro?sslmode=disable'
go tool goose -dir db/migrations postgres "$env:DATABASE_URL" up
go tool goose -dir db/migrations postgres "$env:DATABASE_URL" down
go tool goose -dir db/migrations postgres "$env:DATABASE_URL" status
```

`up`は未適用分の適用、`down`は1件の取り消し、`status`は適用状態の確認です。

Schema変更後は、`db/queries`のSQLからDBアクセスコードを再生成して検査します。

```bash
go tool sqlc generate
go tool sqlc vet
```

## 本番での実行

デプロイ工程でAWS Secrets Managerから`DATABASE_URL`を取得し、アプリケーションの切り替え前に次を1回実行します。

```bash
go tool goose -dir db/migrations postgres "$DATABASE_URL" up
```

アプリケーション起動処理からは実行しません。本番での`down`は自動実行せず、影響を確認したうえで個別に判断します。
