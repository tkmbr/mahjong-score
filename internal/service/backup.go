package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/tkmbr/mahjong-score/internal/domain"
)

const (
	BackupFormat  = "mahjong-score"
	BackupVersion = 1
)

type Backup struct {
	Format     string          `json:"format"`
	Version    int             `json:"version"`
	ExportedAt time.Time       `json:"exportedAt"`
	Sessions   []BackupSession `json:"sessions"`
}

type BackupSession struct {
	Name     string       `json:"name"`
	PlayedAt string       `json:"playedAt"`
	Games    []BackupGame `json:"games"`
}

type BackupGame struct {
	CreatedAt time.Time       `json:"createdAt"`
	Results   []domain.Result `json:"results"`
}

func ValidateBackup(backup Backup) ([]domain.SessionWithGames, error) {
	if backup.Format != BackupFormat {
		return nil, errors.New("Mahjong Scoreのバックアップファイルではありません")
	}
	if backup.Version != BackupVersion {
		return nil, fmt.Errorf("対応していないバックアップバージョンです: %d", backup.Version)
	}
	if len(backup.Sessions) > 10000 {
		return nil, errors.New("対局日の件数が多すぎます")
	}

	sessions := make([]domain.SessionWithGames, 0, len(backup.Sessions))
	gameCount := 0
	for sessionIndex, inputSession := range backup.Sessions {
		name := strings.TrimSpace(inputSession.Name)
		if name == "" {
			return nil, fmt.Errorf("%d件目の対局日の名前が空です", sessionIndex+1)
		}
		playedAt, err := time.Parse("2006-01-02", inputSession.PlayedAt)
		if err != nil {
			return nil, fmt.Errorf("%d件目の対局日の日付が正しくありません", sessionIndex+1)
		}

		item := domain.SessionWithGames{
			Session: domain.Session{Name: name, PlayedAt: playedAt},
			Games:   make([]domain.Game, 0, len(inputSession.Games)),
		}
		for gameIndex, inputGame := range inputSession.Games {
			gameCount++
			if gameCount > 100000 {
				return nil, errors.New("半荘結果の件数が多すぎます")
			}
			if inputGame.CreatedAt.IsZero() {
				return nil, fmt.Errorf("%d件目の対局日の%d件目の記録日時が正しくありません", sessionIndex+1, gameIndex+1)
			}
			results, err := ValidateResults(inputGame.Results)
			if err != nil {
				return nil, fmt.Errorf("%d件目の対局日の%d件目の半荘結果: %w", sessionIndex+1, gameIndex+1, err)
			}
			item.Games = append(item.Games, domain.Game{
				CreatedAt: inputGame.CreatedAt,
				Results:   results,
			})
		}
		sessions = append(sessions, item)
	}
	return sessions, nil
}
