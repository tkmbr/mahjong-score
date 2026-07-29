package domain

import "time"

const PlayerCount = 4

type Session struct {
	ID       int64     `json:"id"`
	Name     string    `json:"name"`
	PlayedAt time.Time `json:"playedAt"`
}

type Result struct {
	PlayerName string  `json:"playerName"`
	RawScore   int     `json:"rawScore"`
	Rank       int     `json:"rank"`
	Point      float64 `json:"point"`
}

type Game struct {
	ID        int64     `json:"id"`
	SessionID int64     `json:"sessionId"`
	CreatedAt time.Time `json:"createdAt"`
	Results   []Result  `json:"results"`
}

type Rule struct {
	StartingScore int
	ReturnScore   int
	RankBonus     [PlayerCount]float64
}

var StandardRule = Rule{
	StartingScore: 25000,
	ReturnScore:   30000,
	RankBonus:     [PlayerCount]float64{20, 10, -10, -20},
}
