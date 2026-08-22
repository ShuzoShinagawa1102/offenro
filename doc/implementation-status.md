# 実装状況

## 現在の実装状況

| 機能 | DB Schema | sqlc／Repository | Core | Web API |
|---|---:|---:|---:|---:|
| Merchant登録・取得・更新 | あり | あり | あり | Management APIあり |
| MerchantCapability登録・取得・更新 | あり | あり | あり | Management APIあり |
| Capability検証・有効化 | あり | Transactionあり | あり | Management APIあり |
| IncentiveRule登録・取得・更新 | あり | あり | あり | Management APIあり |
| CommerceDomain一覧 | あり | あり | あり | Management APIあり |
| Agent登録・取得 | あり | あり | あり | Management APIあり |
| Search | Indexあり | あり | あり | あり |
| Cart／CartItem | あり | あり | あり | あり |
| Checkout | あり | Transactionあり | Revalidateあり | あり |
| Purchase／PurchaseItem | あり | 生成・取得あり | あり | 取得あり |
| Checkout時Offer再確認 | 参照値を保存 | あり | あり | Merchant Contractあり |
| Payment／Stripe | Schemaのみ | 未実装 | 未実装 | 未実装 |

Management APIはPrototype用Bearer Tokenで保護する。Agent APIの認証、Merchantごとの認証情報管理、管理画面は未実装である。

## 完了済みの境界

```text
Merchant登録
→ Capability／API URL登録
→ Incentive登録
→ Merchant Live API検証
→ Search対象として有効化

Search
→ Cart ACTIVE
→ Offer Revalidate
→ Purchase CREATED
→ Cart CHECKED_OUT
```

`Purchase.CONFIRMED`、Payment処理、Stripe接続、Dashboardは今回の実装対象外とする。

Checkoutでは検索時の`offer_id`から内部の`merchant_offer_ref`を解決し、Merchant Live APIへ再確認する。Offer失効、価格・通貨変更時は409を返し、Cartは`ACTIVE`のまま維持する。外部通信障害は502とする。

## 次の実装単位

- Merchant側の注文・予約作成と`Purchase.CONFIRMED`遷移
- Payment Use Case、Stripe PaymentIntent、署名検証済みWebhook
- Management APIのMerchant単位認可とSecret管理
- Dashboard向けRead Model／専用Query
