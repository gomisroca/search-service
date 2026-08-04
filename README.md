# Search Service (Python)

A Python/FastAPI port of [`Search Service (Go)`](https://github.com/gomisroca/search-service/tree/go). Same
HTTP API contract, swap one for the other without changing any caller.
The interesting comparison is in `app/index/` where the Go version's
`sync.RWMutex` and the Python version's `asyncio.Lock` serve the same
purpose through completely different mechanisms.

## Running it

```bash
pip install -r requirements.txt
uvicorn app.main:app --host 0.0.0.0 --port 8080
```

### Tests

```bash
pytest -v
```

The test suite has two tiers:

- `tests/test_tokenizer.py` and `tests/test_scorer.py`: pure Python
  standard library, no external deps, always runnable.
- `tests/test_index.py`: requires only `pytest` and `pytest-asyncio`;
  covers the inverted index including the concurrent write+read test via
  `asyncio.gather`.
- `tests/test_api.py`: requires `fastapi`, `pydantic`, and `httpx`;
  full API-level tests via FastAPI's `TestClient`.

### Docker

```bash
docker build -t search-service-py .
docker run -p 8080:8080 \
  -e API_KEY=your-secret \
  search-service-py
```

## Config

Same env vars as the Go version: `PORT`, `API_KEY`, `MAX_DOCUMENTS`,
`DEFAULT_SEARCH_LIMIT`, `MAX_SEARCH_LIMIT`, `MIN_TOKEN_LENGTH`.

## API

Same as the Go version. See its README for the full API reference. The
HTTP contract is identical: same routes, same request/response shapes, same
`X-API-Key` header.

## What's different, and why

**`asyncio.Lock` vs `sync.RWMutex`, the core difference.**

The Go version uses `sync.RWMutex`:

- Many goroutines can hold `RLock` simultaneously, running on different
  CPU cores in true parallel.
- A write (`Add`/`Delete`) takes the exclusive `Lock`, blocking all readers.
- `go test -race` instruments every memory access and proves the mutex
  is used correctly.

The Python version uses `asyncio.Lock`, but only for writes:

```python
async def add(self, doc) -> None:
    tokens = tokenize(doc.content, self._min_token_length)  # no await
    term_counts = term_frequencies(tokens)                   # no await

    async with self._write_lock:          # only the write path locks
        if doc.id in self._docs:
            self._remove_postings(...)
        # ... insert new postings
```

`Search` has no lock at all:

```python
def search(self, query: str, limit: int) -> list[SearchResult]:
    # Read-only. No await. GIL ensures we see a consistent snapshot.
    query_tokens = tokenize(query, ...)
    for term in query_tokens:
        postings = self._inverted.get(term, [])  # atomic dict read
        ...
```

Why is this safe? FastAPI route handlers are async coroutines on a single
event-loop thread. The GIL means two coroutines can never execute Python
bytecode at literally the same instant. More importantly, `search()` is a
synchronous method with no `await`, it runs to completion without ever
yielding control to the event loop, so no other coroutine can interleave.

The write path IS multi-step (remove old postings, insert new ones) and
DOES use `await` (inside `async with self._write_lock`). Without the lock,
two concurrent `add()` calls could interleave at the await point and both
try to remove the same old postings. The lock prevents that.

**The practical consequence:** Go's `RWMutex` allows many searches to run
truly in parallel across CPU cores. Python's GIL serialises all coroutines
onto one thread, so concurrent searches don't actually overlap, they take
turns. For CPU-bound indexing of large documents, Go wins decisively. For
typical web service traffic where individual searches are fast (milliseconds),
the difference is negligible.

**`asyncio.gather` concurrency test vs `go test -race`.**

`TestConcurrentReadWrite` in `tests/test_index.py` fires 5 writers and 10
readers simultaneously via `asyncio.gather`. This catches logical races
across `await` points, the Python-specific failure class, but can't detect
the OS-thread data races that `go test -race` catches in Go, because the GIL
prevents those from occurring in the first place.

**Pydantic validation vs manual validation.**

The Go version validates `id` and `content` by hand in the handler.
Pydantic's `@field_validator` does the same work declaratively in `models.py`,
and FastAPI returns a structured `422` automatically on validation failure.

**`Protocol` instead of a shared type.**

`app/index/index.py` defines a `DocumentLike` Protocol rather than importing
the Pydantic `Document` model. This keeps the index package decoupled from
FastAPI and Pydantic, the index can be tested with plain `@dataclass`
objects, no framework required. Go achieves the same separation by passing
`models.Document` which is also a plain struct with no framework coupling.

## Possible next steps

- **Persistent index**: serialize to disk on shutdown, reload on startup.
- **Stemming**: reduce "running"/"runs"/"ran" to "run" before indexing.
- **OR semantics**: `?mode=or` for union instead of intersection.
- **Phrase search**: requires storing token positions in postings.
