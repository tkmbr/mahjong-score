package service

import (
	"testing"

	"github.com/tkmbr/mahjong-score/internal/domain"
)

func TestCalculateStandardRule(t *testing.T) {
	input := []domain.Result{
		{PlayerName: "東", RawScore: 42100},
		{PlayerName: "南", RawScore: 28700},
		{PlayerName: "西", RawScore: 19400},
		{PlayerName: "北", RawScore: 9800},
	}

	got, err := Calculate(input, domain.StandardRule)
	if err != nil {
		t.Fatalf("Calculate returned error: %v", err)
	}

	wantPoints := []float64{52.1, 8.7, -20.6, -40.2}
	for i := range got {
		if got[i].Rank != i+1 {
			t.Errorf("result %d rank = %d, want %d", i, got[i].Rank, i+1)
		}
		if got[i].Point != wantPoints[i] {
			t.Errorf("result %d point = %.1f, want %.1f", i, got[i].Point, wantPoints[i])
		}
	}
}

func TestCalculateRejectsDuplicateScores(t *testing.T) {
	input := []domain.Result{
		{PlayerName: "A", RawScore: 25000},
		{PlayerName: "B", RawScore: 25000},
		{PlayerName: "C", RawScore: 30000},
		{PlayerName: "D", RawScore: 20000},
	}

	if _, err := Calculate(input, domain.StandardRule); err != ErrDuplicateScore {
		t.Fatalf("Calculate error = %v, want %v", err, ErrDuplicateScore)
	}
}
