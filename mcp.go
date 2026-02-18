package main

import (
	"bufio"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

//go:embed mcp-tools-schema.json
var toolsFS embed.FS

type MCPServer struct {
	client *UmamiClient
	stdin  io.Reader
	stdout io.Writer
}

func NewMCPServer(client *UmamiClient) *MCPServer {
	return &MCPServer{
		client: client,
		stdin:  os.Stdin,
		stdout: os.Stdout,
	}
}

func (s *MCPServer) Run() error {
	scanner := bufio.NewScanner(s.stdin)
	for scanner.Scan() {
		var rawMsg json.RawMessage
		if err := json.Unmarshal(scanner.Bytes(), &rawMsg); err != nil {
			s.send(Response{JSONRPC: "2.0", ID: nil, Error: &Error{Code: -32700, Message: "Parse error"}})
			continue
		}

		var msgType struct {
			ID     any    `json:"id"`
			Method string `json:"method"`
		}
		if err := json.Unmarshal(rawMsg, &msgType); err != nil {
			s.send(Response{JSONRPC: "2.0", ID: nil, Error: &Error{Code: -32700, Message: "Parse error"}})
			continue
		}

		if msgType.ID != nil {
			var req Request
			if err := json.Unmarshal(rawMsg, &req); err != nil {
				s.send(Response{JSONRPC: "2.0", ID: nil, Error: &Error{Code: -32700, Message: "Parse error"}})
				continue
			}
			s.send(s.HandleRequest(req))
		}
		// Notifications (no id): silently ignore
	}
	return scanner.Err()
}

func (s *MCPServer) HandleRequest(req Request) Response {
	var result any
	var rpcErr *Error

	switch req.Method {
	case "initialize":
		result, rpcErr = s.processInitialize()
	case "tools/list":
		result, rpcErr = s.processToolsList()
	case "tools/call":
		result, rpcErr = s.processToolCall(req.Params)
	case "resources/list":
		result = map[string]any{"resources": []any{}}
	case "prompts/list":
		result = map[string]any{"prompts": []any{}}
	default:
		rpcErr = &Error{Code: -32601, Message: "Method not found"}
	}

	if rpcErr != nil {
		return Response{JSONRPC: "2.0", ID: req.ID, Error: rpcErr}
	}
	return Response{JSONRPC: "2.0", ID: req.ID, Result: result}
}

func (s *MCPServer) send(resp Response) {
	data, _ := json.Marshal(resp)
	_, _ = fmt.Fprintf(s.stdout, "%s\n", data)
}

func (s *MCPServer) processInitialize() (any, *Error) {
	return map[string]any{
		"protocolVersion": "2024-11-05",
		"serverInfo": map[string]string{
			"name":    "umami-mcp",
			"version": version,
		},
		"capabilities": map[string]any{
			"tools": map[string]any{},
		},
	}, nil
}

func (s *MCPServer) processToolsList() (any, *Error) {
	toolsData, err := toolsFS.ReadFile("mcp-tools-schema.json")
	if err != nil {
		return nil, &Error{Code: -32603, Message: fmt.Sprintf("Failed to load tools: %v", err)}
	}

	var tools []map[string]any
	if err := json.Unmarshal(toolsData, &tools); err != nil {
		return nil, &Error{Code: -32603, Message: fmt.Sprintf("Failed to parse tools: %v", err)}
	}

	return map[string]any{"tools": tools}, nil
}

func (s *MCPServer) processToolCall(rawParams json.RawMessage) (any, *Error) {
	var params struct {
		Name      string          `json:"name"`
		Arguments json.RawMessage `json:"arguments"`
	}

	if err := json.Unmarshal(rawParams, &params); err != nil {
		return nil, &Error{Code: -32602, Message: "Invalid params"}
	}

	switch params.Name {
	case "get_websites":
		return s.execGetWebsites()
	case "get_stats":
		return s.execGetStats(params.Arguments)
	case "get_pageviews":
		return s.execGetPageViews(params.Arguments)
	case "get_metrics":
		return s.execGetMetrics(params.Arguments)
	case "get_active":
		return s.execGetActive(params.Arguments)
	default:
		return nil, &Error{Code: -32602, Message: fmt.Sprintf("Unknown tool: %s", params.Name)}
	}
}
