# API境界

OffenroのOpenAPI Contractは、利用者と実装責任で次の3種類に分ける。

| 分類 | 利用者 | 実装者 | 配置 |
|---|---|---|---|
| Agent API | Agent開発者 | Offenro | `api/protocol`、`api/domains/{domain}/agent.openapi.yaml` |
| Management API | Offenro運用者。将来はMerchant Console | Offenro | `api/management` |
| Merchant Live API | Offenro | 各Merchant | `api/domains/{domain}/merchant.openapi.yaml` |

Management API ContractをFrontend配下へ置かない。将来のMerchant ConsoleはManagement APIを利用する別クライアントとする。

Domainの`schemas.yaml`はAgent APIとMerchant Live APIの共有モデルの正である。ただしOffer識別子の公開境界は分ける。

```text
Merchant Live API: merchant_offer_ref
             ↓ Offenroが暗号化・認証したOpaque IDへ変換
Agent API:          offer_id
```

- `merchant_offer_ref`はMerchantがOfferを再特定するためのOpaque参照値であり、Agentへ公開しない。
- `offer_id`はOffenroが発行し、Merchant、Capability、Domain、`merchant_offer_ref`、有効期限を安全に解決できるものとする。
- 検索候補そのものはDBへ保存しない。
- Cart追加後は`offer_id`と`merchant_offer_ref`を別Columnとして保持する。

Merchant Live APIの注文・予約作成は`Idempotency-Key`を必須とし、同じキーから結果を取得する照会APIもDomain Contractに定義する。OffenroはMerchantへ送信する前にキーを永続化し、応答が不明な場合は照会してから同じキーで安全に再送する。

Management APIは認証なしで公開しない。今回のPrototypeでは環境変数による管理Tokenで保護し、Merchant自身のAccount・API Key管理は後工程とする。
