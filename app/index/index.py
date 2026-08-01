"""
Inverted index with TF-IDF ranking.
"""

from __future__ import annotations

import asyncio
from dataclasses import dataclass, field
from datetime import datetime, timezone
from typing import Any, Protocol

from app.index.scorer import tfidf
from app.index.tokenizer import term_frequencies, tokenize

class DocumentLike(Protocol):
    id: str
    content: str
    indexed_at: datetime
    metadata: Any


@dataclass
class SearchResult:
    document: Any
    score: float


@dataclass
class _Posting:
    doc_id: str
    term_count: int


@dataclass
class _DocStats:
    doc: Any
    token_count: int
    term_counts: dict[str, int] = field(default_factory=dict)


class Index:
    def __init__(self, min_token_length: int = 2) -> None:
        self._min_token_length = min_token_length
        # term > list of postings
        self._inverted: dict[str, list[_Posting]] = {}
        # doc_id > stats
        self._docs: dict[str, _DocStats] = {}
        # Protects multi-step write operations (Add, Delete).
        self._write_lock = asyncio.Lock()

    async def add(self, doc: DocumentLike) -> None:
        tokens = tokenize(doc.content, self._min_token_length)
        term_counts = term_frequencies(tokens)

        async with self._write_lock:
            # Remove old postings if this doc already exists.
            if doc.id in self._docs:
                self._remove_postings(doc.id, self._docs[doc.id].term_counts)

            # Insert new postings.
            for term, count in term_counts.items():
                if term not in self._inverted:
                    self._inverted[term] = []
                self._inverted[term].append(_Posting(doc_id=doc.id, term_count=count))

            self._docs[doc.id] = _DocStats(
                doc=doc,
                token_count=len(tokens),
                term_counts=term_counts,
            )

    async def delete(self, doc_id: str) -> bool:
        async with self._write_lock:
            if doc_id not in self._docs:
                return False
            self._remove_postings(doc_id, self._docs[doc_id].term_counts)
            del self._docs[doc_id]
            return True

    def _remove_postings(self, doc_id: str, term_counts: dict[str, int]) -> None:
        """Must be called with _write_lock held."""
        for term in term_counts:
            if term not in self._inverted:
                continue
            self._inverted[term] = [
                p for p in self._inverted[term] if p.doc_id != doc_id
            ]
            if not self._inverted[term]:
                del self._inverted[term]

    def get(self, doc_id: str) -> DocumentLike | None:
        stats = self._docs.get(doc_id)
        return stats.doc if stats else None

    def search(self, query: str, limit: int) -> list[SearchResult]:
        query_tokens = tokenize(query, self._min_token_length)
        if not query_tokens:
            return []

        doc_count = len(self._docs)

        matching_ids: set[str] | None = None
        for term in query_tokens:
            postings = self._inverted.get(term, [])
            term_doc_ids = {p.doc_id for p in postings}
            if matching_ids is None:
                matching_ids = term_doc_ids
            else:
                matching_ids &= term_doc_ids
            if not matching_ids:
                return []

        if not matching_ids:
            return []

        # Score each matching document.
        results = []
        for doc_id in matching_ids:
            stats = self._docs[doc_id]
            score = self._score(stats, query_tokens, doc_count)
            results.append(SearchResult(document=stats.doc, score=score))

        # Sort descending by score, then by id for stable output.
        results.sort(key=lambda r: (-r.score, r.document.id))

        return results[:limit] if limit > 0 else results

    def _score(self, stats: _DocStats, query_tokens: list[str], doc_count: int) -> float:
        total = 0.0
        for term in query_tokens:
            term_count = stats.term_counts.get(term, 0)
            docs_with_term = len(self._inverted.get(term, []))
            total += tfidf(term_count, stats.token_count, doc_count, docs_with_term)
        return total

    def __len__(self) -> int:
        return len(self._docs)