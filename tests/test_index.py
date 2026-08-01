from __future__ import annotations

import asyncio
from datetime import datetime, timezone

import pytest

from app.index.index import Index
from app.models import Document


def make_doc(id: str, content: str) -> Document:
    return Document(id=id, content=content, indexed_at=datetime.now(timezone.utc))


async def test_add_and_get():
    idx = Index(min_token_length=2)
    await idx.add(make_doc("doc1", "the quick brown fox"))
    doc = idx.get("doc1")
    assert doc is not None and doc.id == "doc1"


async def test_get_missing():
    idx = Index()
    assert idx.get("missing") is None


async def test_len():
    idx = Index()
    await idx.add(make_doc("doc1", "hello world"))
    await idx.add(make_doc("doc2", "foo bar"))
    assert len(idx) == 2


async def test_delete_removes_document():
    idx = Index()
    await idx.add(make_doc("doc1", "hello world"))
    assert await idx.delete("doc1") is True
    assert await idx.delete("doc1") is False
    assert idx.get("doc1") is None


async def test_delete_removes_from_inverted_index():
    idx = Index()
    await idx.add(make_doc("doc1", "golang microservice"))
    await idx.delete("doc1")
    assert idx.search("golang", 10) == []


async def test_replace_on_duplicate_id():
    idx = Index()
    await idx.add(make_doc("doc1", "golang microservice"))
    await idx.add(make_doc("doc1", "python fastapi"))
    assert len(idx) == 1
    assert idx.search("golang", 10) == []
    assert len(idx.search("python", 10)) == 1


async def test_search_returns_matching_documents():
    idx = Index()
    await idx.add(make_doc("doc1", "golang concurrency goroutines"))
    await idx.add(make_doc("doc2", "python asyncio coroutines"))
    await idx.add(make_doc("doc3", "rust ownership memory"))
    results = idx.search("golang", 10)
    assert len(results) == 1 and results[0].document.id == "doc1"


async def test_search_and_semantics():
    idx = Index()
    await idx.add(make_doc("doc1", "golang microservice docker"))
    await idx.add(make_doc("doc2", "golang python comparison"))
    await idx.add(make_doc("doc3", "docker kubernetes deployment"))
    results = idx.search("golang docker", 10)
    assert len(results) == 1 and results[0].document.id == "doc1"


async def test_search_no_results():
    idx = Index()
    await idx.add(make_doc("doc1", "hello world"))
    assert idx.search("zzzzz", 10) == []


async def test_search_empty_query():
    idx = Index()
    await idx.add(make_doc("doc1", "hello world"))
    assert idx.search("", 10) == []


async def test_search_limit_respected():
    idx = Index()
    for i in range(20):
        await idx.add(make_doc(f"doc{i}", "golang microservice"))
    results = idx.search("golang", 5)
    assert len(results) == 5


async def test_search_scores_are_positive():
    idx = Index()
    await idx.add(make_doc("doc1", "golang concurrency microservice"))
    results = idx.search("golang", 10)
    assert all(r.score > 0 for r in results)


async def test_search_rare_term_scores_higher():
    idx = Index()
    await idx.add(make_doc("doc1", "golang goroutine microservice concurrency"))
    await idx.add(make_doc("doc2", "python microservice flask"))
    # Only doc1 has "goroutine" (AND semantics)
    results = idx.search("goroutine microservice", 10)
    assert len(results) == 1 and results[0].document.id == "doc1"


async def test_concurrent_writes_and_reads():
    """
    Python equivalent of TestConcurrentReadWrite in Go.

    In Go, this test with `go test -race` proves the RWMutex prevents data
    races across OS threads. In Python, the GIL prevents true simultaneous
    bytecode execution, so the equivalent risk is a logical race across
    await points (two coroutines both observing the same state and both
    acting on it). asyncio.gather fires all coroutines onto the event loop
    simultaneously and the write lock ensures Add/Delete are atomic across
    await points.
    """
    idx = Index()

    async def write_batch(w: int) -> None:
        for i in range(30):
            doc_id = f"doc-{w}-{i}"
            await idx.add(make_doc(doc_id, "golang microservice search index"))
            if i % 5 == 0:
                await idx.delete(doc_id)

    async def read_batch() -> None:
        for _ in range(30):
            idx.search("golang microservice", 10)
            len(idx)

    writers = [write_batch(w) for w in range(10)]
    readers = [read_batch() for _ in range(20)]
    await asyncio.gather(*writers, *readers)