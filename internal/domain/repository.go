package domain

import "context"

type Repository interface {
	CreateSession(ctx context.Context, session Session) (Session, error)
	ListSessions(ctx context.Context) ([]Session, error)
	CreateGame(ctx context.Context, game Game) (Game, error)
	ListGames(ctx context.Context, sessionID int64) ([]Game, error)
}
