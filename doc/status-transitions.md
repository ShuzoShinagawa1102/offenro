# 購入フローStatus管理表

## 識別子と処理単位

現在、購入フロー全体を表す専用の`Session`モデルは存在しない。

| 識別子 | 役割 |
|---|---|
| `agent_id` | Offenroを利用するAgent |
| `buyer_ref` | Agent側Userを表すOpaqueな参照値 |
| `offer_id` | Search結果からMerchant Offerを安全に再特定するOpaque ID |
| `cart_id` | Cart作成後からCheckoutまでの中心識別子 |
| `purchase_id` | Checkout後の購入処理を追跡する中心識別子 |

Searchはステートレスである。SearchからCart追加までは`offer_id`、Cart作成後は`cart_id`、Checkout後は`purchase_id`で処理を相関させる。

## Status管理表

`—`はレコード未生成、`複数`はPurchase Itemごとに異なる状態を持ち得ることを表す。

| No | イベント | Cart | CartItem | Purchase | MerchantFulfillment | Payment | 実装段階 |
|---:|---|---|---|---|---|---|---|
| 1 | Offer検索 | — | — | — | — | — | 現行 |
| 2 | Cart作成 | `ACTIVE` | — | — | — | — | 現行 |
| 3 | OfferをCartへ追加 | `ACTIVE` | `ACTIVE` | — | — | — | 現行 |
| 4 | Cart Item削除 | `ACTIVE` | `REMOVED` | — | — | — | 現行 |
| 5 | Checkout時のOffer再確認失敗 | `ACTIVE` | `ACTIVE` | — | — | — | 現行 |
| 6 | Checkout成功 | `CHECKED_OUT` | `ACTIVE` | `CREATED` | — | — | 現行 |
| 7 | Merchant注文・予約の送信準備 | `CHECKED_OUT` | `ACTIVE` | `CREATED` | `PENDING` | — | 現行 |
| 8 | Merchantへ送信したが結果不明 | `CHECKED_OUT` | `ACTIVE` | `CREATED` | `UNKNOWN` | — | 現行 |
| 9 | 全Merchant注文・予約が成立 | `CHECKED_OUT` | `ACTIVE` | `CONFIRMED` | `CONFIRMED` | — | 現行 |
| 10 | 全Merchantが拒否し不成立が確定 | `CHECKED_OUT` | `ACTIVE` | `CANCELLED` | `REJECTED` | — | 現行 |
| 11 | 一部成功・一部失敗 | `CHECKED_OUT` | `ACTIVE` | `CREATED` | 複数 | — | 現行 |
| 12 | Stripe決済開始 | `CHECKED_OUT` | `ACTIVE` | `CREATED` | —または`PENDING` | `PENDING` | 将来 |
| 13 | 支払手段の承認 | `CHECKED_OUT` | `ACTIVE` | `CREATED` | —または`PENDING` | `AUTHORIZED` | 将来 |
| 14 | 決済確定 | `CHECKED_OUT` | `ACTIVE` | Purchase状態とは独立 | 状態を維持 | `CAPTURED` | 将来 |
| 15 | 決済失敗 | `CHECKED_OUT` | `ACTIVE` | `CREATED`または`CANCELLED` | 状態を維持 | `FAILED` | 将来 |

## Statusの責務

### Cart

```text
ACTIVE
├─ CHECKED_OUT
├─ ABANDONED
└─ EXPIRED
```

Cartは購入候補の編集状態だけを表す。`CHECKED_OUT`はPurchaseが生成済みという意味であり、Merchant注文成立や決済成功を意味しない。

### Purchase

```text
CREATED
├─ CONFIRMED
└─ CANCELLED
```

- `CREATED`: Purchase生成済み。Merchant側の注文・予約は未成立または処理途中。
- `CONFIRMED`: 全Purchase ItemについてMerchant側の注文・予約が成立。
- `CANCELLED`: 全Purchase ItemについてMerchantが拒否し、購入不成立が確定。

一部のMerchantだけが成功した状態を`CONFIRMED`または`CANCELLED`へ丸めない。Purchaseを`CREATED`のまま保持し、ItemごとのMerchantFulfillment状態で判別する。

### MerchantFulfillment

MerchantFulfillmentは、Purchase ItemをMerchant側で実際の注文・予約へ変換するProtocol共通モデルである。

```text
PENDING
├─ CONFIRMED
├─ REJECTED
└─ UNKNOWN
    ├─ PENDING
    ├─ CONFIRMED
    └─ REJECTED
```

| Status | 意味 |
|---|---|
| `PENDING` | Merchant送信前または処理中 |
| `CONFIRMED` | Merchant側の注文・予約成立 |
| `REJECTED` | 在庫、空室、入力条件等により不成立 |
| `UNKNOWN` | Timeout等によりMerchant側の成立有無を断定できない |

`UNKNOWN`を単純な失敗として再送すると二重予約・二重注文になり得るため、Merchant APIには冪等キーと状態照会手段を要求する。

### Payment（将来実装）

```text
PENDING
├─ AUTHORIZED
│  └─ CAPTURED
├─ CAPTURED
├─ FAILED
└─ CANCELLED
```

Payment状態はPurchase状態と統合しない。Stripe接続時に、決済をMerchant注文の前後どちらで確定するかをCommerce Policyとして決定する。

## 汎用部分とDomain固有部分

| 汎用部分（Core／Protocol） | Domain固有部分 |
|---|---|
| PurchaseからFulfillmentを生成する | Hotel予約Requestを組み立てる |
| ItemごとのStatus、冪等性、再試行を管理する | Shoes注文Requestを組み立てる |
| 全Fulfillment成功時だけPurchaseを`CONFIRMED`にする | Generated Merchant Clientで各APIを呼ぶ |
| 部分成功・結果不明を集約する | Merchant Responseを共通結果へ変換する |
| Repository InterfaceとTransaction境界 | Domain OpenAPIの注文・予約Schema |

Coreへ`travel.hotel`や`retail.shoes`の分岐を書かない。Domain ExtensionがMerchant注文・予約Adapterを提供する。

## 実装済みの条件

- Purchase ItemごとのMerchantFulfillmentを永続化する。
- Merchantへの送信前に冪等キーを保存する。
- Hotelは予約、Shoesは注文としてDomain固有Contractを生成する。
- 全Fulfillmentが`CONFIRMED`の場合だけPurchaseを`CONFIRMED`へ変更する。
- `UNKNOWN`は同じ冪等キーでMerchantへ照会してから安全に再送する。
- `REJECTED`、`UNKNOWN`、部分成功でPurchaseを誤って`CONFIRMED`にしない。
- Stripeの成功を仮定するMock実装は置かない。
- 将来のPayment判定挿入箇所は購入オーケストレーションの1か所にコメントで明示する。
