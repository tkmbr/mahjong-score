# Mahjong Score

麻雀の半荘結果を表形式で入力・集計する、小さなWebアプリです。

## 技術構成

- Go 1.24
- SQLite
- HTML / CSS / Vanilla JavaScript
- Docker Compose

フロントエンドにビルド工程はなく、静的ファイルはGoの実行ファイルに埋め込みます。

## 起動

```powershell
docker compose up --build
```

ブラウザで <http://localhost:8080> を開きます。SQLiteデータはDockerボリュームに保存されます。

## 開発コマンド

```powershell
docker compose build
docker compose run --rm app go test ./...
docker compose down
```

開発中にSQLiteのスキーマを変更した場合は、migrationを行わずボリュームを作り直します。

```powershell
docker compose down --volumes
docker compose up --build
```

## 現在のMVP

- 対局日の作成
- 4人分の名前とスコアの入力
- 半荘結果の一覧表示
- SQLiteへの永続化

スコアは1000点単位の整数で入力します。たとえば40,000点は`40`です。
ウマ・オカなどのルール計算は行わず、入力値をそのまま保存します。
4人分のスコアは、合計が`0`になる必要があります。
