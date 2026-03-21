package domain

import "testing"

func TestNormalizeWritingLanguageDefaultsToEnglish(t *testing.T) {
	cases := []string{"", "en", "EN-US", "english"}
	for _, value := range cases {
		if got := NormalizeWritingLanguage(value); got != "en" {
			t.Fatalf("NormalizeWritingLanguage(%q) = %q", value, got)
		}
	}
}

func TestWritingLanguageSupportIsExplicit(t *testing.T) {
	if !WritingLanguageSupported("en") {
		t.Fatal("expected english to be supported")
	}
	if WritingLanguageSupported("es") {
		t.Fatal("expected spanish to remain unsupported until explicitly added")
	}
}
