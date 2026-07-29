package sqlite

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
)

func TestOpenMigratesLegacyRawScore(t *testing.T) {
	path := filepath.Join(t.TempDir(), "legacy.db")
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}

	const legacySchema = `
CREATE TABLE sessions (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL,
	played_at TEXT NOT NULL
);
CREATE TABLE games (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	session_id INTEGER NOT NULL REFERENCES sessions(id),
	created_at TEXT NOT NULL
);
CREATE TABLE game_results (
	game_id INTEGER NOT NULL REFERENCES games(id),
	position INTEGER NOT NULL,
	player_name TEXT NOT NULL,
	raw_score INTEGER NOT NULL,
	rank INTEGER NOT NULL,
	point REAL NOT NULL,
	PRIMARY KEY (game_id, position)
);
INSERT INTO sessions (id, name, played_at)
	VALUES (1, '旧データ', '2026-07-30T00:00:00Z');
INSERT INTO games (id, session_id, created_at)
	VALUES (1, 1, '2026-07-30T00:00:00Z');
INSERT INTO game_results (game_id, position, player_name, raw_score, rank, point)
	VALUES (1, 0, '東', 40000, 1, 50);`
	if _, err := db.Exec(legacySchema); err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}

	repository, err := Open(path)
	if err != nil {
		t.Fatalf("Open returned error: %v", err)
	}
	defer repository.Close()

	games, err := repository.ListGames(context.Background(), 1)
	if err != nil {
		t.Fatalf("ListGames returned error: %v", err)
	}
	if len(games) != 1 || len(games[0].Results) != 1 {
		t.Fatalf("games = %#v, want one game with one result", games)
	}
	if games[0].Results[0].Score != 40 {
		t.Errorf("migrated score = %d, want 40", games[0].Results[0].Score)
	}
}
