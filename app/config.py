from __future__ import annotations

import os
from dataclasses import dataclass


@dataclass(frozen=True)
class Settings:
    port: int
    api_key: str
    max_documents: int
    default_search_limit: int
    max_search_limit: int
    min_token_length: int


def load_settings() -> Settings:
    return Settings(
        port=int(os.getenv("PORT", "8080")),
        api_key=os.getenv("API_KEY", ""),
        max_documents=int(os.getenv("MAX_DOCUMENTS", "0")),
        default_search_limit=int(os.getenv("DEFAULT_SEARCH_LIMIT", "10")),
        max_search_limit=int(os.getenv("MAX_SEARCH_LIMIT", "100")),
        min_token_length=int(os.getenv("MIN_TOKEN_LENGTH", "2")),
    )