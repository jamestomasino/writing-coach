import json
import math
import os
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer

import spacy
import textdescriptives as td


NLP = spacy.load(os.getenv("SPACY_MODEL", "en_core_web_sm"))
NLP.add_pipe("textdescriptives/descriptive_stats")
NLP.add_pipe("textdescriptives/readability")
NLP.add_pipe("textdescriptives/dependency_distance")
NLP.add_pipe("textdescriptives/pos_proportions")


def safe_number(value, default=0.0):
    if value is None:
        return default
    if isinstance(value, float) and math.isnan(value):
        return default
    return float(value)


def clamp_metric(value):
    return max(0, int(round(value)))


def normalized_domain(payload):
    domain = str(payload.get("domain", "")).strip().lower()
    if domain:
        return domain

    combined = " ".join(
        str(payload.get(key, "")).strip().lower()
        for key in ("writing_type", "assignment_format", "template_key", "tree_slug")
    )
    if "fantasy" in combined:
        return "fantasy"
    if "fiction" in combined or "scene" in combined or "short story" in combined or "memoir" in combined:
        return "fiction"
    if "technical" in combined or "documentation" in combined or "how-to" in combined or "guide" in combined:
        return "technical"
    if "academic" in combined or "essay" in combined or "research" in combined:
        return "academic"
    if "marketing" in combined or "landing page" in combined:
        return "marketing"
    if "thought leadership" in combined or "journalism" in combined or "reporting" in combined:
        return "thought_leadership"
    if "professional" in combined or "memo" in combined or "email" in combined or "grant" in combined:
        return "professional"
    return "general"


def message_for(domain, key):
    messages = {
        "clarity_long_sentences": {
            "technical": "Several sentences are carrying too much syntactic load for instructional writing; break a few into cleaner steps or clearer explanations.",
            "academic": "Several sentences are carrying too much syntactic load; simplify a few so the argument is easier to follow.",
            "professional": "Several sentences are carrying too much syntactic load; shorten them so the key action or request is clearer.",
            "marketing": "Several sentences are carrying too much syntactic load; tighten them so the message lands faster.",
            "thought_leadership": "Several sentences are carrying too much syntactic load; tighten them so the idea progression stays clear.",
            "default": "Several sentences are carrying a lot of syntactic weight; break a few into cleaner units.",
        },
        "passive_voice": {
            "technical": "Passive constructions appear often enough to blur who should act or what produces each result.",
            "academic": "Passive constructions appear often enough to blur agency; clarify who is making each claim or producing each result.",
            "professional": "Passive constructions appear often enough to blur ownership; make it clearer who is responsible for what.",
            "marketing": "Passive constructions appear often enough to weaken momentum; make key actions sound more direct.",
            "thought_leadership": "Passive constructions appear often enough to soften the force of the claims; clarify who is doing what.",
            "default": "Passive constructions appear often enough to blur who is causing the action.",
        },
        "readability": {
            "technical": "The prose is reading as syntactically demanding for instructional writing; simplify some lines so the reader can act without rereading.",
            "academic": "The prose is reading as syntactically demanding; simplify some sentence frames so the argument stays legible.",
            "professional": "The prose is reading as syntactically demanding; simplify some lines so the main decision or request is unmistakable.",
            "marketing": "The prose is reading as syntactically demanding; simplify some lines so the value lands on first pass.",
            "thought_leadership": "The prose is reading as syntactically demanding; simplify some lines so the central idea remains clear.",
            "default": "The prose is reading as syntactically demanding; simplify some sentence frames so the central movement stays legible.",
        },
        "dependency_distance": {
            "technical": "Dependency distance is elevated, which often means instructions or explanations are stretching too far before they resolve.",
            "academic": "Dependency distance is elevated, which often means claims and qualifiers are stretching too far before they resolve.",
            "professional": "Dependency distance is elevated, which often means key actions are delayed by too much setup.",
            "marketing": "Dependency distance is elevated, which often means the main point arrives later than it should.",
            "thought_leadership": "Dependency distance is elevated, which often means ideas are stretching too far before they resolve.",
            "default": "Dependency distance is elevated, which often means clauses are stretching too far before they resolve.",
        },
        "low_variety": {
            "technical": "Lexical variety is low for the draft length; check for repeated filler phrasing or repeated setup language.",
            "academic": "Lexical variety is low for the draft length; check for repeated abstractions or recycled transitions.",
            "professional": "Lexical variety is low for the draft length; check for repeated filler phrasing or repeated framing.",
            "marketing": "Lexical variety is low for the draft length; check for repeated buzzwords or recycled value language.",
            "thought_leadership": "Lexical variety is low for the draft length; check for repeated abstractions or repeated framing phrases.",
            "default": "Lexical variety is low for the draft length; check for repeated filler phrasing or recycled abstractions.",
        },
        "adv_ratio": {
            "technical": "Modifier usage is high enough to merit a pass for leaner, more direct phrasing.",
            "academic": "Modifier usage is high enough to merit a pass for sharper, more precise wording.",
            "professional": "Modifier usage is high enough to merit a pass for more direct wording.",
            "marketing": "Modifier usage is high enough to merit a pass for stronger, more specific copy.",
            "thought_leadership": "Modifier usage is high enough to merit a pass for sharper wording.",
            "default": "Adverb usage is high enough to merit a pass for stronger verb choices.",
        },
        "pronoun_ratio": {
            "technical": "Pronouns outnumber concrete noun references; confirm that each step still makes the actor and object explicit.",
            "academic": "Pronouns outnumber concrete noun references; confirm that each claim still has a clear referent.",
            "professional": "Pronouns outnumber concrete noun references; confirm that ownership and referents stay unmistakable.",
            "marketing": "Pronouns outnumber concrete noun references; confirm that the product, reader, or offer stays unmistakable.",
            "thought_leadership": "Pronouns outnumber concrete noun references; confirm that every referent stays unmistakable.",
            "default": "Pronouns outnumber concrete noun references; confirm that each pronoun stays unmistakable.",
        },
    }
    variants = messages.get(key, {})
    return variants.get(domain, variants.get("default", ""))


def analyze_text(text, payload):
    domain = normalized_domain(payload)
    doc = NLP(text)
    metrics = td.extract_dict(doc)[0]
    sentences = list(doc.sents)
    sentence_lengths = [sum(1 for token in sentence if not token.is_space and not token.is_punct) for sentence in sentences]
    long_sentences = [length for length in sentence_lengths if length >= 28]
    passive_sentences = [
        sentence
        for sentence in sentences
        if any(token.dep_.endswith("pass") for token in sentence)
    ]

    noun_count = sum(1 for token in doc if token.pos_ in {"NOUN", "PROPN"})
    pronoun_count = sum(1 for token in doc if token.pos_ == "PRON")

    metrics_payload = {
        "nlp_long_sentences": len(long_sentences),
        "nlp_passive_sentences": len(passive_sentences),
        "nlp_pronouns": pronoun_count,
        "nlp_nouns": noun_count,
        "nlp_readability_grade": clamp_metric(safe_number(metrics.get("flesch_kincaid_grade"))),
        "nlp_dependency_distance": clamp_metric(safe_number(metrics.get("dependency_distance_mean")) * 100),
        "nlp_unique_token_ratio": clamp_metric(safe_number(metrics.get("proportion_unique_tokens")) * 100),
    }

    findings = []
    word_count = clamp_metric(metrics.get("n_tokens", 0))
    readability_grade = safe_number(metrics.get("flesch_kincaid_grade"))
    dependency_distance = safe_number(metrics.get("dependency_distance_mean"))
    unique_token_ratio = safe_number(metrics.get("proportion_unique_tokens"))
    adv_ratio = safe_number(metrics.get("pos_prop_ADV"))

    long_sentence_threshold = max(2, len(sentences) // 3)
    if domain in {"technical", "marketing", "professional"}:
        long_sentence_threshold = max(2, len(sentences) // 4)
    if len(long_sentences) >= long_sentence_threshold and word_count >= 120:
        findings.append(
            {
                "category": "clarity",
                "severity": "warning",
                "message": message_for(domain, "clarity_long_sentences"),
            }
        )
    if len(passive_sentences) >= 2:
        findings.append(
            {
                "category": "causal clarity",
                "severity": "warning",
                "message": message_for(domain, "passive_voice"),
            }
        )
    readability_threshold = 11
    if domain in {"technical", "marketing", "professional"}:
        readability_threshold = 10
    if readability_grade >= readability_threshold and word_count >= 150:
        findings.append(
            {
                "category": "readability",
                "severity": "warning",
                "message": message_for(domain, "readability"),
            }
        )
    dependency_threshold = 3.4
    if domain in {"technical", "marketing", "professional"}:
        dependency_threshold = 3.2
    if dependency_distance >= dependency_threshold and word_count >= 120:
        findings.append(
            {
                "category": "sentence control",
                "severity": "warning",
                "message": message_for(domain, "dependency_distance"),
            }
        )
    if unique_token_ratio <= 0.42 and word_count >= 120:
        findings.append(
            {
                "category": "diction variety",
                "severity": "note",
                "message": message_for(domain, "low_variety"),
            }
        )
    if adv_ratio >= 0.09 and word_count >= 80:
        findings.append(
            {
                "category": "prose precision",
                "severity": "note",
                "message": message_for(domain, "adv_ratio"),
            }
        )
    if pronoun_count > noun_count and word_count >= 120:
        findings.append(
            {
                "category": "referent clarity",
                "severity": "note",
                "message": message_for(domain, "pronoun_ratio"),
            }
        )

    return {"metrics": metrics_payload, "findings": findings}


class Handler(BaseHTTPRequestHandler):
    def do_GET(self):
        if self.path != "/health":
            self.send_response(404)
            self.end_headers()
            return
        self.send_response(200)
        self.send_header("Content-Type", "application/json")
        self.end_headers()
        self.wfile.write(b'{"ok":true}')

    def do_POST(self):
        if self.path != "/analyze":
            self.send_response(404)
            self.end_headers()
            return

        length = int(self.headers.get("Content-Length", "0"))
        raw = self.rfile.read(length)
        try:
            payload = json.loads(raw)
        except json.JSONDecodeError:
            self.send_response(400)
            self.end_headers()
            return

        text = str(payload.get("text", "")).strip()
        response = analyze_text(text, payload) if text else {"metrics": {}, "findings": []}

        body = json.dumps(response).encode("utf-8")
        self.send_response(200)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        self.wfile.write(body)

    def log_message(self, format, *args):
        return


if __name__ == "__main__":
    port = int(os.getenv("PORT", "8020"))
    server = ThreadingHTTPServer(("0.0.0.0", port), Handler)
    server.serve_forever()
