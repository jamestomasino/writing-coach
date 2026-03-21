package analyzer

import (
	"fmt"
	"strings"

	"github.com/tomasino/writing-coach/internal/domain"
)

func normalizeWritingLanguage(value string) string {
	return domain.NormalizeWritingLanguage(value)
}

func writingLanguageLabel(value string) string {
	label := strings.TrimSpace(domain.WritingLanguageLabel(value))
	if label == "" {
		return "this language"
	}
	return label
}

func deterministicLanguageSupported(value string) bool {
	return domain.WritingLanguageSupported(value)
}

func languageToolCode(value string) string {
	return domain.WritingLanguageToolCode(value)
}

func unsupportedLanguageWarning(analyzerName, value string) string {
	return fmt.Sprintf("%s skipped: deterministic coaching for %s is not configured yet", analyzerName, writingLanguageLabel(value))
}
