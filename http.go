package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
)

const maxBodySize = 1 << 20 // 1 MB

type session struct {
	server *MCPServer
}

type HTTPHandler struct {
	sessions       sync.Map
	sessionCount   atomic.Int64
	maxSessions    int
	allowedOrigins []string
}

func NewHTTPHandler(allowedOrigins []string, maxSessions int) *HTTPHandler {
	if maxSessions <= 0 {
		maxSessions = 1000
	}
	return &HTTPHandler{
		allowedOrigins: allowedOrigins,
		maxSessions:    maxSessions,
	}
}

func (h *HTTPHandler) setCORS(w http.ResponseWriter, r *http.Request) {
	if len(h.allowedOrigins) == 0 {
		w.Header().Set("Access-Control-Allow-Origin", "*")
	} else {
		origin := r.Header.Get("Origin")
		for _, allowed := range h.allowedOrigins {
			if allowed == origin {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Vary", "Origin")
				break
			}
		}
	}
	w.Header().Set("Access-Control-Allow-Headers",
		"Content-Type, Authorization, Mcp-Session-Id, "+
			"X-Umami-Host, X-Umami-Username, X-Umami-Password, X-Umami-Api-Key, X-Umami-Team-Id")
	w.Header().Set("Access-Control-Expose-Headers", "Mcp-Session-Id")
}

func (h *HTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.setCORS(w, r)

	switch r.Method {
	case http.MethodOptions:
		w.WriteHeader(http.StatusNoContent)
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
	body, err := io.ReadAll(io.LimitReader(r.Body, maxBodySize+1))
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	if len(body) > maxBodySize {
		http.Error(w, "Request body too large", http.StatusRequestEntityTooLarge)
		return
	}

	var msg struct {
		ID     any    `json:"id"`
		Method string `json:"method"`
	}
	if err := json.Unmarshal(body, &msg); err != nil {
		writeJSONRPCError(w, nil, &Error{Code: -32700, Message: "Parse error"})
		return
	}

	if msg.ID == nil {
		w.WriteHeader(http.StatusAccepted)
		return
	}

	var req Request
	if err := json.Unmarshal(body, &req); err != nil {
		writeJSONRPCError(w, nil, &Error{Code: -32700, Message: "Parse error"})
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

type umamiCreds struct {
	host     string
	username string
	password string
	apiKey   string
}

func (c umamiCreds) valid() bool {
	if c.host == "" {
		return false
	}
	if c.apiKey != "" {
		return true
	}
	return c.username != "" && c.password != ""
}

func extractUmamiCreds(r *http.Request) umamiCreds {
	creds := umamiCreds{
		host:     r.Header.Get("X-Umami-Host"),
		username: r.Header.Get("X-Umami-Username"),
		password: r.Header.Get("X-Umami-Password"),
		apiKey:   r.Header.Get("X-Umami-Api-Key"),
	}
	if creds.valid() {
		return creds
	}

	query := r.URL.Query()
	qHost := query.Get("umamiHost")
	qUser := query.Get("umamiUsername")
	qPass := query.Get("umamiPassword")
	qKey := query.Get("umamiApiKey")
	if qHost != "" || qUser != "" || qPass != "" || qKey != "" {
		log.Printf("DEPRECATED: credentials in query params — use X-Umami-* headers instead")
	}
	if creds.host == "" {
		creds.host = qHost
	}
	if creds.username == "" {
		creds.username = qUser
	}
	if creds.password == "" {
		creds.password = qPass
	}
	if creds.apiKey == "" {
		creds.apiKey = qKey
	}
	return creds
}

const missingCredsMsg = "Missing required credentials: provide X-Umami-Host plus either " +
	"X-Umami-Api-Key (Umami Cloud) or X-Umami-Username and X-Umami-Password (self-hosted)"

func (h *HTTPHandler) handleInitialize(w http.ResponseWriter, r *http.Request, req Request) {
	creds := extractUmamiCreds(r)

	if !creds.valid() {
		writeJSONRPCError(w, req.ID, &Error{
			Code:    -32602,
			Message: missingCredsMsg,
		})
		return
	}

	if int(h.sessionCount.Load()) >= h.maxSessions {
		writeJSONRPCError(w, req.ID, &Error{
			Code:    -32603,
			Message: "Maximum sessions reached",
		})
		return
	}

	var client *UmamiClient
	if creds.apiKey != "" {
		client = NewUmamiClientWithAPIKey(creds.host, creds.apiKey)
	} else {
		client = NewUmamiClient(creds.host, creds.username, creds.password)
	}
	if teamID := r.Header.Get("X-Umami-Team-Id"); teamID != "" {
		client.teamID = teamID
	}
	if err := client.Authenticate(); err != nil {
		writeJSONRPCError(w, req.ID, &Error{
			Code:    -32603,
			Message: fmt.Sprintf("Authentication failed: %v", err),
		})
		return
	}

	sessionID := generateSessionID()
	srv := NewMCPServer(client)
	h.sessions.Store(sessionID, &session{server: srv})
	h.sessionCount.Add(1)

	resp := srv.HandleRequest(req)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Mcp-Session-Id", sessionID)
	data, _ := json.Marshal(resp)
	_, _ = w.Write(data)

	log.Printf("New session %s for %s", sessionID, creds.host)
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

	h.sessionCount.Add(-1)
	w.WriteHeader(http.StatusOK)
}

func generateSessionID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func (h *HTTPHandler) handleServerCard(w http.ResponseWriter, _ *http.Request) {
	toolsData, _ := toolsFS.ReadFile("tools.json")
	promptsData, _ := promptsFS.ReadFile("prompts.json")

	var tools []json.RawMessage
	_ = json.Unmarshal(toolsData, &tools)

	var prompts []json.RawMessage
	_ = json.Unmarshal(promptsData, &prompts)

	card := map[string]any{
		"serverInfo": map[string]string{
			"name":    "umami-mcp",
			"version": version,
		},
		"tools":   tools,
		"prompts": prompts,
		"resources": []map[string]any{
			{
				"uri":         "umami://websites",
				"name":        "Website List",
				"description": "List of all websites configured in Umami",
				"mimeType":    "application/json",
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	data, _ := json.Marshal(card)
	_, _ = w.Write(data)
}

func writeJSONRPCError(w http.ResponseWriter, id any, rpcErr *Error) {
	resp := Response{JSONRPC: "2.0", ID: id, Error: rpcErr}
	w.Header().Set("Content-Type", "application/json")
	data, _ := json.Marshal(resp)
	_, _ = w.Write(data)
}

func parseOrigins(raw string) []string {
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	origins := make([]string, 0, len(parts))
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			origins = append(origins, trimmed)
		}
	}
	return origins
}
