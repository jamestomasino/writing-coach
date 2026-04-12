package review

import _ "embed"

//go:embed testdata/objective_eval_corpus.json
var defaultObjectiveEvalCorpus []byte

func DefaultObjectiveEvalCorpus() []byte {
	if len(defaultObjectiveEvalCorpus) == 0 {
		return nil
	}
	buf := make([]byte, len(defaultObjectiveEvalCorpus))
	copy(buf, defaultObjectiveEvalCorpus)
	return buf
}
