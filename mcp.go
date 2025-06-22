package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

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
		var req Request
		if err := json.Unmarshal(scanner.Bytes(), &req); err != nil {
			s.sendError(nil, -32700, "Parse error")
			continue
		}

		switch req.Method {
		case "initialize":
			s.handleInitialize(req)
		case "tools/list":
			s.handleToolsList(req)
		case "tools/call":
			s.handleToolCall(req)
		default:
			s.sendError(req.ID, -32601, "Method not found")
		}
	}
	return scanner.Err()
}

func (s *MCPServer) send(resp Response) {
	data, _ := json.Marshal(resp)
	fmt.Fprintf(s.stdout, "%s\n", data)
}

func (s *MCPServer) sendError(id interface{}, code int, message string) {
	s.send(Response{
		JSONRPC: "2.0",
		ID:      id,
		Error: &Error{
			Code:    code,
			Message: message,
		},
	})
}

func (s *MCPServer) sendResult(id interface{}, result interface{}) {
	s.send(Response{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	})
}
func (s *MCPServer) handleInitialize(req Request) {
	result := map[string]interface{}{
		"protocolVersion": "2024-11-05",
		"serverInfo": map[string]string{
			"name":    "umami-mcp",
			"version": version,
		},
		"capabilities": map[string]interface{}{
			"tools": map[string]interface{}{},
		},
	}
	s.sendResult(req.ID, result)
}

func (s *MCPServer) handleToolsList(req Request) {
	tools := []map[string]interface{}{
		{
			"name":        "get_websites",
			"description": "Get list of all websites configured in Umami",
			"inputSchema": map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
		{
			"name":        "get_stats",
			"description": "Get statistics for a website",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"website_id": map[string]interface{}{
						"type":        "string",
						"description": "The website ID",
					},
					"start_date": map[string]interface{}{
						"type":        "string",
						"description": "Start date timestamp in milliseconds",
					},
					"end_date": map[string]interface{}{
						"type":        "string",
						"description": "End date timestamp in milliseconds",
					},
				},
				"required": []string{"website_id", "start_date", "end_date"},
			},
		},
		{
			"name":        "get_pageviews",
			"description": "Get page view data for a website",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"website_id": map[string]interface{}{
						"type":        "string",
						"description": "The website ID",
					},
					"start_date": map[string]interface{}{
						"type":        "string",
						"description": "Start date timestamp in milliseconds",
					},
					"end_date": map[string]interface{}{
						"type":        "string",
						"description": "End date timestamp in milliseconds",
					},
					"unit": map[string]interface{}{
						"type":        "string",
						"description": "Time unit (hour, day, month, year)",
						"enum":        []string{"hour", "day", "month", "year"},
						"default":     "day",
					},
				},
				"required": []string{"website_id", "start_date", "end_date"},
			},
		},
		{
			"name":        "get_metrics",
			"description": "Get metrics for a website (browsers, OS, devices, etc)",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"website_id": map[string]interface{}{
						"type":        "string",
						"description": "The website ID",
					},
					"start_date": map[string]interface{}{
						"type":        "string",
						"description": "Start date timestamp in milliseconds",
					},
					"end_date": map[string]interface{}{
						"type":        "string",
						"description": "End date timestamp in milliseconds",
					},
					"metric_type": map[string]interface{}{
						"type":        "string",
						"description": "Type of metric",
						"enum":        []string{"url", "referrer", "browser", "os", "device", "country", "event"},
					},
					"limit": map[string]interface{}{
						"type":        "integer",
						"description": "Maximum results to return",
						"default":     10,
					},
				},
				"required": []string{"website_id", "start_date", "end_date", "metric_type"},
			},
		},
		{
			"name":        "get_active",
			"description": "Get current active visitors for a website",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"website_id": map[string]interface{}{
						"type":        "string",
						"description": "The website ID",
					},
				},
				"required": []string{"website_id"},
			},
		},
	}
	s.sendResult(req.ID, map[string]interface{}{"tools": tools})
}
func (s *MCPServer) handleToolCall(req Request) {
	var params struct {
		Name      string          `json:"name"`
		Arguments json.RawMessage `json:"arguments"`
	}

	if err := json.Unmarshal(req.Params, &params); err != nil {
		s.sendError(req.ID, -32602, "Invalid params")
		return
	}

	switch params.Name {
	case "get_websites":
		s.handleGetWebsites(req.ID)
	case "get_stats":
		s.handleGetStats(req.ID, params.Arguments)
	case "get_pageviews":
		s.handleGetPageViews(req.ID, params.Arguments)
	case "get_metrics":
		s.handleGetMetrics(req.ID, params.Arguments)
	case "get_active":
		s.handleGetActive(req.ID, params.Arguments)
	default:
		s.sendError(req.ID, -32602, fmt.Sprintf("Unknown tool: %s", params.Name))
	}
}