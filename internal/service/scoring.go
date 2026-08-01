package service

import (
	"errors"
	"strings"

	"github.com/tkmbr/mahjong-score/internal/domain"
)

var (
	ErrPlayerCount      = errors.New("3人または4人分の結果を入力してください")
	ErrBlankPlayerScore = errors.New("名前が空欄のプレイヤーにはスコアを入力できません")
	ErrScoreTotal       = errors.New("スコアの合計が0になるように入力してください")
)

func ValidateResults(input []domain.Result) ([]domain.Result, error) {
	results := make([]domain.Result, 0, domain.MaxPlayerCount)
	for _, result := range input {
		result.PlayerName = strings.TrimSpace(result.PlayerName)
		if result.PlayerName == "" {
			if result.Score != 0 {
				return nil, ErrBlankPlayerScore
			}
			continue
		}
		results = append(results, result)
	}
	if len(results) < domain.MinPlayerCount || len(results) > domain.MaxPlayerCount {
		return nil, ErrPlayerCount
	}

	seenNames := make(map[string]struct{}, domain.MaxPlayerCount)
	total := 0
	for i := range results {
		if _, exists := seenNames[results[i].PlayerName]; exists {
			return nil, errors.New("プレイヤー名が重複しています")
		}
		seenNames[results[i].PlayerName] = struct{}{}
		total += results[i].Score
	}
	if total != 0 {
		return nil, ErrScoreTotal
	}
	return results, nil
}
