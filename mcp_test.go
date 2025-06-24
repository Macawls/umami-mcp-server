package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestMCPServer_HandleInitialize(t *testing.T) {
	var output bytes.Buffer
	server := &MCPServer{client: &UmamiClient{}, stdout: &output}
	
	server.handleInitialize(Request{JSONRPC: "2.0", ID: 1, Method: "initialize"})

	var resp Response
	if err := json.Unmarshal(output.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if resp.Error != nil {
		t.Errorf("Expected no error, got: %v", resp.Error)
	}

	result, ok := resp.Result.(map[string]interface{})
	if !ok {
		t.Fatal("Result is not a map")
	}
	if result["protocolVersion"] != "2024-11-05" {
		t.Errorf("Wrong protocol version: %v", result["protocolVersion"])
	}
}

func TestMCPServer_HandleToolsList(t *testing.T) {
	var output bytes.Buffer
	server := &MCPServer{client: &UmamiClient{}, stdout: &output}
	
	server.handleToolsList(Request{JSONRPC: "2.0", ID: 2, Method: "tools/list"})

	var resp Response
	if err := json.Unmarshal(output.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}
	
	if resp.Error != nil {
		t.Fatalf("Unexpected error: %v", resp.Error)
	}

	result, ok := resp.Result.(map[string]interface{})
	if !ok {
		t.Fatal("Result is not a map")
	}
	toolsInterface, ok := result["tools"].([]interface{})
	if !ok {
		t.Fatal("Tools is not an array")
	}
	
	if len(toolsInterface) != 5 {
		t.Fatalf("Expected 5 tools, got %d", len(toolsInterface))
	}

	expectedTools := []string{"get_websites", "get_stats", "get_pageviews", "get_metrics", "get_active"}
	for i, toolInterface := range toolsInterface {
		tool, ok := toolInterface.(map[string]interface{})
		if !ok {
			t.Errorf("Tool %d is not a map", i)
			continue
		}
		
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
	input := strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"unknown"}` + "\n")
	var output bytes.Buffer
	server := &MCPServer{client: &UmamiClient{}, stdin: input, stdout: &output}

	_ = server.Run()

	var resp Response
	if err := json.Unmarshal(output.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}
	
	if resp.Error == nil || resp.Error.Code != -32601 {
		t.Error("Expected error -32601 for unknown method")
	}
}

func TestMCPServer_ToolsJSONValidity(t *testing.T) {
	toolsData, err := toolsFS.ReadFile("mcp-tools-schema.json")
	if err != nil {
		t.Fatalf("Failed to read tools JSON: %v", err)
	}

	var tools []map[string]interface{}
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