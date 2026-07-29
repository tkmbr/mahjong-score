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

## 現在のMVP

- 対局日の作成
- 4人分の名前と素点の入力
- 順位と収支の自動計算
- 半荘結果の一覧表示
- SQLiteへの永続化

標準ルールは「25,000点持ち・30,000点返し、順位ウマ +20 / +10 / -10 / -20」です。
