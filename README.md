# Search Service (Go)

A standalone full-text search microservice. POST documents into it, search
across them instantly with TF-IDF ranking. Any service can use it: image
metadata, webhook event history, notes, product descriptions, whatever.
No external search engine required; the index lives in memory and is built
from scratch.

## Running it

```bash
go run .
# or
go build -o server . && ./server
```

### Tests

```bash
go test -race ./...
```

### Docker

```bash
docker build -t search-service .
docker run -p 8080:8080 \
  -e API_KEY=your-secret \
  search-service
```

## Config

| Variable               | Default   | Description                                                     |
| ---------------------- | --------- | --------------------------------------------------------------- |
| `PORT`                 | `8080`    | HTTP port                                                       |
| `API_KEY`              | _(empty)_ | If set, required via `X-API-Key` on all routes except `/health` |
| `MAX_DOCUMENTS`        | `0`       | Max documents in the index. `0` = no limit                      |
| `DEFAULT_SEARCH_LIMIT` | `10`      | Results per search when `?limit=` is omitted                    |
| `MAX_SEARCH_LIMIT`     | `100`     | Hard cap on `?limit=`                                           |
| `MIN_TOKEN_LENGTH`     | `2`       | Tokens shorter than this are skipped during indexing            |

## API

### `POST /documents`

Index a document. If a document with the same `id` already exists it is
replaced. Old postings are removed, new ones are added.

```json
{
  "id": "doc-001",
  "content": "Go microservices with goroutines and channels",
  "metadata": {
    "author": "alice",
    "url": "https://example.com/post/1",
    "tags": ["go", "microservices"]
  }
}
```

`metadata` is optional. It's stored and returned in results but not
indexed or searchable.

**Response `201 Created`:**

```json
{ "id": "doc-001", "indexed_at": "2026-07-24T09:00:00Z" }
```

### `GET /search?q=goroutines+microservices&limit=5`

Search across all indexed documents. Multi-word queries use AND semantics.
Only documents containing **all** query terms are returned. Results are
ranked by TF-IDF score (descending).

**Response `200 OK`:**

```json
{
  "query": "goroutines microservices",
  "total": 1,
  "results": [
    {
      "document": {
        "id": "doc-001",
        "content": "Go microservices with goroutines and channels",
        "metadata": { "author": "alice" },
        "indexed_at": "2026-07-24T09:00:00Z"
      },
      "score": 0.847
    }
  ]
}
```

### `GET /documents/{id}`

Fetch a specific document by ID. Returns `404` if not found.

### `DELETE /documents/{id}`

Remove a document from the index. Returns `404` if not found.

### `GET /health`

```json
{ "status": "ok", "documents": "42" }
```

Always open, no API key required.

## How it works

### Inverted index

The core data structure is an **inverted index**: a map from every term to
the list of documents containing it:

```
"goroutine"   → [doc-001, doc-007, doc-023]
"microservice" → [doc-001, doc-002, doc-007]
"python"      → [doc-002, doc-009]
```

When you search for "goroutine microservice", the service looks up both
posting lists and intersects them: only documents in both lists (doc-001,
doc-007) are candidates. Those candidates are then scored and sorted.

### TF-IDF ranking

Each candidate document is scored using **TF-IDF**:

- **TF (term frequency):** how often the query term appears in _this_
  document, normalised by document length. A 10-word document with "goroutine"
  3 times scores higher than a 1000-word document with it twice.
- **IDF (inverse document frequency):** how rare the term is across _all_
  documents. "goroutine" in 3 of 1000 documents scores much higher than
  "microservice" in 500 of 1000 - it's more discriminating.
- **TF-IDF = TF × IDF** - terms that are both frequent in the document _and_
  rare in the corpus produce the highest scores.

We use the smoothed IDF formula `log((1+N)/(1+df)) + 1` to avoid division
by zero and keep scores positive even when a term appears in every document.

### Concurrency

Searches and reads take `sync.RWMutex.RLock()` - many can run simultaneously.
Indexing and deleting take the full `Lock()`. This is the standard Go
pattern for read-heavy workloads: readers never block each other, and a
writer only blocks when it actually needs to mutate the map.

The `TestConcurrentReadWrite` test in `index_test.go` exercises this with
30 goroutines running simultaneously. Running it with `go test -race`
instruments every memory access and confirms there's no data race - not
just that the results happen to be correct on this run.

## Calling it from another service

```js
// Index a document (e.g. after an image upload)
await fetch("https://search.example.com/documents", {
  method: "POST",
  headers: { "Content-Type": "application/json", "X-API-Key": "..." },
  body: JSON.stringify({
    id: imageId,
    content: `${filename} ${tags.join(" ")} ${description}`,
    metadata: { url: imageUrl, uploadedBy: userId },
  }),
});

// Search
const res = await fetch("https://search.example.com/search?q=sunset+iceland", {
  headers: { "X-API-Key": "..." },
});
const { results } = await res.json();
```
