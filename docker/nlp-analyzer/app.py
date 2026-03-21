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


def analyze_text(text):
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

    if len(long_sentences) >= max(2, len(sentences) // 3) and word_count >= 120:
        findings.append(
            {
                "category": "clarity",
                "severity": "warning",
                "message": "Several sentences are carrying a lot of syntactic weight; break a few into cleaner beats.",
            }
        )
    if len(passive_sentences) >= 2:
        findings.append(
            {
                "category": "causal clarity",
                "severity": "warning",
                "message": "Passive constructions appear often enough to blur who is causing the action.",
            }
        )
    if readability_grade >= 11 and word_count >= 150:
        findings.append(
            {
                "category": "readability",
                "severity": "warning",
                "message": "The prose is reading as syntactically demanding; simplify some sentence frames so the dramatic turn stays legible.",
            }
        )
    if dependency_distance >= 3.4 and word_count >= 120:
        findings.append(
            {
                "category": "sentence control",
                "severity": "warning",
                "message": "Dependency distance is elevated, which often means clauses are stretching too far before they resolve.",
            }
        )
    if unique_token_ratio <= 0.42 and word_count >= 120:
        findings.append(
            {
                "category": "diction variety",
                "severity": "note",
                "message": "Lexical variety is low for the draft length; check for repeated filler phrasing or recycled abstractions.",
            }
        )
    if adv_ratio >= 0.09 and word_count >= 80:
        findings.append(
            {
                "category": "prose precision",
                "severity": "note",
                "message": "Adverb usage is high enough to merit a pass for stronger verb choices.",
            }
        )
    if pronoun_count > noun_count and word_count >= 120:
        findings.append(
            {
                "category": "referent clarity",
                "severity": "note",
                "message": "Pronouns outnumber concrete noun references; confirm that every \"he,\" \"she,\" or \"they\" stays unmistakable.",
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
        response = analyze_text(text) if text else {"metrics": {}, "findings": []}

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
