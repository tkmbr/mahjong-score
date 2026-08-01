package domain

import "time"

const (
	MinPlayerCount = 3
	MaxPlayerCount = 4
)

type Session struct {
	ID       int64     `json:"id"`
	Name     string    `json:"name"`
	PlayedAt time.Time `json:"playedAt"`
}

type Result struct {
	PlayerName string `json:"playerName"`
	Score      int    `json:"score"`
}

type Game struct {
	ID        int64     `json:"id"`
	SessionID int64     `json:"sessionId"`
	CreatedAt time.Time `json:"createdAt"`
	Results   []Result  `json:"results"`
}

type SessionWithGames struct {
	Session Session
	Games   []Game
}
