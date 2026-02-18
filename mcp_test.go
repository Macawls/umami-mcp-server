package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
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

	capabilities, ok := result["capabilities"].(map[string]any)
	if !ok {
		t.Fatal("Capabilities is not a map")
	}
	for _, cap := range []string{"tools", "prompts", "resources"} {
		if _, ok := capabilities[cap]; !ok {
			t.Errorf("Missing capability: %s", cap)
		}
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
	toolsData, err := toolsFS.ReadFile("tools.json")
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

func TestMCPServer_HandlePromptsList(t *testing.T) {
	server := &MCPServer{client: &UmamiClient{}}

	resp := server.HandleRequest(Request{JSONRPC: "2.0", ID: 1, Method: "prompts/list"})

	if resp.Error != nil {
		t.Fatalf("Unexpected error: %v", resp.Error)
	}

	result, ok := resp.Result.(map[string]any)
	if !ok {
		t.Fatal("Result is not a map")
	}

	prompts, ok := result["prompts"].([]map[string]any)
	if !ok {
		t.Fatal("Prompts is not []map[string]any")
	}

	if len(prompts) != 4 {
		t.Fatalf("Expected 4 prompts, got %d", len(prompts))
	}

	expectedPrompts := []string{"analytics-report", "top-pages", "visitor-insights", "realtime-check"}
	for i, prompt := range prompts {
		name, ok := prompt["name"].(string)
		if !ok {
			t.Errorf("Prompt %d name is not a string", i)
			continue
		}
		if name != expectedPrompts[i] {
			t.Errorf("Prompt %d: expected %s, got %s", i, expectedPrompts[i], name)
		}

		desc, hasDesc := prompt["description"].(string)
		if !hasDesc || desc == "" {
			t.Errorf("Prompt %s missing description", name)
		}

		if _, hasArgs := prompt["arguments"]; !hasArgs {
			t.Errorf("Prompt %s missing arguments", name)
		}
	}
}

func TestMCPServer_HandlePromptsGet(t *testing.T) {
	server := &MCPServer{client: &UmamiClient{}}

	params, _ := json.Marshal(map[string]any{
		"name":      "analytics-report",
		"arguments": map[string]string{"days": "14"},
	})

	resp := server.HandleRequest(Request{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "prompts/get",
		Params:  params,
	})

	if resp.Error != nil {
		t.Fatalf("Unexpected error: %v", resp.Error)
	}

	result, ok := resp.Result.(map[string]any)
	if !ok {
		t.Fatal("Result is not a map")
	}

	messages, ok := result["messages"].([]map[string]any)
	if !ok {
		t.Fatal("Messages is not []map[string]any")
	}

	if len(messages) != 1 {
		t.Fatalf("Expected 1 message, got %d", len(messages))
	}

	if messages[0]["role"] != "user" {
		t.Errorf("Expected role 'user', got %v", messages[0]["role"])
	}

	content, ok := messages[0]["content"].(map[string]any)
	if !ok {
		t.Fatal("Content is not a map")
	}

	text, ok := content["text"].(string)
	if !ok {
		t.Fatal("Text is not a string")
	}

	if !strings.Contains(text, "14") {
		t.Error("Expected message to contain interpolated days value '14'")
	}

	if strings.Contains(text, "{days}") {
		t.Error("Template variable {days} was not interpolated")
	}
}

func TestMCPServer_HandlePromptsGetNotFound(t *testing.T) {
	server := &MCPServer{client: &UmamiClient{}}

	params, _ := json.Marshal(map[string]any{
		"name": "nonexistent-prompt",
	})

	resp := server.HandleRequest(Request{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "prompts/get",
		Params:  params,
	})

	if resp.Error == nil {
		t.Fatal("Expected error for unknown prompt")
	}

	if resp.Error.Code != -32602 {
		t.Errorf("Expected error code -32602, got %d", resp.Error.Code)
	}
}

func TestMCPServer_HandleResourcesList(t *testing.T) {
	server := &MCPServer{client: &UmamiClient{}}

	resp := server.HandleRequest(Request{JSONRPC: "2.0", ID: 1, Method: "resources/list"})

	if resp.Error != nil {
		t.Fatalf("Unexpected error: %v", resp.Error)
	}

	result, ok := resp.Result.(map[string]any)
	if !ok {
		t.Fatal("Result is not a map")
	}

	resources, ok := result["resources"].([]map[string]any)
	if !ok {
		t.Fatal("Resources is not []map[string]any")
	}

	if len(resources) != 1 {
		t.Fatalf("Expected 1 resource, got %d", len(resources))
	}

	r := resources[0]
	if r["uri"] != "umami://websites" {
		t.Errorf("Expected URI 'umami://websites', got %v", r["uri"])
	}
	if r["name"] != "Website List" {
		t.Errorf("Expected name 'Website List', got %v", r["name"])
	}
	if r["mimeType"] != "application/json" {
		t.Errorf("Expected mimeType 'application/json', got %v", r["mimeType"])
	}
}

func TestMCPServer_HandleResourcesRead(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/auth/login":
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintln(w, `{"token":"test-token"}`)
		case "/api/websites":
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintln(w, `{"data":[{"id":"site1","name":"Test Site",`+
				`"domain":"example.com","createdAt":"2024-01-01T00:00:00Z"}],`+
				`"count":1,"page":1,"pageSize":100}`)
		}
	}))
	defer ts.Close()

	client := NewUmamiClient(ts.URL, "admin", "password")
	if err := client.Authenticate(); err != nil {
		t.Fatalf("Failed to authenticate: %v", err)
	}

	server := &MCPServer{client: client}

	params, _ := json.Marshal(map[string]any{
		"uri": "umami://websites",
	})

	resp := server.HandleRequest(Request{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "resources/read",
		Params:  params,
	})

	if resp.Error != nil {
		t.Fatalf("Unexpected error: %v", resp.Error)
	}

	result, ok := resp.Result.(map[string]any)
	if !ok {
		t.Fatal("Result is not a map")
	}

	contents, ok := result["contents"].([]map[string]any)
	if !ok {
		t.Fatal("Contents is not []map[string]any")
	}

	if len(contents) != 1 {
		t.Fatalf("Expected 1 content entry, got %d", len(contents))
	}

	text, ok := contents[0]["text"].(string)
	if !ok {
		t.Fatal("Text is not a string")
	}

	if !strings.Contains(text, "site1") {
		t.Error("Expected resource content to contain website ID 'site1'")
	}
}
