from __future__ import annotations

from datetime import datetime
from typing import Any

from pydantic import BaseModel


class IndexRequest(BaseModel):
    id: str
    content: str
    metadata: dict[str, Any] | None = None


class Document(BaseModel):
    id: str
    content: str
    metadata: dict[str, Any] | None = None
    indexed_at: datetime


class SearchResult(BaseModel):
    document: Document
    score: float


class SearchResponse(BaseModel):
    query: str
    total: int
    results: list[SearchResult]


class IndexResponse(BaseModel):
    id: str
    indexed_at: datetime