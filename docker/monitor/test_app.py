import app


def test_get_env_existing(monkeypatch):
    monkeypatch.setenv("TEST_KEY", "value123")
    assert app.get_env("TEST_KEY", "fallback") == "value123"


def test_get_env_fallback(monkeypatch):
    monkeypatch.delenv("TEST_KEY", raising=False)
    assert app.get_env("TEST_KEY", "fallback") == "fallback"


def test_get_version(monkeypatch):
    monkeypatch.setenv("VERSION", "1.0.0")
    assert app.get_version() == "1.0.0"


def test_get_version_default(monkeypatch):
    monkeypatch.delenv("VERSION", raising=False)
    assert app.get_version() == "dev"


def test_status_version_render(monkeypatch):
    monkeypatch.setenv("VERSION", "9.9.9")

    client = app.app.test_client()
    res = client.get("/status")

    assert b"9.9.9" in res.data


def test_status_route():
    client = app.app.test_client()

    response = client.get("/status")

    assert response.status_code == 200
    assert b"Monitor Dashboard" in response.data
    assert b"Monitor Service Version" in response.data
