package main

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestMCPServer_HandleInitialize(t *testing.T) {
	server := &MCPServer{client: &UmamiClient{}}

	resp := server.HandleRequest(Request{JSONRPC: "2.0", ID: 1, Method: "initialize"})

	if resp.Error != nil {
		t.Errorf("Expected no error, got: %v", resp.Error)
	}

	result, ok := resp.Result.(map[string]any)
	if !ok {
		t.Fatal("Result is not a map")
	}
	if result["protocolVersion"] != "2024-11-05" {
		t.Errorf("Wrong protocol version: %v", result["protocolVersion"])
	}
}

func TestMCPServer_HandleToolsList(t *testing.T) {
	server := &MCPServer{client: &UmamiClient{}}

	resp := server.HandleRequest(Request{JSONRPC: "2.0", ID: 2, Method: "tools/list"})

	if resp.Error != nil {
		t.Fatalf("Unexpected error: %v", resp.Error)
	}

	result, ok := resp.Result.(map[string]any)
	if !ok {
		t.Fatal("Result is not a map")
	}
	toolsInterface, ok := result["tools"].([]map[string]any)
	if !ok {
		t.Fatal("Tools is not []map[string]any")
	}

	if len(toolsInterface) != 5 {
		t.Fatalf("Expected 5 tools, got %d", len(toolsInterface))
	}

	expectedTools := []string{"get_websites", "get_stats", "get_pageviews", "get_metrics", "get_active"}
	for i, tool := range toolsInterface {
		name, ok := tool["name"].(string)
		if !ok {
			t.Errorf("Tool %d name is not a string", i)
			continue
		}

		if name != expectedTools[i] {
			t.Errorf("Tool %d: expected %s, got %s", i, expectedTools[i], name)
		}

		desc, hasDesc := tool["description"].(string)
		_, hasSchema := tool["inputSchema"]
		if !hasDesc || desc == "" || !hasSchema {
			t.Errorf("Tool %s missing required fields", name)
		}

		if name == "get_websites" && !strings.Contains(desc, "CRITICAL") {
			t.Error("get_websites must emphasize CRITICAL importance")
		}
	}
}

func TestMCPServer_UnknownMethod(t *testing.T) {
	server := &MCPServer{client: &UmamiClient{}}

	resp := server.HandleRequest(Request{JSONRPC: "2.0", ID: 1, Method: "unknown"})

	if resp.Error == nil || resp.Error.Code != -32601 {
		t.Error("Expected error -32601 for unknown method")
	}
}

func TestMCPServer_ToolsJSONValidity(t *testing.T) {
	toolsData, err := toolsFS.ReadFile("mcp-tools-schema.json")
	if err != nil {
		t.Fatalf("Failed to read tools JSON: %v", err)
	}

	var tools []map[string]any
	if err := json.Unmarshal(toolsData, &tools); err != nil {
		t.Fatalf("Failed to parse tools JSON: %v", err)
	}

	if len(tools) != 5 {
		t.Fatalf("Expected 5 tools, got %d", len(tools))
	}

	for i, tool := range tools {
		_, hasName := tool["name"]
		_, hasDesc := tool["description"]
		_, hasSchema := tool["inputSchema"]
		if !hasName || !hasDesc || !hasSchema {
			t.Errorf("Tool %d missing required fields", i)
		}
	}
}
