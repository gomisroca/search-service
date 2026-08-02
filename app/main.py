"""
Entry point. Run with:
    uvicorn app.main:app --host 0.0.0.0 --port 8080
"""

from __future__ import annotations

from contextlib import asynccontextmanager
from datetime import datetime, timezone

from fastapi import Depends, FastAPI, HTTPException, Query
from fastapi.middleware.cors import CORSMiddleware

from app.auth import make_api_key_dependency
from app.config import load_settings
from app.index.index import Index
from app.models import (
    Document,
    IndexRequest,
    IndexResponse,
    SearchResponse,
    SearchResult,
)

settings = load_settings()
index = Index(min_token_length=settings.min_token_length)


@asynccontextmanager
async def lifespan(_: FastAPI):
    if not settings.api_key:
        print("WARNING: API_KEY is not set, all endpoints are open.")
    print(
        f"Search Service starting on :{settings.port} "
        f"(min_token_length={settings.min_token_length})"
    )
    yield


app = FastAPI(title="Search Service", lifespan=lifespan)

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_methods=["GET", "POST", "DELETE", "OPTIONS"],
    allow_headers=["Content-Type", "X-API-Key"],
)

require_api_key = make_api_key_dependency(settings.api_key)


@app.get("/health")
async def health() -> dict:
    return {"status": "ok", "documents": str(len(index))}


@app.post("/documents", dependencies=[Depends(require_api_key)], status_code=201)
async def index_document(req: IndexRequest) -> IndexResponse:
    if settings.max_documents > 0 and len(index) >= settings.max_documents:
        if index.get(req.id) is None:
            raise HTTPException(status_code=503, detail="Index is full")

    now = datetime.now(timezone.utc)
    doc = Document(
        id=req.id,
        content=req.content,
        metadata=req.metadata,
        indexed_at=now,
    )
    await index.add(doc)
    return IndexResponse(id=req.id, indexed_at=now)


@app.get("/documents/{doc_id}", dependencies=[Depends(require_api_key)])
async def get_document(doc_id: str) -> Document:
    doc = index.get(doc_id)
    if doc is None:
        raise HTTPException(status_code=404, detail="Document not found")
    return doc


@app.delete("/documents/{doc_id}", dependencies=[Depends(require_api_key)])
async def delete_document(doc_id: str) -> dict:
    if not await index.delete(doc_id):
        raise HTTPException(status_code=404, detail="Document not found")
    return {"deleted": doc_id}


@app.get("/search", dependencies=[Depends(require_api_key)])
async def search(
    q: str = Query(..., description="Search query"),
    limit: int = Query(default=0, ge=0),
) -> SearchResponse:
    if not q.strip():
        raise HTTPException(status_code=400, detail="q parameter cannot be empty")

    resolved_limit = limit if limit > 0 else settings.default_search_limit
    if resolved_limit > settings.max_search_limit:
        resolved_limit = settings.max_search_limit

    raw_results = index.search(q, resolved_limit)

    results = [
        SearchResult(document=r.document, score=r.score)
        for r in raw_results
    ]

    return SearchResponse(query=q, total=len(results), results=results)