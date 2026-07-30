package domain

import "context"

type Repository interface {
	CreateSession(ctx context.Context, session Session) (Session, error)
	ListSessions(ctx context.Context) ([]Session, error)
	UpdateSession(ctx context.Context, session Session) (bool, error)
	DeleteSession(ctx context.Context, sessionID int64) (bool, error)
	CreateGame(ctx context.Context, game Game) (Game, error)
	ListGames(ctx context.Context, sessionID int64) ([]Game, error)
	UpdateGame(ctx context.Context, game Game) (bool, error)
	DeleteGame(ctx context.Context, sessionID, gameID int64) (bool, error)
	ImportSessions(ctx context.Context, sessions []SessionWithGames) error
}
