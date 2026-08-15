# Offenro

## 概要

* Offenroは、**成果報酬型のエージェントコマースを成立させるためのプロトコル**である。
* Offenroを利用するエージェント開発者は、ユーザに購買アプリやAgentを提供する。
* Agentを介して購買が成立した場合、ユーザから支払われた代金をもとに、**Merchant・Agent・Offenro Protocol**へ事前に定めた金額を分配する。
* AgentはOffenroが提供するWeb APIを利用し、ユーザの要求に合う商品・サービスのOfferを検索・購入する。
* Offenroの検索対象は、事前に提携・登録されたMerchantが提供する商品・サービスである。
* Merchantは、ホテル、小売、旅行、美容、製造業サプライチェーン、ライブコマースなど、様々な購買Domainを想定する。
* Merchantは、Offenroが定義するOpenAPI仕様に従い、商品検索・購入等のAPIを実装し、そのAPI URLをOffenroへ登録する。（管理サイト想定）
* OffenroはMerchant APIから商品・サービスの比較的静的な情報を取得し、**どのMerchantへ問い合わせるべきかを判断するためのDiscovery Index**を作成する。
* 価格・在庫など変動性の高い情報は古いIndexを利用せず、実際の検索時にMerchant APIへ問い合わせて取得する。
* Discovery Indexは定期的にMerchant APIから再構築・更新する。

---

## 基本フロー

```text
User
↓
Agent
↓
Offenro Web API
↓
SearchOffers
↓
Merchant Discovery
↓
候補Merchant
↓
Merchant API
↓
Offer
↓
Agent
↓
Purchase
↓
Payment / Allocation
↓
Merchant・Agent・Protocolへ分配
```

---

## 仕様

### User

* Agentを利用して商品・サービスを検索・購入する。
* 商品代金を支払う。
* 必要に応じてキャンセル、Refund、Issue等の取引に関与する。

### Agent

* Userの要求を解釈する。
* 検索対象のDomainと検索条件を確定する。
* OffenroのDomain別Web APIを利用してOfferを検索する。
* Userが選択したOfferをOffenro経由で購入する。
* 購買成果に応じてAgent Incentiveを受け取る。

自然言語からDomain・Conditionを確定する処理は、基本的にAgent側の責務とする。

Offenro Coreには、構造化済みのConditionを渡す。

### Offenro Protocol

主な責務は以下。

* Merchant / Domain管理
* Merchant Discovery
* Offer検索
* Purchase管理
* Attribution管理
* Payment管理
* Allocation管理
* Agent Incentive管理
* Settlement管理
* Refund等に伴う資金状態管理

Offer検索の内部イメージ：

```text
Domain固有Condition
↓
SearchOffers
↓
Domain対応Merchant取得
↓
Discovery Indexから候補Merchantを絞る
↓
上位MerchantへLive API Request
↓
Offerを統合
↓
Agentへ返却
```

### Merchant

* Offenro Merchant Protocolに従ったAPIを実装する。
* API URLと対応DomainをOffenroへ登録する。
* 商品・サービスのCatalog情報を提供する。
* 検索条件に応じた最新のOfferを返す。
* 購入された商品・サービスをUserへ提供する。
* Fulfillment、Refund等の取引状態をOffenroへ連携する。

Merchant側APIはDomainごとに定義する。

例：

```text
travel.hotel
├─ Catalog
├─ Search
└─ Purchase
```

### PSP

Stripe等のPayment Service Providerを利用する。

主な役割：

* Userからの決済
* MerchantへのTransfer
* AgentへのTransfer
* Refund
* Payout

Offenro自身が直接資金を保管するのではなく、実際の資金移動はPSPを利用する。

---

## Merchant Discovery

**注意：ここはまだプロトタイプなので変わる可能性がありです。これだけにとらわれないようにしてください。**

すべてのMerchantへ毎回問い合わせるのではなく、検索条件から**Offerを持っている可能性が高いMerchantだけを事前に絞る**。

そのためにDiscovery Indexを利用する。

```text
Merchant Capability
「このMerchantは何のDomainを扱うか」

↓

Discovery Index
「このConditionなら、どのMerchantへ聞くべきか」

↓

Merchant Live API
「今、本当に買えるOfferは何か」
```

Discovery Indexには、価格・在庫等の動的情報は原則保存しない。

Domainごとに検索軸を定義する。

例：

```text
travel.hotel
prefecture_code → Merchant
```

将来的には、

```text
retail.shoes
brand / size → Merchant
```

などへ拡張可能とする。

---

## v0.1 Prototypeの仕様

最初のDomainは、

```text
travel.hotel
```

に限定する。

### 検索条件

```text
TravelHotelCondition
├─ Destination
│  └─ PrefectureCode
├─ Stay
│  ├─ CheckIn
│  └─ CheckOut
├─ Guests
│  ├─ Adults
│  └─ Rooms
└─ Filters
   └─ MaxPrice
```

### Merchant

Prototypeでは30 Merchantを用意する。

* 全国型: 10 Merchant
* 北海道: 2
* 沖縄: 2
* 東北: 2
* 関東: 3
* 東海: 2
* 近畿: 3
* 中国: 2
* 四国: 2
* 九州: 2

全国型Merchantは各200 Hotel、それ以外は各20 Hotel程度を保持する。

合計約2,400 Merchant-Hotelレコードを利用する。

### Discovery

`travel.hotel`では、

```text
prefecture_code
```

をDiscovery Indexの検索軸とする。

例えば神奈川県を検索した場合、

```text
prefecture_code = 14
↓
神奈川県のHotelを扱うMerchantを逆引き
↓
上位10 Merchant
↓
各MerchantへLive Search
↓
最大50 Offer
```

という流れとする。

---

## 技術構成

* Go
* `net/http`
* OpenAPI 3.0.3
* `oapi-codegen`

  * `std-http-server`
  * `strict-server`
* Web Frameworkは現時点では使用しない。
* DBは現時点では未導入。
* Merchant Registry / Discovery IndexはPrototypeではInMemory。
* 将来的にPostgreSQL + pgx + sqlc等への移行を想定する。

---

## ディレクトリ

```text
offenro/
├─ api/
│  └─ openapi.yaml
├─ cmd/
│  ├─ server/
│  └─ merchant-mock/
├─ internal/
│  ├─ api/
│  ├─ server/
│  ├─ search/
│  ├─ merchant/
│  ├─ discovery/
│  └─ mockdata/
├─ oapi-codegen.yaml
├─ go.mod
└─ go.sum
```

主な役割：

```text
api
→ 外部Web API仕様

cmd
→ 実行プログラム

server
→ Web APIとProtocol Coreの境界

search
→ SearchOffers等のProtocol Core

merchant
→ Merchant Registry / Merchant API通信

discovery
→ Merchant Discovery / Index

mockdata
→ Prototype用Merchant・Hotelデータ
```

---

## 開発コマンド

### OpenAPIからGoコード生成

```bash
go tool oapi-codegen -config oapi-codegen.yaml api/openapi.yaml
```

生成先：

```text
internal/api/openapi.gen.go
```

`openapi.gen.go`は直接編集しない。

### Go Module整理

```bash
go mod tidy
```

### 全体のコンパイル・テスト

```bash
go test ./...
```

実装変更後は基本的にこれを実行し、エラーがないことを確認する。

### Offenro Server起動

```bash
go run ./cmd/server
```

### Merchant Mock起動

```bash
go run ./cmd/merchant-mock
```

Discovery IndexをMerchant Mockから構築する場合は、Merchant Mockを先に起動する。

```text
Terminal 1
go run ./cmd/merchant-mock

Terminal 2
go run ./cmd/server
```

### Health Check

```bash
curl http://localhost:8080/health
```

期待値：

```json
{"status":"ok"}
```

### Hotel Offer検索

```bash
curl -X POST http://localhost:8080/v1/travel/hotels/search \
  -H "Content-Type: application/json" \
  -d '{
    "destination": {
      "prefecture_code": "14"
    },
    "stay": {
      "check_in": "2026-09-10",
      "check_out": "2026-09-12"
    },
    "guests": {
      "adults": 2,
      "rooms": 1
    },
    "filters": {
      "max_price": 60000
    }
  }'
```

---

## Version

段階的にProtocolを拡張する。

### v0.1 Prototype

* `travel.hotel`
* Agent向けHotel Search API
* Merchant Capability Registry
* Discovery Index
* Merchant Live Search
* 複数MerchantからのOffer統合

以降、

* Purchase
* Payment
* Allocation
* Refund
* Agent Incentive
* Settlement
* Domain追加

を順次実装する。

---

## 留意点

* **Hotel専用システムを作らない。汎用化を意識**
* 現在は`travel.hotel`のみを実装するが、常に他Domainへ拡張できる構造を維持する。
* Domain固有のCondition・Response・Discovery Logic・Merchant APIと、Domain非依存のProtocol Coreを分離する。（重要）
* 動的な価格・在庫を古いIndexに依存させない。
* 最初から過剰なFrameworkやInfrastructureを導入しない。
* Prototypeで実際に必要になったものから段階的に追加する。
