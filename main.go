package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

type Request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      any             `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type Response struct {
	JSONRPC string `json:"jsonrpc"`
	ID      any    `json:"id"`
	Result  any    `json:"result,omitempty"`
	Error   *Error `json:"error,omitempty"`
}

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "-v" || os.Args[1] == "--version") {
		fmt.Printf("umami-mcp %s (%s) built %s\n", version, commit, date)
		os.Exit(0)
	}

	transport := os.Getenv("TRANSPORT")
	if transport == "" {
		transport = "stdio"
	}

	switch transport {
	case "http":
		port := os.Getenv("PORT")
		if port == "" {
			port = "8080"
		}
		handler := NewHTTPHandler()
		mux := http.NewServeMux()
		mux.Handle("/mcp", handler)
		mux.HandleFunc("/.well-known/mcp/server-card.json", handler.handleServerCard)
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, "/app/index.html")
		})
		srv := &http.Server{
			Addr:              ":" + port,
			Handler:           mux,
			ReadHeaderTimeout: 10 * time.Second,
		}
		log.Printf("Starting HTTP transport on :%s", port)
		if err := srv.ListenAndServe(); err != nil {
			log.Fatalf("HTTP server error: %v", err)
		}
	default:
		config, err := LoadConfig()
		if err != nil {
			log.Fatalf("Failed to load config: %v", err)
		}

		client := NewUmamiClient(config.UmamiURL, config.Username, config.Password)
		if err := client.Authenticate(); err != nil {
			log.Fatalf("Failed to authenticate with Umami: %v", err)
		}

		server := NewMCPServer(client)
		if err := server.Run(); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	}
}
