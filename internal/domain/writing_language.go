package domain

import "strings"

const DefaultWritingLanguage = "en"

func NormalizeWritingLanguage(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "en", "en-us", "en-gb", "english":
		return DefaultWritingLanguage
	default:
		return strings.ToLower(strings.TrimSpace(value))
	}
}

func WritingLanguageLabel(code string) string {
	switch NormalizeWritingLanguage(code) {
	case "en":
		return "English"
	default:
		return strings.TrimSpace(code)
	}
}

func WritingLanguageSupported(code string) bool {
	switch NormalizeWritingLanguage(code) {
	case "en":
		return true
	default:
		return false
	}
}

func WritingLanguageToolCode(code string) string {
	switch NormalizeWritingLanguage(code) {
	case "en":
		return "en-US"
	default:
		return ""
	}
}
