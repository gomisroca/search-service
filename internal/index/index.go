// Implements in-memory inverted index with TF-IDF scoring.
//
// Reads run with RLock for concurrency. Index and Delete take the full lock.

package index

import (
	"search-service/internal/models"
	"sort"
	"sync"
	"time"
)

// records one document's revelancy for a given term
type posting struct {
	docID 		string
	termCount 	int 	// how many times the term appears in the document
}

// holds the data needed at query time for TF-IDF, without rescanning the document content.
type docStats struct {
	doc 		models.Document
	tokenCount 	int 			// total tokens in document (TF denominator)
	termCounts 	map[string]int 	// term > count in the document
}

//  1. invertedIndex: term > []posting
//     Fast lookup: "which documents contain this term?"
//
//  2. docs: docID > docStats
//     Fast retrieval: "what are this document's token counts for scoring?"
type Index struct {
	mu 				sync.RWMutex
	invertedIndex 	map[string][]posting
	docs 			map[string]docStats
	minTokenLen 	int
}

func New(minTokenLen int) *Index {
	return &Index{
		invertedIndex: make(map[string][]posting),
		docs:          make(map[string]docStats),
		minTokenLen:   minTokenLen,
	}
}

func (idx *Index) Add(doc models.Document) {
	tokens := Tokenize(doc.Content, idx.minTokenLen)
	termCounts := TermFrequencies(tokens)

	idx.mu.Lock()
	defer idx.mu.Unlock()

	// If the document already exists, remove its old postings first
	if old, exists := idx.docs[doc.ID]; exists {
		idx.removePostings(doc.ID, old.termCounts)
	}

	// Add the new postings
	for term, count := range termCounts {
		idx.invertedIndex[term] = append(idx.invertedIndex[term], posting{
			docID: doc.ID,
			termCount: count,
		})
	}

	idx.docs[doc.ID] = docStats{
		doc: doc,
		tokenCount: len(tokens),
		termCounts: termCounts,
	}
}

// Removes a document from the index. Return false if not found.
func (idx *Index) Delete(docID string) bool {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	stats, exists := idx.docs[docID]
	if !exists {
		return false
	}

	idx.removePostings(docID, stats.termCounts)
	delete(idx.docs, docID)
	return true
}

// Remove a document's postings from the inverted index.
func (idx *Index) removePostings(docID string, termCounts map[string]int) {
	for term := range termCounts {
		postings := idx.invertedIndex[term]
		filtered := postings[:0]
		for _, p := range postings {
			if p.docID != docID {
				filtered = append(filtered, p)
			}
		}
		if len(filtered) == 0 {
			delete(idx.invertedIndex, term)
		} else {
			idx.invertedIndex[term] = filtered
		}
	}
}

func (idx *Index) Get(docID string) (models.Document, bool) {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	stats, ok := idx.docs[docID]
	if !ok {
		return models.Document{}, false
	}

	return stats.doc, true
}

func (idx *Index) Len() int {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	return len(idx.docs)
}

// Find documents matching the query and rank them by TF-IDF.
func (idx *Index) Search(query string, limit int) []models.SearchResult {
	queryTokens := Tokenize(query, idx.minTokenLen)
	if len(queryTokens) == 0 {
		return nil
	}

	idx.mu.RLock()
	defer idx.mu.RUnlock()

	docCount := len(idx.docs)

	// For each query token, find the set of matching document IDs
	var matchingIDs map[string]bool
	for _, term := range queryTokens {
		postings := idx.invertedIndex[term]
		termDocs := make(map[string]bool, len(postings))
		for _, p := range postings {
			termDocs[p.docID] = true
		}
		if matchingIDs == nil {
			matchingIDs = termDocs
		} else {
			for id := range matchingIDs {
				if !termDocs[id] {
					delete(matchingIDs, id)
				}
			}
		}
	}

	// Score each document
	results := make([]models.SearchResult, 0, len(matchingIDs))
	for docID := range matchingIDs {
		stats := idx.docs[docID]
		score := idx.scoreDocument(stats, queryTokens, docCount)
		results = append(results, models.SearchResult{
			Document: stats.doc,
			Score:    score,
		})
	}

	sort.Slice(results, func(i, j int) bool {
		if results[i].Score == results[j].Score {
			return results[i].Score > results[j].Score
		}
		return results[i].Document.ID < results[j].Document.ID
	})

	if limit > 0 && len(results) > limit {
		results = results[:limit]
	}

	return results
}

func (idx *Index) scoreDocument(stats docStats, queryTokens []string, docCount int) float64 {
	var total float64

	for _, term := range queryTokens {
		termCount := stats.termCounts[term]
		docsWithTerm := len(idx.invertedIndex[term])
		total += TFIDF(termCount, stats.tokenCount, docCount, docsWithTerm)
	}

	return total
}

func (idx *Index) IndexedAt(DocID string) (time.Time, bool) {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	stats, ok := idx.docs[DocID]
	if !ok {
		return time.Time{}, false
	}

	return stats.doc.IndexedAt, true
}