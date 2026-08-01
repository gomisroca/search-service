"""
TF-IDF scoring - direct translation of Go's internal/index/scorer.go.
 
The formulas are identical. Python's math.log is the natural log, same as
Go's math.Log, so scores produced by both versions are numerically equivalent
for the same input.
"""
 
from __future__ import annotations
 
import math
 
 
def tf(term_count: int, total_tokens: int) -> float:
    """
    Term frequency: proportion of document tokens that are this term.
    Normalising by document length prevents longer documents from ranking
    higher purely because they have more words.
 
        TF(term, doc) = count(term in doc) / total tokens in doc
    """
    if total_tokens == 0:
        return 0.0
    return term_count / total_tokens
 
 
def idf(doc_count: int, docs_with_term: int) -> float:
    """
    Inverse document frequency: how rare the term is across the corpus.
    Uses the smoothed formula to avoid log(0) when a term appears in all
    documents.
 
        IDF(term) = log((1 + N) / (1 + df)) + 1
    """
    if doc_count == 0:
        return 0.0
    return math.log((1 + doc_count) / (1 + docs_with_term)) + 1
 
 
def tfidf(term_count: int, total_tokens: int, doc_count: int, docs_with_term: int) -> float:
    """TF x IDF - the standard relevance score."""
    return tf(term_count, total_tokens) * idf(doc_count, docs_with_term)
 
