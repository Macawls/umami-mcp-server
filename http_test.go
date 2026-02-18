package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func setupTestUmamiServer() *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/auth/login", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{"token":"test-token"}`)
	})
	mux.HandleFunc("/api/websites", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{"data":[]}`)
	})
	return httptest.NewServer(mux)
}

func mcpURL(umamiURL string) string {
	return "/mcp?umamiHost=" + umamiURL +
		"&umamiUsername=admin&umamiPassword=pass"
}

func initializeSession(
	t *testing.T, handler *HTTPHandler, umamiURL string,
) string {
	t.Helper()
	body := `{"jsonrpc":"2.0","id":1,"method":"initialize"}`
	req := httptest.NewRequest(
		http.MethodPost, mcpURL(umamiURL), strings.NewReader(body),
	)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("initialize returned %d: %s", w.Code, w.Body.String())
	}

	sessionID := w.Header().Get("Mcp-Session-Id")
	if sessionID == "" {
		t.Fatal("No Mcp-Session-Id in response")
	}
	return sessionID
}

func TestHTTP_Initialize(t *testing.T) {
	umami := setupTestUmamiServer()
	defer umami.Close()

	handler := NewHTTPHandler()
	_ = initializeSession(t, handler, umami.URL)

	w := httptest.NewRecorder()
	initBody := `{"jsonrpc":"2.0","id":1,"method":"initialize"}`
	req := httptest.NewRequest(
		http.MethodPost, mcpURL(umami.URL), strings.NewReader(initBody),
	)
	handler.ServeHTTP(w, req)

	var resp Response
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if resp.Error != nil {
		t.Errorf("Expected no error, got: %v", resp.Error)
	}

	result, ok := resp.Result.(map[string]any)
	if !ok {
		t.Fatal("Result is not a map")
	}
	if result["protocolVersion"] != "2025-03-26" {
		t.Errorf("Wrong protocol version: %v", result["protocolVersion"])
	}
}

func TestHTTP_ToolsList(t *testing.T) {
	umami := setupTestUmamiServer()
	defer umami.Close()

	handler := NewHTTPHandler()
	sessionID := initializeSession(t, handler, umami.URL)

	body := `{"jsonrpc":"2.0","id":2,"method":"tools/list"}`
	req := httptest.NewRequest(
		http.MethodPost, "/mcp", strings.NewReader(body),
	)
	req.Header.Set("Mcp-Session-Id", sessionID)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("tools/list returned %d", w.Code)
	}

	var resp Response
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}
	if resp.Error != nil {
		t.Fatalf("Unexpected error: %v", resp.Error)
	}
}

func TestHTTP_Notification(t *testing.T) {
	handler := NewHTTPHandler()
	body := `{"jsonrpc":"2.0","method":"notifications/initialized"}`
	req := httptest.NewRequest(
		http.MethodPost, "/mcp", strings.NewReader(body),
	)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusAccepted {
		t.Errorf("Expected 202, got %d", w.Code)
	}
}

func TestHTTP_MissingSessionHeader(t *testing.T) {
	handler := NewHTTPHandler()
	body := `{"jsonrpc":"2.0","id":1,"method":"tools/list"}`
	req := httptest.NewRequest(
		http.MethodPost, "/mcp", strings.NewReader(body),
	)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", w.Code)
	}
}

func TestHTTP_DeleteSession(t *testing.T) {
	umami := setupTestUmamiServer()
	defer umami.Close()

	handler := NewHTTPHandler()
	sessionID := initializeSession(t, handler, umami.URL)

	req := httptest.NewRequest(http.MethodDelete, "/mcp", http.NoBody)
	req.Header.Set("Mcp-Session-Id", sessionID)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}

	body := `{"jsonrpc":"2.0","id":1,"method":"tools/list"}`
	req2 := httptest.NewRequest(
		http.MethodPost, "/mcp", strings.NewReader(body),
	)
	req2.Header.Set("Mcp-Session-Id", sessionID)
	w2 := httptest.NewRecorder()
	handler.ServeHTTP(w2, req2)

	if w2.Code != http.StatusNotFound {
		t.Errorf("Expected 404 after delete, got %d", w2.Code)
	}
}

func TestHTTP_OptionsPreflight(t *testing.T) {
	handler := NewHTTPHandler()
	req := httptest.NewRequest(http.MethodOptions, "/mcp", http.NoBody)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected 204, got %d", w.Code)
	}
	if w.Header().Get("Access-Control-Allow-Origin") == "" {
		t.Error("Missing Access-Control-Allow-Origin header")
	}
	if w.Header().Get("Access-Control-Expose-Headers") == "" {
		t.Error("Missing Access-Control-Expose-Headers header")
	}
}

func TestHTTP_GetMethodNotAllowed(t *testing.T) {
	handler := NewHTTPHandler()
	req := httptest.NewRequest(http.MethodGet, "/mcp", http.NoBody)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected 405, got %d", w.Code)
	}
}
