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
