package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/tkmbr/mahjong-score/internal/domain"
	_ "modernc.org/sqlite"
)

type Repository struct {
	db *sql.DB
}

func Open(path string) (*Repository, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("create database directory: %w", err)
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)

	repository := &Repository{db: db}
	if err := repository.migrate(context.Background()); err != nil {
		db.Close()
		return nil, err
	}
	return repository, nil
}

func (r *Repository) Close() error {
	return r.db.Close()
}

func (r *Repository) migrate(ctx context.Context) error {
	const schema = `
PRAGMA foreign_keys = ON;
PRAGMA journal_mode = WAL;

CREATE TABLE IF NOT EXISTS sessions (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL,
	played_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS games (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	session_id INTEGER NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
	created_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS game_results (
	game_id INTEGER NOT NULL REFERENCES games(id) ON DELETE CASCADE,
	position INTEGER NOT NULL,
	player_name TEXT NOT NULL,
	score INTEGER NOT NULL,
	PRIMARY KEY (game_id, position)
);`
	if _, err := r.db.ExecContext(ctx, schema); err != nil {
		return fmt.Errorf("migrate database: %w", err)
	}
	return r.migrateLegacyResults(ctx)
}

func (r *Repository) migrateLegacyResults(ctx context.Context) error {
	rows, err := r.db.QueryContext(ctx, `PRAGMA table_info(game_results)`)
	if err != nil {
		return fmt.Errorf("inspect game results schema: %w", err)
	}

	hasRawScore := false
	for rows.Next() {
		var cid, notNull, primaryKey int
		var name, columnType string
		var defaultValue any
		if err := rows.Scan(&cid, &name, &columnType, &notNull, &defaultValue, &primaryKey); err != nil {
			rows.Close()
			return fmt.Errorf("inspect game results column: %w", err)
		}
		if name == "raw_score" {
			hasRawScore = true
		}
	}
	if err := rows.Close(); err != nil {
		return err
	}
	if !hasRawScore {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	const migration = `
CREATE TABLE game_results_new (
	game_id INTEGER NOT NULL REFERENCES games(id) ON DELETE CASCADE,
	position INTEGER NOT NULL,
	player_name TEXT NOT NULL,
	score INTEGER NOT NULL,
	PRIMARY KEY (game_id, position)
);
INSERT INTO game_results_new (game_id, position, player_name, score)
	SELECT game_id, position, player_name, raw_score / 1000
	FROM game_results;
DROP TABLE game_results;
ALTER TABLE game_results_new RENAME TO game_results;`
	if _, err := tx.ExecContext(ctx, migration); err != nil {
		return fmt.Errorf("migrate legacy game results: %w", err)
	}
	return tx.Commit()
}

func (r *Repository) CreateSession(ctx context.Context, session domain.Session) (domain.Session, error) {
	result, err := r.db.ExecContext(ctx,
		`INSERT INTO sessions (name, played_at) VALUES (?, ?)`,
		session.Name, session.PlayedAt.Format(time.RFC3339),
	)
	if err != nil {
		return domain.Session{}, err
	}
	session.ID, err = result.LastInsertId()
	return session, err
}

func (r *Repository) ListSessions(ctx context.Context) ([]domain.Session, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, name, played_at FROM sessions ORDER BY played_at DESC, id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sessions := make([]domain.Session, 0)
	for rows.Next() {
		var session domain.Session
		var playedAt string
		if err := rows.Scan(&session.ID, &session.Name, &playedAt); err != nil {
			return nil, err
		}
		session.PlayedAt, err = time.Parse(time.RFC3339, playedAt)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, session)
	}
	return sessions, rows.Err()
}

func (r *Repository) CreateGame(ctx context.Context, game domain.Game) (domain.Game, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Game{}, err
	}
	defer tx.Rollback()

	game.CreatedAt = time.Now()
	result, err := tx.ExecContext(ctx,
		`INSERT INTO games (session_id, created_at) VALUES (?, ?)`,
		game.SessionID, game.CreatedAt.Format(time.RFC3339),
	)
	if err != nil {
		return domain.Game{}, err
	}
	game.ID, err = result.LastInsertId()
	if err != nil {
		return domain.Game{}, err
	}

	for position, score := range game.Results {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO game_results (game_id, position, player_name, score)
			VALUES (?, ?, ?, ?)`,
			game.ID, position, score.PlayerName, score.Score,
		)
		if err != nil {
			return domain.Game{}, err
		}
	}

	if err := tx.Commit(); err != nil {
		return domain.Game{}, err
	}
	return game, nil
}

func (r *Repository) ListGames(ctx context.Context, sessionID int64) ([]domain.Game, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT g.id, g.created_at, r.player_name, r.score
		FROM games g
		JOIN game_results r ON r.game_id = g.id
		WHERE g.session_id = ?
		ORDER BY g.id DESC, r.position ASC`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	games := make([]domain.Game, 0)
	indexByID := make(map[int64]int)
	for rows.Next() {
		var gameID int64
		var createdAt string
		var result domain.Result
		if err := rows.Scan(&gameID, &createdAt, &result.PlayerName, &result.Score); err != nil {
			return nil, err
		}
		index, exists := indexByID[gameID]
		if !exists {
			parsed, err := time.Parse(time.RFC3339, createdAt)
			if err != nil {
				return nil, err
			}
			index = len(games)
			indexByID[gameID] = index
			games = append(games, domain.Game{ID: gameID, SessionID: sessionID, CreatedAt: parsed})
		}
		games[index].Results = append(games[index].Results, result)
	}
	return games, rows.Err()
}
