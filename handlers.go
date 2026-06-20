package main

import (
	"encoding/json"
	"fmt"
)

func (s *MCPServer) execGetWebsites(args json.RawMessage) (any, *Error) {
	var params struct {
		IncludeTeams bool `json:"includeTeams"`
	}

	if len(args) > 0 {
		_ = json.Unmarshal(args, &params)
	}

	websites, err := s.client.GetWebsites(params.IncludeTeams)
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

func textContent(result any) map[string]any {
	data, _ := json.MarshalIndent(result, "", "  ")
	return map[string]any{"content": []map[string]string{{
		"type": "text",
		"text": string(data),
	}}}
}

func (s *MCPServer) dateRangeQuery(
	args json.RawMessage, errLabel string,
	query func(websiteID, startDate, endDate string) (any, error),
) (any, *Error) {
	var params struct {
		WebsiteID string `json:"website_id"`
		StartDate string `json:"start_date"`
		EndDate   string `json:"end_date"`
	}

	if err := json.Unmarshal(args, &params); err != nil {
		return nil, &Error{Code: -32602, Message: "Invalid arguments"}
	}

	if err := validateWebsiteID(params.WebsiteID); err != nil {
		return nil, &Error{Code: -32602, Message: "Invalid website_id"}
	}

	result, err := query(params.WebsiteID, normalizeDate(params.StartDate), normalizeDate(params.EndDate))
	if err != nil {
		return nil, &Error{Code: -32603, Message: fmt.Sprintf("Failed to get %s: %v", errLabel, err)}
	}

	return textContent(result), nil
}

func (s *MCPServer) execGetStats(args json.RawMessage) (any, *Error) {
	return s.dateRangeQuery(args, "stats", func(id, start, end string) (any, error) {
		return s.client.GetStats(id, start, end)
	})
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

	if err := validateWebsiteID(params.WebsiteID); err != nil {
		return nil, &Error{Code: -32602, Message: "Invalid website_id"}
	}

	if params.Unit == "" {
		params.Unit = "day"
	}

	params.StartDate = normalizeDate(params.StartDate)
	params.EndDate = normalizeDate(params.EndDate)

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

	if err := validateWebsiteID(params.WebsiteID); err != nil {
		return nil, &Error{Code: -32602, Message: "Invalid website_id"}
	}

	if params.Limit == 0 {
		params.Limit = 10
	}

	params.StartDate = normalizeDate(params.StartDate)
	params.EndDate = normalizeDate(params.EndDate)

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

	if err := validateWebsiteID(params.WebsiteID); err != nil {
		return nil, &Error{Code: -32602, Message: "Invalid website_id"}
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

func (s *MCPServer) execGetSessions(args json.RawMessage) (any, *Error) {
	var params struct {
		WebsiteID string `json:"website_id"`
		StartDate string `json:"start_date"`
		EndDate   string `json:"end_date"`
		Search    string `json:"search"`
		Page      int    `json:"page"`
		PageSize  int    `json:"page_size"`
	}

	if err := json.Unmarshal(args, &params); err != nil {
		return nil, &Error{Code: -32602, Message: "Invalid arguments"}
	}

	if err := validateWebsiteID(params.WebsiteID); err != nil {
		return nil, &Error{Code: -32602, Message: "Invalid website_id"}
	}

	params.StartDate = normalizeDate(params.StartDate)
	params.EndDate = normalizeDate(params.EndDate)

	sessions, err := s.client.GetSessions(
		params.WebsiteID, params.StartDate, params.EndDate, params.Search, params.Page, params.PageSize,
	)
	if err != nil {
		return nil, &Error{Code: -32603, Message: fmt.Sprintf("Failed to get sessions: %v", err)}
	}

	data, _ := json.MarshalIndent(sessions, "", "  ")
	content := []map[string]string{{
		"type": "text",
		"text": string(data),
	}}

	return map[string]any{"content": content}, nil
}

func (s *MCPServer) execGetSessionStats(args json.RawMessage) (any, *Error) {
	return s.dateRangeQuery(args, "session stats", func(id, start, end string) (any, error) {
		return s.client.GetSessionStats(id, start, end)
	})
}

func (s *MCPServer) execGetSessionActivity(args json.RawMessage) (any, *Error) {
	var params struct {
		WebsiteID string `json:"website_id"`
		SessionID string `json:"session_id"`
		StartDate string `json:"start_date"`
		EndDate   string `json:"end_date"`
	}

	if err := json.Unmarshal(args, &params); err != nil {
		return nil, &Error{Code: -32602, Message: "Invalid arguments"}
	}

	if err := validateWebsiteID(params.WebsiteID); err != nil {
		return nil, &Error{Code: -32602, Message: "Invalid website_id"}
	}
	if err := validateSessionID(params.SessionID); err != nil {
		return nil, &Error{Code: -32602, Message: "Invalid session_id"}
	}

	params.StartDate = normalizeDate(params.StartDate)
	params.EndDate = normalizeDate(params.EndDate)

	activity, err := s.client.GetSessionActivity(
		params.WebsiteID, params.SessionID, params.StartDate, params.EndDate,
	)
	if err != nil {
		return nil, &Error{Code: -32603, Message: fmt.Sprintf("Failed to get session activity: %v", err)}
	}

	data, _ := json.MarshalIndent(activity, "", "  ")
	content := []map[string]string{{
		"type": "text",
		"text": string(data),
	}}

	return map[string]any{"content": content}, nil
}
