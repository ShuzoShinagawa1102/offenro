# DB環境

## 環境の使い分け

- ローカル：Docker ComposeのPostgreSQL 16
- 本番：AWS Aurora PostgreSQL
- 接続先の差分：`DATABASE_URL`のみ

ローカル用の設定は`.env.example`を`.env`へコピーして使用します。`.env`はGit管理対象外です。

```bash
cp .env.example .env
```

本番の接続情報はリポジトリや環境ファイルへ保存せず、AWS Secrets Managerから実行環境へ`DATABASE_URL`として注入します。

## ローカルDB

起動：

```bash
docker compose up -d
```

接続確認：

```bash
docker compose exec postgres psql -U offenro -d offenro
```

接続情報：

```text
postgres://offenro:offenro@localhost:5432/offenro?sslmode=disable
```

停止：

```bash
docker compose down
```

停止してもデータはDocker Volumeへ保持されます。

## マイグレーション

アプリケーション起動時にはマイグレーションを実行しません。ローカルでは開発者が明示的に実行し、本番ではデプロイ工程の独立したステップとして実行します。手順は[マイグレーション](migrations.md)を参照してください。
