package main

import (
	"bufio"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

//go:embed tools.json
var toolsFS embed.FS

//go:embed prompts.json
var promptsFS embed.FS

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
		result = s.processInitialize()
	case "tools/list":
		result, rpcErr = s.processToolsList()
	case "tools/call":
		result, rpcErr = s.processToolCall(req.Params)
	case "prompts/list":
		result, rpcErr = s.processPromptsList()
	case "prompts/get":
		result, rpcErr = s.processPromptsGet(req.Params)
	case "resources/list":
		result = s.processResourcesList()
	case "resources/read":
		result, rpcErr = s.processResourcesRead(req.Params)
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

func (s *MCPServer) processInitialize() any {
	return map[string]any{
		"protocolVersion": "2024-11-05",
		"serverInfo": map[string]string{
			"name":    "umami-mcp",
			"version": version,
		},
		"capabilities": map[string]any{
			"tools":     map[string]any{},
			"prompts":   map[string]any{},
			"resources": map[string]any{},
		},
	}
}

func (s *MCPServer) processToolsList() (any, *Error) {
	toolsData, err := toolsFS.ReadFile("tools.json")
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

func (s *MCPServer) processPromptsList() (any, *Error) {
	data, err := promptsFS.ReadFile("prompts.json")
	if err != nil {
		return nil, &Error{Code: -32603, Message: fmt.Sprintf("Failed to load prompts: %v", err)}
	}

	var prompts []map[string]any
	if err := json.Unmarshal(data, &prompts); err != nil {
		return nil, &Error{Code: -32603, Message: fmt.Sprintf("Failed to parse prompts: %v", err)}
	}

	return map[string]any{"prompts": prompts}, nil
}

var promptTemplates = map[string]string{
	"analytics-report": "First call get_websites to find the target website. " +
		"Then generate a comprehensive analytics report " +
		"covering the last {days} days by calling:\n" +
		"- get_stats for overall visitor metrics\n" +
		"- get_pageviews (unit: day) for traffic trends\n" +
		"- get_metrics for top pages (type: url), referrers (type: referrer), " +
		"countries (type: country), browsers (type: browser), " +
		"and devices (type: device)\n\n" +
		"Summarize the findings in a clear, well-structured report.",

	"top-pages": "First call get_websites to find the target website. " +
		"Then call get_metrics with type \"url\" over the last {days} days, " +
		"limited to {limit} results, to find the most visited pages. " +
		"Present the results as a ranked list with page paths and view counts.",

	"visitor-insights": "First call get_websites to find the target website. " +
		"Then break down visitors over the last {days} days " +
		"by calling get_metrics for each:\n" +
		"- type \"country\" for geographic distribution\n" +
		"- type \"device\" for device types\n" +
		"- type \"browser\" for browser usage\n" +
		"- type \"os\" for operating systems\n\n" +
		"Present the breakdown with counts and percentages.",

	"realtime-check": "First call get_websites to find the target website. " +
		"Then call get_active to check the current number of active visitors. " +
		"Report the real-time visitor count.",
}

var promptDefaults = map[string]map[string]string{
	"analytics-report": {"days": "30"},
	"top-pages":        {"days": "7", "limit": "10"},
	"visitor-insights":  {"days": "30"},
	"realtime-check":    {},
}

func (s *MCPServer) processPromptsGet(rawParams json.RawMessage) (any, *Error) {
	var params struct {
		Name      string            `json:"name"`
		Arguments map[string]string `json:"arguments"`
	}

	if err := json.Unmarshal(rawParams, &params); err != nil {
		return nil, &Error{Code: -32602, Message: "Invalid params"}
	}

	tmpl, ok := promptTemplates[params.Name]
	if !ok {
		return nil, &Error{Code: -32602, Message: fmt.Sprintf("Unknown prompt: %s", params.Name)}
	}

	// Apply defaults then overrides
	defaults := promptDefaults[params.Name]
	merged := make(map[string]string)
	for k, v := range defaults {
		merged[k] = v
	}
	for k, v := range params.Arguments {
		merged[k] = v
	}

	// Interpolate arguments into template
	message := tmpl
	for k, v := range merged {
		message = strings.ReplaceAll(message, "{"+k+"}", v)
	}

	return map[string]any{
		"description": params.Name,
		"messages": []map[string]any{
			{
				"role": "user",
				"content": map[string]any{
					"type": "text",
					"text": message,
				},
			},
		},
	}, nil
}

func (s *MCPServer) processResourcesList() any {
	return map[string]any{
		"resources": []map[string]any{
			{
				"uri":         "umami://websites",
				"name":        "Website List",
				"description": "List of all websites configured in Umami with their IDs and creation dates",
				"mimeType":    "application/json",
			},
		},
	}
}

func (s *MCPServer) processResourcesRead(rawParams json.RawMessage) (any, *Error) {
	var params struct {
		URI string `json:"uri"`
	}

	if err := json.Unmarshal(rawParams, &params); err != nil {
		return nil, &Error{Code: -32602, Message: "Invalid params"}
	}

	if params.URI != "umami://websites" {
		return nil, &Error{Code: -32602, Message: fmt.Sprintf("Unknown resource: %s", params.URI)}
	}

	websites, err := s.client.GetWebsites()
	if err != nil {
		return nil, &Error{Code: -32603, Message: fmt.Sprintf("Failed to get websites: %v", err)}
	}

	data, err := json.Marshal(websites)
	if err != nil {
		return nil, &Error{Code: -32603, Message: fmt.Sprintf("Failed to marshal websites: %v", err)}
	}

	return map[string]any{
		"contents": []map[string]any{
			{
				"uri":      "umami://websites",
				"mimeType": "application/json",
				"text":     string(data),
			},
		},
	}, nil
}
