"""
API-level tests. These require fastapi, httpx, and pydantic.
"""

from __future__ import annotations

import importlib

import pytest


def _build_app(monkeypatch, **env):
    for k, v in env.items():
        monkeypatch.setenv(k, v)
    import app.config
    import app.main
    importlib.reload(app.config)
    importlib.reload(app.main)
    return app.main.app


class TestHealth:
    def test_health(self, monkeypatch):
        from fastapi.testclient import TestClient
        app = _build_app(monkeypatch)
        with TestClient(app) as client:
            res = client.get("/health")
            assert res.status_code == 200
            assert res.json()["status"] == "ok"


class TestIndexDocument:
    def test_index_returns_201(self, monkeypatch):
        from fastapi.testclient import TestClient
        app = _build_app(monkeypatch)
        with TestClient(app) as client:
            res = client.post("/documents", json={
                "id": "doc1",
                "content": "golang microservices goroutines",
            })
        assert res.status_code == 201
        assert res.json()["id"] == "doc1"

    def test_index_requires_id(self, monkeypatch):
        from fastapi.testclient import TestClient
        app = _build_app(monkeypatch)
        with TestClient(app) as client:
            res = client.post("/documents", json={"content": "hello"})
        assert res.status_code == 422

    def test_index_requires_content(self, monkeypatch):
        from fastapi.testclient import TestClient
        app = _build_app(monkeypatch)
        with TestClient(app) as client:
            res = client.post("/documents", json={"id": "doc1"})
        assert res.status_code == 422

    def test_index_api_key_required_when_set(self, monkeypatch):
        from fastapi.testclient import TestClient
        app = _build_app(monkeypatch, API_KEY="secret")
        with TestClient(app) as client:
            res = client.post("/documents", json={"id": "d", "content": "hello world"})
            assert res.status_code == 401
            res = client.post("/documents",
                              json={"id": "d", "content": "hello world"},
                              headers={"X-API-Key": "secret"})
            assert res.status_code == 201


class TestGetDocument:
    def test_get_existing(self, monkeypatch):
        from fastapi.testclient import TestClient
        app = _build_app(monkeypatch)
        with TestClient(app) as client:
            client.post("/documents", json={"id": "doc1", "content": "hello world"})
            res = client.get("/documents/doc1")
        assert res.status_code == 200
        assert res.json()["id"] == "doc1"

    def test_get_missing_returns_404(self, monkeypatch):
        from fastapi.testclient import TestClient
        app = _build_app(monkeypatch)
        with TestClient(app) as client:
            assert client.get("/documents/missing").status_code == 404


class TestDeleteDocument:
    def test_delete_existing(self, monkeypatch):
        from fastapi.testclient import TestClient
        app = _build_app(monkeypatch)
        with TestClient(app) as client:
            client.post("/documents", json={"id": "doc1", "content": "hello world"})
            res = client.delete("/documents/doc1")
            assert res.status_code == 200
            assert client.get("/documents/doc1").status_code == 404

    def test_delete_missing_returns_404(self, monkeypatch):
        from fastapi.testclient import TestClient
        app = _build_app(monkeypatch)
        with TestClient(app) as client:
            assert client.delete("/documents/ghost").status_code == 404


class TestSearch:
    def test_search_returns_results(self, monkeypatch):
        from fastapi.testclient import TestClient
        app = _build_app(monkeypatch)
        with TestClient(app) as client:
            client.post("/documents", json={"id": "d1", "content": "golang goroutines concurrency"})
            client.post("/documents", json={"id": "d2", "content": "python asyncio fastapi"})
            res = client.get("/search?q=goroutines")
        assert res.status_code == 200
        body = res.json()
        assert body["total"] == 1
        assert body["results"][0]["document"]["id"] == "d1"

    def test_search_and_semantics(self, monkeypatch):
        from fastapi.testclient import TestClient
        app = _build_app(monkeypatch)
        with TestClient(app) as client:
            client.post("/documents", json={"id": "d1", "content": "golang microservice docker"})
            client.post("/documents", json={"id": "d2", "content": "python microservice flask"})
            res = client.get("/search?q=golang+docker")
        assert res.json()["total"] == 1

    def test_search_empty_query_returns_400(self, monkeypatch):
        from fastapi.testclient import TestClient
        app = _build_app(monkeypatch)
        with TestClient(app) as client:
            assert client.get("/search?q=").status_code == 400

    def test_search_missing_q_returns_422(self, monkeypatch):
        from fastapi.testclient import TestClient
        app = _build_app(monkeypatch)
        with TestClient(app) as client:
            assert client.get("/search").status_code == 422

    def test_search_respects_limit(self, monkeypatch):
        from fastapi.testclient import TestClient
        app = _build_app(monkeypatch)
        with TestClient(app) as client:
            for i in range(10):
                client.post("/documents", json={"id": f"d{i}", "content": "golang microservice"})
            res = client.get("/search?q=golang&limit=3")
        assert res.json()["total"] == 3

    def test_search_no_results(self, monkeypatch):
        from fastapi.testclient import TestClient
        app = _build_app(monkeypatch)
        with TestClient(app) as client:
            client.post("/documents", json={"id": "d1", "content": "hello world"})
            res = client.get("/search?q=zzzzz")
        body = res.json()
        assert res.status_code == 200
        assert body["total"] == 0