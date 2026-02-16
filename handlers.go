package main

import (
	"encoding/json"
	"fmt"
)

func (s *MCPServer) handleGetWebsites(id any, args json.RawMessage) {
	var params struct {
		IncludeTeams  bool `json:"includeTeams"`
		IncludeTeams2 bool `json:"include_teams"`
	}

	if len(args) > 0 {
		if err := json.Unmarshal(args, &params); err != nil {
			s.sendError(id, -32602, "Invalid arguments")
			return
		}
	}

	websites, err := s.client.GetWebsites(params.IncludeTeams || params.IncludeTeams2)
	if err != nil {
		s.sendError(id, -32603, fmt.Sprintf("Failed to get websites: %v", err))
		return
	}

	data, _ := json.MarshalIndent(websites, "", "  ")
	content := []map[string]string{{
		"type": "text",
		"text": string(data),
	}}

	s.sendResult(id, map[string]any{"content": content})
}

func (s *MCPServer) handleGetStats(id any, args json.RawMessage) {
	var params struct {
		WebsiteID string `json:"website_id"`
		StartDate string `json:"start_date"`
		EndDate   string `json:"end_date"`
	}

	if err := json.Unmarshal(args, &params); err != nil {
		s.sendError(id, -32602, "Invalid arguments")
		return
	}

	stats, err := s.client.GetStats(params.WebsiteID, params.StartDate, params.EndDate)
	if err != nil {
		s.sendError(id, -32603, fmt.Sprintf("Failed to get stats: %v", err))
		return
	}

	data, _ := json.MarshalIndent(stats, "", "  ")
	content := []map[string]string{{
		"type": "text",
		"text": string(data),
	}}

	s.sendResult(id, map[string]any{"content": content})
}
func (s *MCPServer) handleGetPageViews(id any, args json.RawMessage) {
	var params struct {
		WebsiteID string `json:"website_id"`
		StartDate string `json:"start_date"`
		EndDate   string `json:"end_date"`
		Unit      string `json:"unit"`
	}

	if err := json.Unmarshal(args, &params); err != nil {
		s.sendError(id, -32602, "Invalid arguments")
		return
	}

	if params.Unit == "" {
		params.Unit = "day"
	}

	pageviews, err := s.client.GetPageViews(params.WebsiteID, params.StartDate, params.EndDate, params.Unit)
	if err != nil {
		s.sendError(id, -32603, fmt.Sprintf("Failed to get page views: %v", err))
		return
	}

	data, _ := json.MarshalIndent(pageviews, "", "  ")
	content := []map[string]string{{
		"type": "text",
		"text": string(data),
	}}

	s.sendResult(id, map[string]any{"content": content})
}
func (s *MCPServer) handleGetMetrics(id any, args json.RawMessage) {
	var params struct {
		WebsiteID  string `json:"website_id"`
		StartDate  string `json:"start_date"`
		EndDate    string `json:"end_date"`
		MetricType string `json:"metric_type"`
		Limit      int    `json:"limit"`
	}

	if err := json.Unmarshal(args, &params); err != nil {
		s.sendError(id, -32602, "Invalid arguments")
		return
	}

	if params.Limit == 0 {
		params.Limit = 10
	}

	metrics, err := s.client.GetMetrics(
		params.WebsiteID, params.StartDate, params.EndDate, params.MetricType, params.Limit,
	)
	if err != nil {
		s.sendError(id, -32603, fmt.Sprintf("Failed to get metrics: %v", err))
		return
	}

	data, _ := json.MarshalIndent(metrics, "", "  ")
	content := []map[string]string{{
		"type": "text",
		"text": string(data),
	}}

	s.sendResult(id, map[string]any{"content": content})
}

func (s *MCPServer) handleGetActive(id any, args json.RawMessage) {
	var params struct {
		WebsiteID string `json:"website_id"`
	}

	if err := json.Unmarshal(args, &params); err != nil {
		s.sendError(id, -32602, "Invalid arguments")
		return
	}

	active, err := s.client.GetActive(params.WebsiteID)
	if err != nil {
		s.sendError(id, -32603, fmt.Sprintf("Failed to get active visitors: %v", err))
		return
	}

	data, _ := json.MarshalIndent(active, "", "  ")
	content := []map[string]string{{
		"type": "text",
		"text": string(data),
	}}

	s.sendResult(id, map[string]any{"content": content})
}
