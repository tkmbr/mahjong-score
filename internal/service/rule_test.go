package service

import (
	"testing"

	"github.com/tkmbr/mahjong-score/internal/domain"
)

func TestValidateRuleCitation(t *testing.T) {
	input := &domain.RuleCitation{
		Revision: " 8baf4cb3b3861d89badfd53424a6d5cce73e904e ",
		Path:     " 5等サンマ/rule.pdf ",
		Title:    " 5等サンマ ",
	}

	got, err := ValidateRuleCitation(input)
	if err != nil {
		t.Fatal(err)
	}
	if got.Title != "5等サンマ" || got.Path != "5等サンマ/rule.pdf" {
		t.Fatalf("ValidateRuleCitation() = %#v", got)
	}
}

func TestValidateRuleCitationAllowsNoRule(t *testing.T) {
	got, err := ValidateRuleCitation(nil)
	if err != nil || got != nil {
		t.Fatalf("ValidateRuleCitation(nil) = (%#v, %v), want (nil, nil)", got, err)
	}
}

func TestValidateRuleCitationRejectsInvalidReference(t *testing.T) {
	tests := []domain.RuleCitation{
		{Revision: "main", Path: "5等サンマ/rule.pdf", Title: "5等サンマ"},
		{Revision: "8baf4cb3b3861d89badfd53424a6d5cce73e904e", Path: "../rule.pdf", Title: "5等サンマ"},
		{Revision: "8baf4cb3b3861d89badfd53424a6d5cce73e904e", Path: "5等サンマ/rule.typ", Title: "5等サンマ"},
	}
	for _, input := range tests {
		if _, err := ValidateRuleCitation(&input); err == nil {
			t.Fatalf("ValidateRuleCitation(%#v) succeeded", input)
		}
	}
}
