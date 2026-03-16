package review

import (
	"sort"
	"strings"

	"github.com/tomasino/writing-coach/internal/domain"
)

type Comparison struct {
	CurrentID            int64
	BaselineID           int64
	WordDelta            int
	AddedWords           []string
	RemovedWords         []string
	PersistingWeaknesses []string
	AddressedWeaknesses  []string
	Summary              string
}

func CompareSubmissions(current, baseline domain.Submission, currentReview domain.Review, baselineReview domain.Review) Comparison {
	added, removed := diffWordSets(baseline.Content, current.Content)
	persisting, addressed := weaknessDelta(baselineReview.Weaknesses, currentReview.Weaknesses)

	summary := "Revision changes are mixed."
	switch {
	case len(addressed) > len(persisting):
		summary = "Revision appears to address more weaknesses than it preserves."
	case len(persisting) > 0:
		summary = "Revision still carries forward several earlier weaknesses."
	case len(addressed) > 0:
		summary = "Revision appears to have addressed earlier weaknesses."
	}

	return Comparison{
		CurrentID:            current.ID,
		BaselineID:           baseline.ID,
		WordDelta:            current.WordCount - baseline.WordCount,
		AddedWords:           added,
		RemovedWords:         removed,
		PersistingWeaknesses: persisting,
		AddressedWeaknesses:  addressed,
		Summary:              summary,
	}
}

func diffWordSets(before, after string) ([]string, []string) {
	beforeSet := wordSet(before)
	afterSet := wordSet(after)
	var added []string
	var removed []string
	for word := range afterSet {
		if !beforeSet[word] {
			added = append(added, word)
		}
	}
	for word := range beforeSet {
		if !afterSet[word] {
			removed = append(removed, word)
		}
	}
	sort.Strings(added)
	sort.Strings(removed)
	if len(added) > 10 {
		added = added[:10]
	}
	if len(removed) > 10 {
		removed = removed[:10]
	}
	return added, removed
}

func wordSet(text string) map[string]bool {
	set := map[string]bool{}
	for _, word := range strings.Fields(strings.ToLower(text)) {
		word = strings.Trim(word, ".,;:!?\"'()[]{}")
		if len(word) < 4 {
			continue
		}
		set[word] = true
	}
	return set
}

func weaknessDelta(before, after []string) ([]string, []string) {
	var persisting []string
	var addressed []string
	joinedAfter := strings.ToLower(strings.Join(after, " || "))
	for _, weakness := range before {
		key := normalizeWeaknessKey(weakness)
		if key == "" {
			continue
		}
		if strings.Contains(joinedAfter, key) {
			persisting = append(persisting, weakness)
		} else {
			addressed = append(addressed, weakness)
		}
	}
	if len(persisting) > 3 {
		persisting = persisting[:3]
	}
	if len(addressed) > 3 {
		addressed = addressed[:3]
	}
	return persisting, addressed
}

func normalizeWeaknessKey(value string) string {
	value = strings.ToLower(value)
	parts := strings.Fields(value)
	if len(parts) > 6 {
		parts = parts[:6]
	}
	return strings.Join(parts, " ")
}
