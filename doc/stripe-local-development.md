# Stripe CLIローカル開発

Stripe連携のローカル開発では、Live環境ではなくOffenro用Sandboxの`offenro-dev`を使用する。

## 初回セットアップ

1. [Stripe公式の手順](https://docs.stripe.com/cli/install)でStripe CLIをインストールし、確認する。

   ```bash
   npm install -g @stripe/cli
   stripe --version
   ```

2. Stripe CLIへログインする。

   ```bash
   stripe login
   ```

3. 表示されたURLをブラウザで開き、接続先にSandboxの`offenro-dev`を選択してアクセスを許可する。

4. ログイン先とSandboxへのAPI接続を確認する。

   ```bash
   stripe login list
   stripe balance retrieve
   ```

   `offenro-dev`のログイン情報が表示され、Balance取得が成功すればセットアップ完了。別の環境へ接続されている場合は、`stripe login`をやり直して`offenro-dev`を選択する。

Sandboxの選択方法は[Stripe公式のSandbox管理手順](https://docs.stripe.com/sandboxes/dashboard/manage)を参照する。

## 環境変数

初回のみ`.env.example`をコピーし、必要になったStripe Sandboxの値をローカルの`.env`へ設定する。

```bash
cp .env.example .env
```

- `STRIPE_SECRET_KEY`: ServerからStripe APIを呼ぶためのSandbox Secret Key
- `STRIPE_PUBLISHABLE_KEY`: Clientで使用するSandbox Publishable Key
- `STRIPE_WEBHOOK_SECRET`: Stripe CLIのWebhook転送開始時に表示される署名検証用Secret

実値は`.env.example`やGit管理対象のファイルへ記載しない。

## 今後のWebhook開発

Payment／Webhook APIは現在未実装。実装後はStripe CLIでイベントをローカルServerへ転送する。

```bash
stripe listen --forward-to http://localhost:8080/<webhook-path>
```

`stripe listen`が表示したWebhook Secretをローカルの`.env`に`STRIPE_WEBHOOK_SECRET`として設定し、署名検証に使用する。必要に応じてSandboxのテストイベントを送信する。

```bash
stripe trigger payment_intent.succeeded
```

Webhookの転送とテスト方法は[Stripe公式Webhook手順](https://docs.stripe.com/webhooks)を参照する。
