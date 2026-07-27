package index

import (
	"fmt"
	"search-service/internal/models"
	"sync"
	"testing"
	"time"
)

func makeDoc(id, content string) models.Document {
	return models.Document{
		ID: id,
		Content: content,
		IndexedAt: time.Now(),
	}
}

func TestAddAndGet(t *testing.T) {
	idx := New(2)
	idx.Add(makeDoc("doc1", "the quick brown fox"))
	doc, ok := idx.Get("doc1")
	if !ok || doc.ID != "doc1" {
		t.Fatal("Expected to find indexed document")
	}
}

func TestGetMissing(t *testing.T) {
	idx := New(2)
	_, ok := idx.Get("missing")
	if ok {
		t.Fatal("Expected not found for missing document")
	}
}

func TestLen(t *testing.T) {
	idx := New(2)
	idx.Add(makeDoc("doc1", "hello world"))
	idx.Add(makeDoc("doc2", "foo bar"))
	if idx.Len() != 2 {
		t.Fatalf("Expected 2 documents, got %d", idx.Len())
	}
}

func TestDeleteRemovesDocument(t *testing.T) {
	idx := New(2)
	idx.Add(makeDoc("doc1", "hello world"))
	if !idx.Delete("doc1") {
		t.Fatal("Expected Delete to return true")
	}
	if idx.Delete("doc1") {
		t.Fatal("Expected second Delete to return false")
	}
	_, ok := idx.Get("doc1")
	if ok {
		t.Fatal("Expected deleted document to be gone")
	}
}

func TestDeleteRemovesFromInvertedIndex(t *testing.T) {
	idx := New(2)
	idx.Add(makeDoc("doc1", "golang microservice"))
	idx.Delete("doc1")
	// After deletion, searching for the term should return nothing
	results := idx.Search("golang", 10)
	if len(results) != 0 {
		t.Fatalf("Expected no results after deletion, got %d", len(results))
	}
}

func TestReplaceOnDuplicateID(t *testing.T) {
	idx := New(2)
	idx.Add(makeDoc("doc1", "golang microservice"))
	idx.Add(makeDoc("doc1", "python fastapi"))
 
	if idx.Len() != 1 {
		t.Fatalf("Expected 1 document after replacement, got %d", idx.Len())
	}
	// Old content should not be searchable
	if results := idx.Search("golang", 10); len(results) != 0 {
		t.Fatal("Expected old content to be gone after replacement")
	}
	// New content should be searchable
	if results := idx.Search("python", 10); len(results) != 1 {
		t.Fatal("Expected new content to be searchable")
	}
}

func TestSearchReturnsMatchingDocuments(t *testing.T) {
	idx := New(2)
	idx.Add(makeDoc("doc1", "golang concurrency goroutines"))
	idx.Add(makeDoc("doc2", "python asyncio coroutines"))
	idx.Add(makeDoc("doc3", "rust ownership memory"))
 
	results := idx.Search("golang", 10)
	if len(results) != 1 || results[0].Document.ID != "doc1" {
		t.Fatalf("Expected doc1, got %v", results)
	}
}
 
func TestSearchANDSemantics(t *testing.T) {
	idx := New(2)
	idx.Add(makeDoc("doc1", "golang microservice docker"))
	idx.Add(makeDoc("doc2", "golang python comparison"))
	idx.Add(makeDoc("doc3", "docker kubernetes deployment"))
 
	// Only doc1 contains both "golang" AND "docker"
	results := idx.Search("golang docker", 10)
	if len(results) != 1 || results[0].Document.ID != "doc1" {
		t.Fatalf("Expected only doc1 for AND query, got %v", results)
	}
}
 
func TestSearchNoResults(t *testing.T) {
	idx := New(2)
	idx.Add(makeDoc("doc1", "hello world"))
	results := idx.Search("zzzzz", 10)
	if len(results) != 0 {
		t.Fatalf("Expected no results for non-existent term")
	}
}
 
func TestSearchEmptyQuery(t *testing.T) {
	idx := New(2)
	idx.Add(makeDoc("doc1", "hello world"))
	results := idx.Search("", 10)
	if results != nil {
		t.Fatal("Expected nil for empty query")
	}
}
 
func TestSearchLimitRespected(t *testing.T) {
	idx := New(2)
	for i := 0; i < 20; i++ {
		idx.Add(makeDoc(fmt.Sprintf("doc%d", i), "golang microservice"))
	}
	results := idx.Search("golang", 5)
	if len(results) != 5 {
		t.Fatalf("Expected 5 results (limit), got %d", len(results))
	}
}
 
func TestSearchRankingRareTermScoresHigher(t *testing.T) {
	idx := New(2)
	// "microservice" appears in both; "goroutine" appears only in doc1.
	// doc1 should rank higher for query "goroutine microservice".
	idx.Add(makeDoc("doc1", "golang goroutine microservice concurrency"))
	idx.Add(makeDoc("doc2", "python microservice flask"))
 
	results := idx.Search("goroutine microservice", 10)
	if len(results) != 1 {
		t.Fatalf("Expected 1 result (AND semantics — only doc1 has 'goroutine'), got %d", len(results))
	}
	if results[0].Document.ID != "doc1" {
		t.Fatalf("Expected doc1 to rank first")
	}
}
 
func TestSearchScoresArePositive(t *testing.T) {
	idx := New(2)
	idx.Add(makeDoc("doc1", "golang concurrency microservice"))
	results := idx.Search("golang", 10)
	for _, r := range results {
		if r.Score <= 0 {
			t.Fatalf("Expected positive score, got %f for %s", r.Score, r.Document.ID)
		}
	}
}
 
// hammers the index with concurrent reads and writes
func TestConcurrentReadWrite(t *testing.T) {
	idx := New(2)
	const writers = 10
	const readers = 20
	const ops = 50
 
	var wg sync.WaitGroup
 
	for w := 0; w < writers; w++ {
		wg.Add(1)
		go func(w int) {
			defer wg.Done()
			for i := 0; i < ops; i++ {
				id := fmt.Sprintf("doc-%d-%d", w, i)
				idx.Add(makeDoc(id, "golang microservice search index"))
				if i%5 == 0 {
					idx.Delete(id)
				}
			}
		}(w)
	}
 
	for r := 0; r < readers; r++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < ops; i++ {
				idx.Search("golang microservice", 10)
				idx.Len()
			}
		}()
	}
 
	wg.Wait()
}
