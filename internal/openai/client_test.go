package openai

import "testing"

func TestNormalizeReview(t *testing.T) {
	value := reviewResponse{
		Summary:    "  a  summary \n",
		Strengths:  []string{" one  strength ", ""},
		Weaknesses: []string{" weak \t point "},
		NextFocus:  " prose precision ",
		SkillScores: []skillScore{
			{Skill: " prose precision ", Score: 3},
		},
	}

	got := normalizeReview(value)
	if got.Summary != "a summary" {
		t.Fatalf("summary = %q", got.Summary)
	}
	if len(got.Strengths) != 1 || got.Strengths[0] != "one strength" {
		t.Fatalf("strengths = %#v", got.Strengths)
	}
	if got.NextFocus != "prose precision" {
		t.Fatalf("next focus = %q", got.NextFocus)
	}
	if got.SkillScores[0].Skill != "prose precision" {
		t.Fatalf("skill = %q", got.SkillScores[0].Skill)
	}
}
