# Commerce／DB開発標準

## 設計単位

Repository Interfaceはテーブル単位の汎用CRUDにしない。CoreのUse CaseとAggregateが必要とする操作を定義する。

```text
HTTP Handler (platform)
        ↓
Use Case / Service (core)
        ↓
Repository Interface (core)
        ↓
PostgreSQL Adapter (platform/postgres)
        ↓
sqlc生成コード
        ↓
db/queries/*.sql
```

- Merchant、Capability、Incentiveの登録・状態遷移は`merchant.ManagementUseCases`が担当する。
- Cart操作、Offer再確認、Purchase生成は`cart.Service`が担当する。
- Merchant注文・予約の冪等性、状態追跡、Purchase状態集約は`fulfillment.Service`が担当する。
- `CartStore`と`PurchaseReader`は利用目的で分け、巨大なテーブルCRUD Interfaceにしない。
- 複数AggregateをまとめるDashboardは、必要になった時点で専用Read Modelと専用SQL Queryを追加する。
- `getPurchase().getParent()`のような暗黙Lazy Loadは使わない。必要な関連はUse Caseが明示的なRepository操作で取得する。

## APIと実装場所

| API | Contract | HTTP Adapter | Core |
|---|---|---|---|
| Agent Protocol API | `api/protocol` | `internal/platform/protocolapi` | `internal/core/cart`等 |
| Agent Domain API | `api/domains/{domain}/agent.openapi.yaml` | `internal/platform/domainapi/{domain}` | `internal/core/search`＋`internal/domain/{domain}` |
| Management API | `api/management` | `internal/platform/managementapi` | `internal/core/merchant` |
| Merchant Live API | `api/domains/{domain}/merchant.openapi.yaml` | Generated ClientをDomainが利用 | `internal/domain/{domain}` |

Management APIが提供する現在の操作はMerchant、Capability、IncentiveRule、Agentの登録・取得・更新、Domain一覧、Capability検証である。物理削除ではなくStatusでLifecycleを管理する。

## SQL追加手順

```text
1. db/migrations/*.sqlへSchema変更を書く
2. db/queries/*.sqlへ名前付きSQLを書く
3. go generate ./...
4. internal/platform/postgresでCore型との変換を追加する
5. go tool sqlc vet && go test ./...
```

論理モデルと設計理由は`doc/database-model.md`、Migration SQLは物理DB Schemaの正、`db/queries`はQueryの正である。Schema変更ではこれらとCoreモデルを同じ変更で更新する。`internal/platform/postgres/generated`は直接編集しない。

金額は通貨の最小単位を表す整数として、OpenAPI／Coreは`int64`、PostgreSQLは`BIGINT`で統一する。割合を表す`incentive_rule.reward_value`だけは`NUMERIC(19, 4)`を使用する。

## Checkoutの整合性境界

```text
Cart ACTIVE
→ offer_idを検証してmerchant_offer_refを解決
→ Active Capabilityを取得
→ Merchant Live APIでOffer Revalidate
→ 同一価格・通貨か確認
→ 1 TransactionでPurchase CREATEDとCart CHECKED_OUTを保存
```

失効・在庫切れ・価格変更は409、Merchant通信障害は502とし、どちらもCartを`ACTIVE`のまま残す。`Purchase.CONFIRMED`はMerchant側の注文・予約成立後にだけ設定する。

## Merchant Fulfillmentの整合性境界

```text
Purchase CREATED
→ Purchase ItemごとにMerchantFulfillment PENDINGと冪等キーをTransactionで保存
→ Transaction終了
→ Domain FulfillerからMerchant APIを呼ぶ
→ Itemごとの結果をCONFIRMED／REJECTED／UNKNOWNとして保存
→ 全件CONFIRMEDの場合だけPurchase CONFIRMED
```

Merchant Capabilityは`purchase_item.capability_id`から導出し、`merchant_fulfillment`へ重複保存しない。結果不明時は同じ冪等キーでMerchantへ照会してから再送する。
