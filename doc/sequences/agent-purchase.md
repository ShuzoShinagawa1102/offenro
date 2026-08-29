# Agent購入シーケンス

## 今回の実装範囲

```mermaid
sequenceDiagram
    actor User
    participant Agent
    participant Offenro as Offenro API
    participant DB as PostgreSQL
    participant Merchant as Merchant Live API

    User->>Agent: 商品・サービスを依頼
    Agent->>Offenro: Domain Search
    Offenro->>DB: Discovery Index／Active Capability取得
    DB-->>Offenro: Merchant候補
    Offenro->>Merchant: Live Search
    Merchant-->>Offenro: merchant_offer_ref＋最新Offer
    Offenro-->>Agent: Offenro offer_id＋Offer

    Agent->>Offenro: POST /v1/carts
    Offenro->>DB: Cart ACTIVE
    Agent->>Offenro: POST /v1/carts/{id}/items (offer_id)
    Offenro->>Offenro: offer_idを検証・merchant_offer_refを復元
    Offenro->>DB: CartItem保存

    Agent->>Offenro: POST /v1/carts/{id}/checkout
    Offenro->>DB: Cart ACTIVE／Items取得
    loop Active CartItem
        Offenro->>Merchant: Revalidate(merchant_offer_ref)
        Merchant-->>Offenro: availability＋最新価格＋Snapshot
    end

    alt 全Offerが有効
        Offenro->>DB: Transaction開始
        Offenro->>DB: Purchase CREATED／PurchaseItems生成
        Offenro->>DB: Cart CHECKED_OUT
        Offenro->>DB: Commit
        Offenro-->>Agent: 201 Purchase CREATED
    else 失効・在庫切れ・価格不整合
        Offenro-->>Agent: 409 Conflict
        Note over Offenro,DB: CartはACTIVEのまま
    end

    Agent->>Offenro: POST /v1/purchases/{id}/confirm
    Offenro->>DB: ItemごとにFulfillment PENDING＋冪等キーを保存
    loop Purchase Item
        Offenro->>Merchant: 注文・予約作成（冪等キー）
        Merchant-->>Offenro: CONFIRMED／REJECTED／結果不明
        Offenro->>DB: Fulfillment状態を更新
    end
    alt 全FulfillmentがCONFIRMED
        Offenro->>DB: Purchase CONFIRMED
        Offenro-->>Agent: 200 Purchase CONFIRMED
    else PENDING／UNKNOWN／一部失敗
        Offenro-->>Agent: 202 Purchase CREATED
    end
```

`Purchase.CREATED`はMerchant側の注文・予約成立を意味しない。`Purchase.CONFIRMED`は全Purchase ItemのMerchant注文が成立した時点にだけ設定する。結果不明時は同じ冪等キーでMerchantの状態を照会し、二重注文を防ぐ。

## 次工程のStripe決済

```mermaid
sequenceDiagram
    actor User
    participant Agent
    participant Offenro
    participant DB as PostgreSQL
    participant Stripe
    participant Merchant

    Agent->>Offenro: Payment開始
    Offenro->>DB: Payment PENDING
    Offenro->>Stripe: PaymentIntent作成(amount, currency)
    Stripe-->>Offenro: client_secret
    Offenro-->>Agent: client_secret
    Agent->>Stripe: Paymentを確認
    Stripe-->>Offenro: Webhook payment_intent.succeeded
    Offenro->>Offenro: Webhook署名検証・冪等性確認
    Offenro->>DB: Payment CAPTURED
    Offenro->>Merchant: 注文・予約作成（後工程）
    Merchant-->>Offenro: 注文・予約成立
    Offenro->>DB: Purchase CONFIRMED
    Offenro-->>Agent: 購入成立
```

Payment状態とPurchase状態を同一視しない。Stripeの最終結果はAgentからの戻りではなく署名検証済みWebhookで反映する。
