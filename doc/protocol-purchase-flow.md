# Protocol Purchase Flow

```text
Search → Cart → Checkout → Purchase
```

- `Search`はMerchantの最新Offerを検索する。検索候補はProtocolへ永続化しない。
- `Cart`はUserが購入候補として明示したOfferを、Snapshotと共通の金額・通貨とともに保持する。
- `Checkout`は有効なCart Itemを検証し、Cartを`CHECKED_OUT`へ変更してPurchaseを同一トランザクションで生成する。
- `Purchase`はCheckout後の取引Snapshotである。Purchaseの生成経路はCheckoutだけとし、`POST /v1/purchases`は設けない。

CartとPurchaseは特定のCommerce Domainに属さないProtocol共通機能である。各ItemがMerchantとCommerce Domainを参照し、Domain固有のOffer内容はSnapshotとして保持する。

Cart作成には登録済みの有効なAgent、Item追加には有効なMerchant Capabilityが必要となる。

## Source of Truth

- Protocolモデル：`api/protocol/schemas.yaml`
- Request、Response、Path等のAPI仕様：`api/protocol/commerce.openapi.yaml`
- DB Schema：`db/migrations/`
- DB Query：`db/queries/`

MarkdownへAPI項目を転記して二重管理しない。

## DBアクセスの責務

`internal/core/cart`は入力検証、Cartの状態遷移、Checkout時のPurchase生成を担当し、DB技術へ依存しない。

`internal/platform/postgres`はCoreのRepository Interfaceを実装し、sqlc生成メソッドによるQuery実行、トランザクション、`DATABASE_URL`による`pgxpool`接続を担当する。SQLは`db/queries`へ記述し、アプリケーション起動時にMigrationは実行しない。
