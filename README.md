# Mahjong Score

麻雀の半荘結果を表形式で入力・集計する、小さなWebアプリです。

## 技術構成

- Go 1.26（Dockerイメージは1.26.5に固定）
- SQLite
- HTML / CSS / Vanilla JavaScript
- Docker Compose

フロントエンドにビルド工程はなく、静的ファイルはGoの実行ファイルに埋め込みます。

## 起動

Windows側のPowerShellでリポジトリのディレクトリを開き、WSL内のDockerを使用します。
WSL内でDocker Engineが起動している必要があります。

```powershell
wsl docker compose up --build
```

ブラウザで <http://localhost:8080> を開きます。SQLiteデータはDockerボリュームに保存されます。

## Tailscale経由で共有する

インターネットへ直接公開せず、招待したTailscaleユーザーだけに共有する場合の手順です。
閲覧者にもTailscaleアカウントとTailscaleクライアントが必要です。

### 1. Tailscale Serveを開始する

アプリを`localhost:8080`で起動した状態で、アプリを実行しているマシン上から次を実行します。

```bash
tailscale serve --bg 8080
```

設定を確認します。

```bash
tailscale serve status
```

次のように表示されれば、TailscaleがHTTPSの443番ポートで受けた通信を
`127.0.0.1:8080`へ転送しています。

```text
https://<machine-name>.<tailnet-name>.ts.net (tailnet only)
|-- / proxy http://127.0.0.1:8080
```

### 2. 閲覧者を招待する

Tailscale管理画面の **Users > Invite external users** から、閲覧者を`Member`として
招待します。閲覧者が招待を承認してTailscaleクライアントからこのtailnetへ接続した後、
User approvalが有効な場合は管理者側でもユーザーを承認します。

Personalプランの無料ユーザー数には所有者自身と参加済みの招待ユーザーが含まれます。

### 3. WebアクセスだけをGrantで許可する

公開元マシンのTailscale IPv4アドレスを確認します。

```bash
tailscale ip -4
```

管理画面の **Access controls** を開き、既存ポリシーをバックアップしてから編集します。
次は、所有者の端末間通信を維持し、招待ユーザーには公開元マシンのHTTPSだけを許可する
最小構成例です。IPアドレスとメールアドレスは実際の値に置き換えます。

```json
{
  "hosts": {
    "mahjong-score": "100.64.0.10"
  },

  "grants": [
    {
      "src": ["owner@example.com"],
      "dst": ["autogroup:self"],
      "ip": ["*"]
    },
    {
      "src": ["guest@example.com"],
      "dst": ["mahjong-score"],
      "ip": ["tcp:443"]
    }
  ],

  "tests": [
    {
      "src": "guest@example.com",
      "accept": ["mahjong-score:443"],
      "deny": [
        "mahjong-score:22",
        "mahjong-score:8080"
      ]
    }
  ]
}
```

Grantで許可するポートは、バックエンドの8080番ではなくTailscale Serveが待ち受ける
TCP 443番です。

`tests`は実際のアクセス権を追加する設定ではありません。ポリシー保存時に、閲覧者から
443番への接続が許可され、22番と8080番への接続が拒否されることを静的に検証します。
将来のポリシー変更で意図せず権限が広がるのを検出するために残します。

#### allow-allルールに注意する

Grantは加算式であり、狭いGrantを追加しても既存の広い許可を上書きできません。
初期設定などに次のようなallow-allのGrantが残っていると、閲覧者も他の端末やポートへ
アクセスできます。

```json
{
  "src": ["*"],
  "dst": ["*"],
  "ip": ["*"]
}
```

旧形式の`acls`を併用している場合も含め、閲覧者に一致する広い許可がないことを確認します。
一方、既存ポリシーを上記例で単純に置き換えると、SSH、Taildrop、exit nodeなど現在利用中の
通信を止める可能性があります。必要な既存ルールを個別に残したうえでallow-allを整理します。

### 4. 接続を確認する

閲覧者が対象tailnetへ接続した状態で、Serveのステータスに表示されたURLを開きます。

```text
https://<machine-name>.<tailnet-name>.ts.net
```

CLIから確認する場合は次を実行します。

```bash
curl -I https://<machine-name>.<tailnet-name>.ts.net
```

443番への接続が成功し、次のような直接接続が失敗することも確認します。

```bash
curl --connect-timeout 5 http://<machine-name>.<tailnet-name>.ts.net:8080
ssh <machine-name>.<tailnet-name>.ts.net
```

Grantはネットワーク接続を制限するもので、Webアプリ内のユーザー認証にはなりません。
機密性の高いデータを扱う場合は、後述の`oauth2-proxy`などアプリケーション層の認証も
追加します。

### 5. Serveを停止する

```bash
tailscale serve off
```

## 開発コマンド

```powershell
wsl docker compose build
wsl sh -lc 'docker run --rm -v "$PWD:/src" -w /src golang:1.26.5-bookworm go test ./...'
wsl docker compose down
```

テストは公式Goイメージ内で実行するため、Windows側にGoをインストールする必要はありません。
ローカルでGoコマンドを直接実行する場合はGo 1.26が必要です。Go 1.21以降の
`GOTOOLCHAIN=auto`（デフォルト）を利用している環境では、必要なGo 1.26
ツールチェーンが自動的にダウンロードされます。インストール済みのGo自体の
バージョンは変更されません。

開発中にSQLiteのスキーマを変更した場合は、migrationを行わずボリュームを作り直します。

```powershell
wsl docker compose down --volumes
wsl docker compose up --build
```

## 現在のMVP

- 対局日の作成
- 対局日の編集・削除
- 4人分の名前とスコアの入力
- 半荘結果の一覧表示
- 半荘結果の編集・削除
- 全データのJSONエクスポート
- JSONバックアップのインポート（既存データへ追加）
- SQLiteへの永続化

スコアは1000点単位の整数で入力します。たとえば40,000点は`40`です。
ウマ・オカなどのルール計算は行わず、入力値をそのまま保存します。
4人分のスコアは、合計が`0`になる必要があります。

## 今後の開発候補

### 認証・認可

現在はTailscale内での利用のみを想定していますが、アプリ自体には認証・認可機能がありません。
今後は、アプリの前段に`oauth2-proxy`を配置し、OIDC対応の認証プロバイダーへ
ログインを委譲する構成を第一候補とします。

想定する構成は次のとおりです。

```text
ブラウザ
  -> Tailscale Serve
  -> oauth2-proxy
  -> Mahjong Score
```

- Mahjong Scoreのポートはホストへ直接公開せず、Docker内部からのみ接続可能にする
- `oauth2-proxy`が認証済みリクエストだけをアプリへ転送する
- Go側でも認証済みユーザーのヘッダーを検証し、直接アクセスを拒否する
- 一般利用者と管理者を分け、削除・インポート・エクスポートなどの権限を制御する
- 操作ログに認証済みユーザーを記録する
- OIDCクライアントシークレットやCookie暗号鍵はGitで管理しない

パスワードをアプリで保持せずに済むよう、認証プロバイダーにはパスキー
（WebAuthn）対応のサービスを利用します。パスキーの登録、署名検証、紛失時の復旧は
認証プロバイダーへ委譲し、Mahjong ScoreはOIDCの認証結果のみを利用する方針です。

将来、外部の認証プロバイダーへ依存しないことが要件になった場合は、
Go用のWebAuthnライブラリを使用した直接実装も検討します。その場合は、固定したHTTPS
ドメイン、RP IDとOriginの厳密な検証、複数パスキー、セッション管理、CSRF対策、
招待制のユーザー登録、アカウント復旧手段をあわせて設計する必要があります。

### 公開範囲と堅牢化

- Dockerの公開先をlocalhostに限定し、Tailscale Serve経由でのみアクセス可能にする
- Tailscale Grantsでアクセス可能なユーザーと端末を限定する
- HTTPサーバーにRead、Write、Idleの各タイムアウトを設定する
- 名前・スコア・インポート件数に実用的な上限を設ける
- SQLiteデータの定期バックアップを用意する
- Content Security Policyなどのセキュリティヘッダーを追加する
