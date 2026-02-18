package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
)

type session struct {
	server *MCPServer
}

type HTTPHandler struct {
	sessions sync.Map // map[string]*session
}

func NewHTTPHandler() *HTTPHandler {
	return &HTTPHandler{}
}

func (h *HTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.handlePost(w, r)
	case http.MethodGet:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	case http.MethodDelete:
		h.handleDelete(w, r)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func (h *HTTPHandler) handlePost(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	var msg struct {
		ID     any    `json:"id"`
		Method string `json:"method"`
	}
	if err := json.Unmarshal(body, &msg); err != nil {
		writeJSONRPC(w, nil, nil, &Error{Code: -32700, Message: "Parse error"}, http.StatusOK)
		return
	}

	if msg.ID == nil {
		w.WriteHeader(http.StatusAccepted)
		return
	}

	var req Request
	if err := json.Unmarshal(body, &req); err != nil {
		writeJSONRPC(w, nil, nil, &Error{Code: -32700, Message: "Parse error"}, http.StatusOK)
		return
	}

	if req.Method == "initialize" {
		h.handleInitialize(w, r, req)
		return
	}

	sessionID := r.Header.Get("Mcp-Session-Id")
	if sessionID == "" {
		http.Error(w, "Missing Mcp-Session-Id header", http.StatusBadRequest)
		return
	}

	val, ok := h.sessions.Load(sessionID)
	if !ok {
		http.Error(w, "Invalid session", http.StatusNotFound)
		return
	}

	sess := val.(*session)
	resp := sess.server.HandleRequest(req)

	w.Header().Set("Content-Type", "application/json")
	data, _ := json.Marshal(resp)
	_, _ = w.Write(data)
}

func (h *HTTPHandler) handleInitialize(w http.ResponseWriter, r *http.Request, req Request) {
	query := r.URL.Query()
	umamiHost := query.Get("umamiHost")
	umamiUsername := query.Get("umamiUsername")
	umamiPassword := query.Get("umamiPassword")

	if umamiHost == "" || umamiUsername == "" || umamiPassword == "" {
		writeJSONRPC(w, req.ID, nil, &Error{
			Code:    -32602,
			Message: "Missing required query params: umamiHost, umamiUsername, umamiPassword",
		}, http.StatusOK)
		return
	}

	client := NewUmamiClient(umamiHost, umamiUsername, umamiPassword)
	if err := client.Authenticate(); err != nil {
		writeJSONRPC(w, req.ID, nil, &Error{
			Code:    -32603,
			Message: fmt.Sprintf("Authentication failed: %v", err),
		}, http.StatusOK)
		return
	}

	sessionID := generateSessionID()
	srv := NewMCPServer(client)
	h.sessions.Store(sessionID, &session{server: srv})

	resp := srv.HandleRequest(req)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Mcp-Session-Id", sessionID)
	data, _ := json.Marshal(resp)
	_, _ = w.Write(data)

	log.Printf("New session %s for %s", sessionID, umamiHost)
}

func (h *HTTPHandler) handleDelete(w http.ResponseWriter, r *http.Request) {
	sessionID := r.Header.Get("Mcp-Session-Id")
	if sessionID == "" {
		http.Error(w, "Missing Mcp-Session-Id header", http.StatusBadRequest)
		return
	}

	if _, ok := h.sessions.LoadAndDelete(sessionID); !ok {
		http.Error(w, "Invalid session", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func generateSessionID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func writeJSONRPC(w http.ResponseWriter, id, result any, rpcErr *Error, status int) {
	resp := Response{JSONRPC: "2.0", ID: id}
	if rpcErr != nil {
		resp.Error = rpcErr
	} else {
		resp.Result = result
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	data, _ := json.Marshal(resp)
	_, _ = w.Write(data)
}
