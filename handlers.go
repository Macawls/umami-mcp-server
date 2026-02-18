package main

import (
	"encoding/json"
	"fmt"
)

func (s *MCPServer) execGetWebsites() (any, *Error) {
	websites, err := s.client.GetWebsites()
	if err != nil {
		return nil, &Error{Code: -32603, Message: fmt.Sprintf("Failed to get websites: %v", err)}
	}

	data, _ := json.MarshalIndent(websites, "", "  ")
	content := []map[string]string{{
		"type": "text",
		"text": string(data),
	}}

	return map[string]any{"content": content}, nil
}

func (s *MCPServer) execGetStats(args json.RawMessage) (any, *Error) {
	var params struct {
		WebsiteID string `json:"website_id"`
		StartDate string `json:"start_date"`
		EndDate   string `json:"end_date"`
	}

	if err := json.Unmarshal(args, &params); err != nil {
		return nil, &Error{Code: -32602, Message: "Invalid arguments"}
	}

	stats, err := s.client.GetStats(params.WebsiteID, params.StartDate, params.EndDate)
	if err != nil {
		return nil, &Error{Code: -32603, Message: fmt.Sprintf("Failed to get stats: %v", err)}
	}

	data, _ := json.MarshalIndent(stats, "", "  ")
	content := []map[string]string{{
		"type": "text",
		"text": string(data),
	}}

	return map[string]any{"content": content}, nil
}

func (s *MCPServer) execGetPageViews(args json.RawMessage) (any, *Error) {
	var params struct {
		WebsiteID string `json:"website_id"`
		StartDate string `json:"start_date"`
		EndDate   string `json:"end_date"`
		Unit      string `json:"unit"`
	}

	if err := json.Unmarshal(args, &params); err != nil {
		return nil, &Error{Code: -32602, Message: "Invalid arguments"}
	}

	if params.Unit == "" {
		params.Unit = "day"
	}

	pageviews, err := s.client.GetPageViews(params.WebsiteID, params.StartDate, params.EndDate, params.Unit)
	if err != nil {
		return nil, &Error{Code: -32603, Message: fmt.Sprintf("Failed to get page views: %v", err)}
	}

	data, _ := json.MarshalIndent(pageviews, "", "  ")
	content := []map[string]string{{
		"type": "text",
		"text": string(data),
	}}

	return map[string]any{"content": content}, nil
}

func (s *MCPServer) execGetMetrics(args json.RawMessage) (any, *Error) {
	var params struct {
		WebsiteID  string `json:"website_id"`
		StartDate  string `json:"start_date"`
		EndDate    string `json:"end_date"`
		MetricType string `json:"metric_type"`
		Limit      int    `json:"limit"`
	}

	if err := json.Unmarshal(args, &params); err != nil {
		return nil, &Error{Code: -32602, Message: "Invalid arguments"}
	}

	if params.Limit == 0 {
		params.Limit = 10
	}

	metrics, err := s.client.GetMetrics(
		params.WebsiteID, params.StartDate, params.EndDate, params.MetricType, params.Limit,
	)
	if err != nil {
		return nil, &Error{Code: -32603, Message: fmt.Sprintf("Failed to get metrics: %v", err)}
	}

	data, _ := json.MarshalIndent(metrics, "", "  ")
	content := []map[string]string{{
		"type": "text",
		"text": string(data),
	}}

	return map[string]any{"content": content}, nil
}

func (s *MCPServer) execGetActive(args json.RawMessage) (any, *Error) {
	var params struct {
		WebsiteID string `json:"website_id"`
	}

	if err := json.Unmarshal(args, &params); err != nil {
		return nil, &Error{Code: -32602, Message: "Invalid arguments"}
	}

	active, err := s.client.GetActive(params.WebsiteID)
	if err != nil {
		return nil, &Error{Code: -32603, Message: fmt.Sprintf("Failed to get active visitors: %v", err)}
	}

	data, _ := json.MarshalIndent(active, "", "  ")
	content := []map[string]string{{
		"type": "text",
		"text": string(data),
	}}

	return map[string]any{"content": content}, nil
}
