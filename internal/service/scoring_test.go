package service

import (
	"testing"

	"github.com/tkmbr/mahjong-score/internal/domain"
)

func TestValidateResults(t *testing.T) {
	input := []domain.Result{
		{PlayerName: " 東 ", Score: 40},
		{PlayerName: "南", Score: 10},
		{PlayerName: "西", Score: -20},
		{PlayerName: "北", Score: -30},
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

func TestValidateResultsAllowsTies(t *testing.T) {
	input := []domain.Result{
		{PlayerName: "A", Score: 12},
		{PlayerName: "B", Score: 12},
		{PlayerName: "C", Score: -12},
		{PlayerName: "D", Score: -12},
	}

	if _, err := ValidateResults(input); err != nil {
		t.Fatalf("ValidateResults returned error: %v", err)
	}
}

func TestValidateResultsAllowsThreePlayers(t *testing.T) {
	input := []domain.Result{
		{PlayerName: "A", Score: 35},
		{PlayerName: "B", Score: 5},
		{PlayerName: "C", Score: -40},
	}

	got, err := ValidateResults(input)
	if err != nil {
		t.Fatalf("ValidateResults returned error: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("len(ValidateResults()) = %d, want 3", len(got))
	}
}

func TestValidateResultsTreatsOneBlankPlayerAsThreePlayers(t *testing.T) {
	input := []domain.Result{
		{PlayerName: "A", Score: 35},
		{PlayerName: "", Score: 0},
		{PlayerName: "B", Score: 5},
		{PlayerName: "C", Score: -40},
	}

	got, err := ValidateResults(input)
	if err != nil {
		t.Fatalf("ValidateResults returned error: %v", err)
	}
	if len(got) != 3 || got[1].PlayerName != "B" {
		t.Fatalf("ValidateResults() = %#v, want blank player removed", got)
	}
}

func TestValidateResultsRejectsScoreForBlankPlayer(t *testing.T) {
	input := []domain.Result{
		{PlayerName: "A", Score: 35},
		{PlayerName: "B", Score: 5},
		{PlayerName: "C", Score: -40},
		{PlayerName: "", Score: 10},
	}

	if _, err := ValidateResults(input); err != ErrBlankPlayerScore {
		t.Fatalf("ValidateResults error = %v, want %v", err, ErrBlankPlayerScore)
	}
}

func TestValidateResultsRejectsFewerThanThreePlayers(t *testing.T) {
	input := []domain.Result{
		{PlayerName: "A", Score: 10},
		{PlayerName: "B", Score: -10},
	}

	if _, err := ValidateResults(input); err != ErrPlayerCount {
		t.Fatalf("ValidateResults error = %v, want %v", err, ErrPlayerCount)
	}
}

func TestValidateResultsRejectsNonZeroTotal(t *testing.T) {
	input := []domain.Result{
		{PlayerName: "A", Score: 40},
		{PlayerName: "B", Score: 30},
		{PlayerName: "C", Score: 20},
		{PlayerName: "D", Score: 10},
	}

	if _, err := ValidateResults(input); err != ErrScoreTotal {
		t.Fatalf("ValidateResults error = %v, want %v", err, ErrScoreTotal)
	}
}
