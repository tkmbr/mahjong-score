package service

import (
	"testing"

	"github.com/tkmbr/mahjong-score/internal/domain"
)

func TestValidateResults(t *testing.T) {
	input := []domain.Result{
		{PlayerName: " 東 ", Score: 40},
		{PlayerName: "南", Score: 30},
		{PlayerName: "西", Score: 20},
		{PlayerName: "北", Score: 10},
	}

	got, err := ValidateResults(input)
	if err != nil {
		t.Fatalf("ValidateResults returned error: %v", err)
	}
	if got[0].PlayerName != "東" {
		t.Errorf("PlayerName = %q, want %q", got[0].PlayerName, "東")
	}
	if got[0].Score != 40 {
		t.Errorf("Score = %d, want 40", got[0].Score)
	}
}

func TestValidateResultsAllowsTiesAndAnyTotal(t *testing.T) {
	input := []domain.Result{
		{PlayerName: "A", Score: 12},
		{PlayerName: "B", Score: 12},
		{PlayerName: "C", Score: -5},
		{PlayerName: "D", Score: 0},
	}

	if _, err := ValidateResults(input); err != nil {
		t.Fatalf("ValidateResults returned error: %v", err)
	}
}
