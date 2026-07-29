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
	raw_score INTEGER NOT NULL,
	rank INTEGER NOT NULL,
	point REAL NOT NULL,
	PRIMARY KEY (game_id, position)
);`
	if _, err := r.db.ExecContext(ctx, schema); err != nil {
		return fmt.Errorf("migrate database: %w", err)
	}
	return nil
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
			INSERT INTO game_results (game_id, position, player_name, raw_score, rank, point)
			VALUES (?, ?, ?, ?, ?, ?)`,
			game.ID, position, score.PlayerName, score.RawScore, score.Rank, score.Point,
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
		SELECT g.id, g.created_at, r.player_name, r.raw_score, r.rank, r.point
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
		if err := rows.Scan(&gameID, &createdAt, &result.PlayerName, &result.RawScore, &result.Rank, &result.Point); err != nil {
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
