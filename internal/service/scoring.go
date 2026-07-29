package service

import (
	"errors"
	"sort"
	"strings"

	"github.com/tkmbr/mahjong-score/internal/domain"
)

var (
	ErrPlayerCount    = errors.New("4人分の結果を入力してください")
	ErrPlayerName     = errors.New("プレイヤー名を入力してください")
	ErrDuplicateScore = errors.New("同点にはまだ対応していません")
	ErrScoreTotal     = errors.New("素点の合計が100,000点になるように入力してください")
)

func Calculate(input []domain.Result, rule domain.Rule) ([]domain.Result, error) {
	if len(input) != domain.PlayerCount {
		return nil, ErrPlayerCount
	}

	results := append([]domain.Result(nil), input...)
	seenNames := make(map[string]struct{}, domain.PlayerCount)
	seenScores := make(map[int]struct{}, domain.PlayerCount)
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
		if _, exists := seenScores[results[i].RawScore]; exists {
			return nil, ErrDuplicateScore
		}
		seenScores[results[i].RawScore] = struct{}{}
		total += results[i].RawScore
	}
	if total != rule.StartingScore*domain.PlayerCount {
		return nil, ErrScoreTotal
	}

	order := make([]int, len(results))
	for i := range order {
		order[i] = i
	}
	sort.Slice(order, func(i, j int) bool {
		return results[order[i]].RawScore > results[order[j]].RawScore
	})

	for rankIndex, resultIndex := range order {
		results[resultIndex].Rank = rankIndex + 1
		base := float64(results[resultIndex].RawScore-rule.ReturnScore) / 1000
		if rankIndex == 0 {
			base += float64((rule.ReturnScore-rule.StartingScore)*domain.PlayerCount) / 1000
		}
		results[resultIndex].Point = round1(base + rule.RankBonus[rankIndex])
	}
	return results, nil
}

func round1(value float64) float64 {
	if value >= 0 {
		return float64(int(value*10+0.5)) / 10
	}
	return float64(int(value*10-0.5)) / 10
}
