import pathlib
import sys
import unittest

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent))

from app import analyze_text, coref_ambiguity_from_corenlp_payload


class NLPAnalyzerPhase4SignalsTests(unittest.TestCase):
    def test_analyze_text_emits_phase4_metrics(self):
        text = (
            "Teams should update the deployment runbook because the last release failed in production. "
            "For example, the error budget dropped to 20 percent after rollback. "
            "It should include a verification checklist before traffic is shifted."
        )
        payload = {"writing_language": "en", "domain": "technical"}

        out = analyze_text(text, payload)
        metrics = out["metrics"]
        self.assertIn("nlp_claim_count", metrics)
        self.assertIn("nlp_evidence_marker_count", metrics)
        self.assertIn("nlp_claim_evidence_coverage", metrics)
        self.assertIn("nlp_coref_ambiguity_count", metrics)
        self.assertIn("nlp_semantic_repetition_ratio", metrics)
        self.assertIn("nlp_topic_drift_score", metrics)

    def test_claim_evidence_coverage_is_low_when_claims_lack_support(self):
        text = (
            "We should rewrite the API docs. "
            "The team must improve onboarding speed. "
            "Writers need to be more precise and direct."
        )
        out = analyze_text(text, {"writing_language": "en", "domain": "technical"})
        self.assertGreaterEqual(out["metrics"]["nlp_claim_count"], 2)
        self.assertLessEqual(out["metrics"]["nlp_claim_evidence_coverage"], 25)

    def test_repetition_ratio_increases_for_redundant_sentences(self):
        text = (
            "The deployment failed because the migration lock was not released. "
            "The deployment failed because the migration lock was not released. "
            "The deployment failed because the migration lock was not released."
        )
        out = analyze_text(text, {"writing_language": "en", "domain": "technical"})
        self.assertGreaterEqual(out["metrics"]["nlp_semantic_repetition_ratio"], 60)

    def test_topic_drift_increases_when_sections_change_subject(self):
        text = (
            "The API migration requires preflight checks and lock coordination. "
            "Database rollback paths must be tested before cutover.\n\n"
            "The scene opens at dusk with two sisters watching the harbor burn. "
            "Their dialogue reveals fear, guilt, and unresolved rivalry."
        )
        out = analyze_text(text, {"writing_language": "en", "domain": "general"})
        self.assertGreaterEqual(out["metrics"]["nlp_topic_drift_score"], 50)

    def test_coref_ambiguity_detects_unanchored_pronoun_heaviness(self):
        text = (
            "It failed again. "
            "They changed it after that, but it still broke. "
            "This made it worse because they moved it twice."
        )
        out = analyze_text(text, {"writing_language": "en", "domain": "technical"})
        self.assertGreaterEqual(out["metrics"]["nlp_coref_ambiguity_count"], 1)

    def test_corenlp_payload_parser_counts_pronominal_only_chains(self):
        payload = {
            "corefs": {
                "1": [
                    {"text": "the deployment runbook", "type": "NOMINAL", "isRepresentativeMention": True},
                    {"text": "it", "type": "PRONOMINAL", "isRepresentativeMention": False},
                ],
                "2": [
                    {"text": "it", "type": "PRONOMINAL", "isRepresentativeMention": True},
                    {"text": "they", "type": "PRONOMINAL", "isRepresentativeMention": False},
                ],
            }
        }
        self.assertEqual(coref_ambiguity_from_corenlp_payload(payload), 1)


if __name__ == "__main__":
    unittest.main()
