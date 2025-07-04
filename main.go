package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
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
