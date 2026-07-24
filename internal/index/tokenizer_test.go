package index

import (
	"reflect"
	"testing"
)


func TestTokenizeLowercases(t *testing.T) {
	tokens := Tokenize("Hello World", 2)
	for _, tok := range tokens {
		for _, r := range tok {
			if r >= 'A' && r <= 'Z' {
				t.Fatalf("Expected lowercase, got %q in tokens %v", tok, tokens)
			}
		}
	}
}

func TestTokenizeStripsPunctuation(t *testing.T) {
	tokens := Tokenize("hello, world! foo-bar.", 2)
	for _, tok := range tokens {
		for _, r := range tok {
			if r == ',' || r == '!' || r == '-' || r == '.' {
				t.Fatalf("Unexpected punctuation in token %q", tok)
			}
		}
	}
}

func TestTokenizeRemovesStopWords(t *testing.T) {
	tokens := Tokenize("the quick brown fox", 2)
	for _, tok := range tokens {
		if stopWords[tok] {
			t.Fatalf("Stop word %q should have been removed", tok)
		}
	}
}

func TestTokenizeRespectsMinLength(t *testing.T) {
	tokens := Tokenize("a go rust python", 4)
	for _, tok := range tokens {
		if len(tok) < 4 {
			t.Fatalf("Token %q is shorter than minLength 4", tok)
		}
	}
}

func TestTokenizeProducesExpectedTokens(t *testing.T) {
	tokens := Tokenize("The quick brown fox jumps", 2)
	want := []string{"quick", "brown", "fox", "jumps"}
	if !reflect.DeepEqual(tokens, want) {
		t.Fatalf("Expected %v, got %v", want, tokens)
	}
}

func TestTokenizeEmptyString(t *testing.T) {
	if tokens := Tokenize("", 2); len(tokens) != 0 {
		t.Fatalf("Expected empty, got %v", tokens)
	}
}
 
func TestTermFrequencies(t *testing.T) {
	freq := TermFrequencies([]string{"go", "rust", "go", "python", "go"})
	if freq["go"] != 3 {
		t.Fatalf("Expected go=3, got %d", freq["go"])
	}
	if freq["rust"] != 1 {
		t.Fatalf("Expected rust=1, got %d", freq["rust"])
	}
}