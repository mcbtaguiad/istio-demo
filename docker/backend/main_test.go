package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

// helper to create a request and record the response
func performRequest(handler http.Handler, method, path string, body any, token string) *httptest.ResponseRecorder {
	var buf bytes.Buffer
	if body != nil {
		json.NewEncoder(&buf).Encode(body)
	}

	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	return rr
}

func TestRegisterLoginProfile(t *testing.T) {
	// Start a mock Redis server
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	defer mr.Close()

	// Override global rdb with mock Redis
	rdb = redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	// Set up the HTTP handlers
	mux := http.NewServeMux()
	mux.HandleFunc("/api/register", registerHandler)
	mux.HandleFunc("/api/login", loginHandler)
	mux.HandleFunc("/api/profile", authMiddleware(profileHandler))

	creds := Credentials{
		Username: "testuser",
		Password: "password123",
	}

	// 1️⃣ Register
	resp := performRequest(mux, "POST", "/api/register", creds, "")
	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", resp.Code)
	}

	var regData map[string]string
	json.NewDecoder(resp.Body).Decode(&regData)
	if _, ok := regData["group"]; !ok {
		t.Fatalf("expected group in response")
	}

	// 2️⃣ Login
	resp = performRequest(mux, "POST", "/api/login", creds, "")
	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", resp.Code)
	}

	var loginData map[string]string
	json.NewDecoder(resp.Body).Decode(&loginData)
	token, ok := loginData["token"]
	if !ok || strings.TrimSpace(token) == "" {
		t.Fatalf("expected token in response")
	}

	// 3️⃣ Profile
	resp = performRequest(mux, "GET", "/api/profile", nil, token)
	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", resp.Code)
	}

	var profile map[string]string
	json.NewDecoder(resp.Body).Decode(&profile)
	if profile["username"] != creds.Username {
		t.Fatalf("expected username %s, got %s", creds.Username, profile["username"])
	}
}
