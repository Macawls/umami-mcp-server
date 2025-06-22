package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestMCPServer_HandleInitialize(t *testing.T) {
	client := &UmamiClient{}
	var output bytes.Buffer
	server := &MCPServer{
		client: client,
		stdout: &output,
	}

	req := Request{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
	}

	server.handleInitialize(req)

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

func TestMCPServer_UnknownMethod(t *testing.T) {
	client := &UmamiClient{}
	input := strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"unknown"}` + "\n")
	var output bytes.Buffer

	server := &MCPServer{
		client: client,
		stdin:  input,
		stdout: &output,
	}

	_ = server.Run()

	var resp Response
	if err := json.Unmarshal(output.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if resp.Error == nil {
		t.Error("Expected error for unknown method")
	}

	if resp.Error.Code != -32601 {
		t.Errorf("Expected error code -32601, got %d", resp.Error.Code)
	}
}