package service

import (
	"errors"
	"path"
	"regexp"
	"strings"

	"github.com/tkmbr/mahjong-score/internal/domain"
)

const RuleRepository = "tkmbr/mahjong-rule"

var commitRevisionPattern = regexp.MustCompile(`^[0-9a-f]{40}$`)

func ValidateRuleCitation(input *domain.RuleCitation) (*domain.RuleCitation, error) {
	if input == nil {
		return nil, nil
	}

	citation := &domain.RuleCitation{
		Revision: strings.TrimSpace(input.Revision),
		Path:     strings.TrimSpace(input.Path),
		Title:    strings.TrimSpace(input.Title),
	}
	if citation.Revision == "" && citation.Path == "" && citation.Title == "" {
		return nil, nil
	}
	if citation.Revision == "" || citation.Path == "" || citation.Title == "" {
		return nil, errors.New("ルールの引用情報が不足しています")
	}
	if !commitRevisionPattern.MatchString(citation.Revision) {
		return nil, errors.New("ルールのコミットSHAが正しくありません")
	}
	if len(citation.Path) > 500 || path.IsAbs(citation.Path) || citation.Path == ".." || strings.HasPrefix(citation.Path, "../") || path.Clean(citation.Path) != citation.Path || strings.Contains(citation.Path, `\`) || !strings.HasSuffix(strings.ToLower(citation.Path), ".pdf") {
		return nil, errors.New("ルールのPDFパスが正しくありません")
	}
	if len(citation.Title) > 100 {
		return nil, errors.New("ルール名が長すぎます")
	}
	return citation, nil
}
