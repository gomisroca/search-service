package index

import (
	"strings"
	"unicode"
)

// set of english words excluded from indexing
var stopWords = map[string]bool{
	"a": true, "an": true, "and": true, "are": true, "as": true,
	"at": true, "be": true, "been": true, "but": true, "by": true,
	"do": true, "for": true, "from": true, "get": true, "has": true,
	"have": true, "he": true, "her": true, "his": true, "how": true,
	"i": true, "if": true, "in": true, "is": true, "it": true,
	"its": true, "just": true, "me": true, "my": true, "no": true,
	"not": true, "of": true, "on": true, "or": true, "our": true,
	"out": true, "she": true, "so": true, "than": true, "that": true,
	"the": true, "their": true, "them": true, "then": true, "there": true,
	"they": true, "this": true, "to": true, "up": true, "us": true,
	"was": true, "we": true, "were": true, "what": true, "when": true,
	"which": true, "who": true, "will": true, "with": true, "you": true,
	"your": true,
}

// convert text into a normalised slice of tokens suitable for indexing
func Tokenize(text string, minLength int) []string {
	// replace punctuation with spaces 
	var b strings.Builder
	for _, r := range text {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(unicode.ToLower(r))
		} else {
			b.WriteRune(' ')
		}
	}

	raw := strings.Fields(b.String())
	out := make([]string, 0, len(raw))
	for _, tok := range raw {
		if len(tok) < minLength {
			continue
		}
		if stopWords[tok] {
			continue
		}
		out = append(out, tok)
	}
	return out
}

// returns a map of token to count within text
// used at indexing time to determine term frequency
func TermFrequencies(tokens []string) map[string]int {
	freq := make(map[string]int, len(tokens))
	for _, t := range tokens {
		freq[t]++
	}
	return freq
}