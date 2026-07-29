package service

import (
	"errors"
	"strings"

	"github.com/tkmbr/mahjong-score/internal/domain"
)

var (
	ErrPlayerCount = errors.New("4人分の結果を入力してください")
	ErrPlayerName  = errors.New("プレイヤー名を入力してください")
	ErrScoreTotal  = errors.New("スコアの合計が0になるように入力してください")
)

func ValidateResults(input []domain.Result) ([]domain.Result, error) {
	if len(input) != domain.PlayerCount {
		return nil, ErrPlayerCount
	}

	results := append([]domain.Result(nil), input...)
	seenNames := make(map[string]struct{}, domain.PlayerCount)
	total := 0
	for i := range results {
		results[i].PlayerName = strings.TrimSpace(results[i].PlayerName)
		if results[i].PlayerName == "" {
			return nil, ErrPlayerName
		}
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
