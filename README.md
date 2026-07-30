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
