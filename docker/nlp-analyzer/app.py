import json
import math
import os
import re
import urllib.error
import urllib.parse
import urllib.request
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer

import spacy
import textdescriptives as td


NLP = spacy.load(os.getenv("SPACY_MODEL", "en_core_web_sm"))
NLP.add_pipe("textdescriptives/descriptive_stats")
NLP.add_pipe("textdescriptives/readability")
NLP.add_pipe("textdescriptives/dependency_distance")
NLP.add_pipe("textdescriptives/pos_proportions")

CLAIM_MARKERS = (
    "should",
    "must",
    "need to",
    "recommend",
    "therefore",
    "argue",
    "suggest",
)

EVIDENCE_MARKERS = (
    "because",
    "for example",
    "for instance",
    "according to",
    "data",
    "evidence",
    "study",
    "result",
    "as shown",
)

AMBIGUOUS_PRONOUNS = {"it", "this", "that", "they", "these", "those"}
STOPWORDS = {
    "the",
    "a",
    "an",
    "and",
    "or",
    "but",
    "if",
    "then",
    "than",
    "to",
    "of",
    "in",
    "on",
    "for",
    "with",
    "at",
    "by",
    "from",
    "as",
    "is",
    "are",
    "was",
    "were",
    "be",
    "been",
    "being",
    "it",
    "this",
    "that",
    "they",
}


def safe_number(value, default=0.0):
    if value is None:
        return default
    if isinstance(value, float) and math.isnan(value):
        return default
    return float(value)


def clamp_metric(value):
    return max(0, int(round(value)))


def sentence_texts(doc):
    return [sent.text.strip() for sent in doc.sents if sent.text.strip()]


def normalized_word_set(text):
    words = re.findall(r"[a-zA-Z][a-zA-Z'-]+", text.lower())
    return {w for w in words if len(w) >= 4 and w not in STOPWORDS}


def bounded_percent(numerator, denominator):
    if denominator <= 0:
        return 0
    return clamp_metric(max(0.0, min(100.0, 100.0 * float(numerator) / float(denominator))))


def claim_evidence_metrics(sentences):
    claims = 0
    evidence_markers = 0
    for sentence in sentences:
        lowered = sentence.lower()
        if any(marker in lowered for marker in CLAIM_MARKERS):
            claims += 1
        if any(marker in lowered for marker in EVIDENCE_MARKERS):
            evidence_markers += 1
    if claims == 0:
        coverage = 100
    else:
        coverage = bounded_percent(min(evidence_markers, claims), claims)
    return claims, evidence_markers, coverage


def coref_ambiguity_count(doc):
    recent_nouns = 0
    ambiguous = 0
    for token in doc:
        if token.pos_ in {"NOUN", "PROPN"}:
            recent_nouns = 2
            continue
        if token.pos_ == "PRON" and token.lower_ in AMBIGUOUS_PRONOUNS:
            if recent_nouns == 0:
                ambiguous += 1
            else:
                recent_nouns -= 1
        elif token.is_sent_start:
            recent_nouns = max(0, recent_nouns - 1)
    return ambiguous


def sentence_jaccard_similarity(left, right):
    left_set = normalized_word_set(left)
    right_set = normalized_word_set(right)
    if not left_set or not right_set:
        return 0.0
    intersection = len(left_set.intersection(right_set))
    union = len(left_set.union(right_set))
    if union == 0:
        return 0.0
    return float(intersection) / float(union)


def semantic_repetition_ratio(sentences):
    if len(sentences) < 2:
        return 0
    similar_pairs = 0
    total_pairs = 0
    for i in range(len(sentences)):
        for j in range(i + 1, len(sentences)):
            total_pairs += 1
            if sentence_jaccard_similarity(sentences[i], sentences[j]) >= 0.65:
                similar_pairs += 1
    return bounded_percent(similar_pairs, total_pairs)


def topic_drift_score(text):
    paragraphs = [p.strip() for p in text.split("\n\n") if p.strip()]
    if len(paragraphs) < 2:
        return 0
    anchor = normalized_word_set(paragraphs[0])
    if not anchor:
        return 0
    overlap_sum = 0.0
    compared = 0
    for paragraph in paragraphs[1:]:
        compared_set = normalized_word_set(paragraph)
        if not compared_set:
            continue
        intersection = len(anchor.intersection(compared_set))
        union = len(anchor.union(compared_set))
        if union == 0:
            continue
        overlap_sum += float(intersection) / float(union)
        compared += 1
    if compared == 0:
        return 0
    overlap_avg = overlap_sum / float(compared)
    return clamp_metric((1.0 - overlap_avg) * 100.0)


def coref_ambiguity_from_corenlp_payload(payload):
    corefs = payload.get("corefs", {}) if isinstance(payload, dict) else {}
    if not isinstance(corefs, dict):
        return 0
    ambiguous = 0
    for _, chain in corefs.items():
        if not isinstance(chain, list) or not chain:
            continue
        has_non_pronominal = False
        for mention in chain:
            if not isinstance(mention, dict):
                continue
            mention_type = str(mention.get("type", "")).strip().upper()
            if mention_type and mention_type != "PRONOMINAL":
                has_non_pronominal = True
                break
        if not has_non_pronominal:
            ambiguous += 1
    return ambiguous


def corenlp_ambiguity_count(text):
    base_url = os.getenv("CORENLP_URL", "").strip()
    if not base_url:
        return 0
    params = {
        "properties": json.dumps(
            {
                "annotators": "tokenize,ssplit,pos,lemma,ner,parse,coref",
                "outputFormat": "json",
                "timeout": 10000,
            }
        )
    }
    endpoint = base_url.rstrip("/") + "/?" + urllib.parse.urlencode(params)
    request = urllib.request.Request(
        endpoint,
        data=text.encode("utf-8"),
        headers={"Content-Type": "text/plain; charset=utf-8"},
        method="POST",
    )
    try:
        with urllib.request.urlopen(request, timeout=2.0) as response:
            payload = json.loads(response.read().decode("utf-8"))
            return coref_ambiguity_from_corenlp_payload(payload)
    except (TimeoutError, urllib.error.URLError, urllib.error.HTTPError, json.JSONDecodeError):
        return 0


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
        "claim_evidence_gap": {
            "technical": "Several claims are not directly supported by explicit evidence markers; add rationale, data points, or concrete examples near each claim.",
            "academic": "Several claims are not directly supported by explicit evidence markers; anchor key assertions with concrete support.",
            "professional": "Several recommendations are not directly supported by explicit rationale or evidence language; make support visible near each ask.",
            "marketing": "Several value claims are not directly supported by proof markers; add concrete supporting evidence or examples.",
            "thought_leadership": "Several central claims are not directly supported by explicit evidence markers; add concrete support near key assertions.",
            "default": "Several claims are not directly supported by explicit evidence markers; add concrete support near key assertions.",
        },
        "semantic_repetition": {
            "default": "Repeated semantic phrasing is high; condense or vary repeated lines so each sentence adds new information.",
        },
        "topic_drift": {
            "default": "Topic continuity shifts sharply between sections; strengthen transitions or tighten focus around one governing objective.",
        },
    }
    variants = messages.get(key, {})
    return variants.get(domain, variants.get("default", ""))


def analyze_text(text, payload):
    if str(payload.get("writing_language", "en")).strip().lower() not in {"", "en", "en-us", "english"}:
        return {
            "metrics": {},
            "findings": [],
            "warnings": ["nlp skipped: deterministic coaching for this language is not configured yet"],
        }
    domain = normalized_domain(payload)
    doc = NLP(text)
    sentences_text = sentence_texts(doc)
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
    claim_count, evidence_marker_count, claim_evidence_coverage = claim_evidence_metrics(sentences_text)
    coref_ambiguity = coref_ambiguity_count(doc)
    corenlp_coref_ambiguity = corenlp_ambiguity_count(text)
    if corenlp_coref_ambiguity > coref_ambiguity:
        coref_ambiguity = corenlp_coref_ambiguity
    repetition_ratio = semantic_repetition_ratio(sentences_text)
    drift_score = topic_drift_score(text)

    metrics_payload = {
        "nlp_long_sentences": len(long_sentences),
        "nlp_passive_sentences": len(passive_sentences),
        "nlp_pronouns": pronoun_count,
        "nlp_nouns": noun_count,
        "nlp_readability_grade": clamp_metric(safe_number(metrics.get("flesch_kincaid_grade"))),
        "nlp_dependency_distance": clamp_metric(safe_number(metrics.get("dependency_distance_mean")) * 100),
        "nlp_unique_token_ratio": clamp_metric(safe_number(metrics.get("proportion_unique_tokens")) * 100),
        "nlp_claim_count": claim_count,
        "nlp_evidence_marker_count": evidence_marker_count,
        "nlp_claim_evidence_coverage": claim_evidence_coverage,
        "nlp_coref_ambiguity_count": coref_ambiguity,
        "nlp_semantic_repetition_ratio": repetition_ratio,
        "nlp_topic_drift_score": drift_score,
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
    if claim_count >= 2 and claim_evidence_coverage < 50:
        findings.append(
            {
                "category": "evidence integration",
                "severity": "warning",
                "message": message_for(domain, "claim_evidence_gap"),
            }
        )
    if repetition_ratio >= 55 and word_count >= 80:
        findings.append(
            {
                "category": "sentence economy",
                "severity": "note",
                "message": message_for(domain, "semantic_repetition"),
            }
        )
    if drift_score >= 50 and word_count >= 120:
        findings.append(
            {
                "category": "structural signposting",
                "severity": "note",
                "message": message_for(domain, "topic_drift"),
            }
        )

    return {"metrics": metrics_payload, "findings": findings, "warnings": []}


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
