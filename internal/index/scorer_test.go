package index

import (
	"math"
	"testing"
)

func approxEqual(a, b, epsilon float64) bool {
	return math.Abs(a-b) < epsilon
}

func TestTFZeroWhenNoTokens(t *testing.T) {
	if TF(5, 0) != 0 {
		t.Fatal("Expected 0 when total tokens is 0")
	}
}

func TestTFProportional(t *testing.T) {
	if !approxEqual(TF(3, 10), 0.3, 1e-9) {
		t.Fatalf("Expected 0.3, got %f", TF(3, 10))
	}
}

func TestIDFZeroWhenNoDocuments(t *testing.T) {
	if IDF(0, 0) != 0 {
		t.Fatal("Expected 0 when doc count is 0")
	}
}

func TestIDFHighWhenTermIsRare(t *testing.T) {
	rareIDF := IDF(1000, 1)
	commonIDF := IDF(1000, 500)
	if rareIDF <= commonIDF {
		t.Fatalf("Expected rare term IDF (%f) > common term IDF (%f)", rareIDF, commonIDF)
	}
}

func TestIDFSmoothedNeverNegative(t *testing.T) {
	if IDF(100, 100) < 0 {
		t.Fatal("Expected non-negative IDF even when term is in all documents")
	}
}

func TestTFIDFCombinesCorrectly(t *testing.T) {
	score := TFIDF(3, 10, 1000, 1)
	expected := TF(3, 10) * IDF(1000, 1)
	if !approxEqual(score, expected, 1e-9) {
		t.Fatalf("TFIDF did not combine TF and IDF correctly")
	}
}