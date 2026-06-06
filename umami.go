package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const metricTypePath = "path"

type UmamiClient struct {
	baseURL     string
	username    string
	password    string
	apiKey      string
	apiBasePath string
	token       string
	teamID      string
	httpClient  *http.Client
}

func NewUmamiClient(baseURL, username, password string) *UmamiClient {
	return &UmamiClient{
		baseURL:     strings.TrimSuffix(baseURL, "/"),
		username:    username,
		password:    password,
		apiBasePath: "/api",
		httpClient:  &http.Client{Timeout: 30 * time.Second},
	}
}

func NewUmamiClientWithAPIKey(baseURL, apiKey string) *UmamiClient {
	return &UmamiClient{
		baseURL:     strings.TrimSuffix(baseURL, "/"),
		apiKey:      apiKey,
		apiBasePath: "/v1",
		httpClient:  &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *UmamiClient) basePath() string {
	if c.apiBasePath == "" {
		return "/api"
	}
	return c.apiBasePath
}

func (c *UmamiClient) websitesPath() string {
	return c.basePath() + "/websites"
}

func (c *UmamiClient) Authenticate() error {
	if c.apiKey != "" {
		return nil
	}

	payload := map[string]string{
		"username": c.username,
		"password": c.password,
	}

	data, _ := json.Marshal(payload)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/auth/login", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("authentication request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("authentication failed with status %d", resp.StatusCode)
	}

	var result struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode auth response: %w", err)
	}

	c.token = result.Token
	return nil
}
func (c *UmamiClient) doRequest(path string, params map[string]string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, http.NoBody)
	if err != nil {
		return nil, err
	}

	if params != nil {
		q := req.URL.Query()
		for k, v := range params {
			q.Add(k, v)
		}
		req.URL.RawQuery = q.Encode()
	}

	if c.apiKey != "" {
		req.Header.Set("x-umami-api-key", c.apiKey)
	} else {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

type Website struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Domain    string    `json:"domain"`
	CreatedAt time.Time `json:"createdAt"`
}

func (c *UmamiClient) GetWebsites(includeTeams bool) ([]Website, error) {
	var endpoint string
	var params map[string]string

	if c.teamID != "" {
		endpoint = fmt.Sprintf("%s/teams/%s/websites", c.basePath(), c.teamID)
	} else {
		endpoint = c.websitesPath()
		if includeTeams {
			params = map[string]string{"includeTeams": "true"}
		}
	}

	data, err := c.doRequest(endpoint, params)
	if err != nil {
		return nil, err
	}

	var result struct {
		Data []Website `json:"data"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	return result.Data, nil
}

type Stats struct {
	PageViews  int              `json:"pageviews"`
	Visitors   int              `json:"visitors"`
	Visits     int              `json:"visits"`
	Bounces    int              `json:"bounces"`
	TotalTime  int              `json:"totaltime"`
	Comparison *StatsComparison `json:"comparison,omitempty"`
}

type StatsComparison struct {
	PageViews int `json:"pageviews"`
	Visitors  int `json:"visitors"`
	Visits    int `json:"visits"`
	Bounces   int `json:"bounces"`
	TotalTime int `json:"totaltime"`
}

func (c *UmamiClient) GetStats(websiteID, startDate, endDate string) (*Stats, error) {
	params := map[string]string{
		"startAt": startDate,
		"endAt":   endDate,
	}

	data, err := c.doRequest(fmt.Sprintf("%s/%s/stats", c.websitesPath(), websiteID), params)
	if err != nil {
		return nil, err
	}

	var stats Stats
	if err := json.Unmarshal(data, &stats); err != nil {
		return nil, err
	}

	return &stats, nil
}

type PageView struct {
	T string `json:"t"`
	Y int    `json:"y"`
}

func (c *UmamiClient) GetPageViews(websiteID, startDate, endDate, unit string) ([]PageView, error) {
	params := map[string]string{
		"startAt": startDate,
		"endAt":   endDate,
		"unit":    unit,
	}

	data, err := c.doRequest(fmt.Sprintf("%s/%s/pageviews", c.websitesPath(), websiteID), params)
	if err != nil {
		return nil, err
	}

	var response struct {
		PageViews []PageView `json:"pageviews"`
		Sessions  []PageView `json:"sessions"`
	}
	if err := json.Unmarshal(data, &response); err != nil {
		var pageviews []PageView
		if err2 := json.Unmarshal(data, &pageviews); err2 != nil {
			return nil, err
		}
		return pageviews, nil
	}

	return response.PageViews, nil
}

type Metric struct {
	X string `json:"x"`
	Y int    `json:"y"`
}

func (c *UmamiClient) GetMetrics(websiteID, startDate, endDate, metricType string, limit int) ([]Metric, error) {
	// Map legacy "url" type to current "path" type (renamed Oct 2025)
	if metricType == "url" {
		metricType = metricTypePath
	}

	params := map[string]string{
		"startAt": startDate,
		"endAt":   endDate,
		"type":    metricType,
		"limit":   fmt.Sprintf("%d", limit),
	}

	data, err := c.doRequest(fmt.Sprintf("%s/%s/metrics", c.websitesPath(), websiteID), params)
	if err != nil {
		return nil, err
	}

	var metrics []Metric
	if err := json.Unmarshal(data, &metrics); err != nil {
		return nil, err
	}

	return metrics, nil
}

func (c *UmamiClient) GetActive(websiteID string) ([]Metric, error) {
	data, err := c.doRequest(fmt.Sprintf("%s/%s/active", c.websitesPath(), websiteID), nil)
	if err != nil {
		return nil, err
	}

	var response []struct {
		X int `json:"x"`
		Y int `json:"y"`
	}
	if err := json.Unmarshal(data, &response); err != nil {
		var singleResponse struct {
			X int `json:"x"`
		}
		if err2 := json.Unmarshal(data, &singleResponse); err2 != nil {
			return nil, err
		}
		return []Metric{{X: fmt.Sprintf("%d", singleResponse.X), Y: singleResponse.X}}, nil
	}

	metrics := make([]Metric, len(response))
	for i, r := range response {
		metrics[i] = Metric{X: fmt.Sprintf("%d", r.X), Y: r.Y}
	}
	return metrics, nil
}

type Session struct {
	ID        string    `json:"id"`
	WebsiteID string    `json:"websiteId"`
	Hostname  string    `json:"hostname"`
	Browser   string    `json:"browser"`
	OS        string    `json:"os"`
	Device    string    `json:"device"`
	Screen    string    `json:"screen"`
	Language  string    `json:"language"`
	Country   string    `json:"country"`
	Region    string    `json:"region"`
	City      string    `json:"city"`
	FirstAt   time.Time `json:"firstAt"`
	LastAt    time.Time `json:"lastAt"`
	Visits    int       `json:"visits"`
	Views     int       `json:"views"`
	CreatedAt time.Time `json:"createdAt"`
}

type SessionList struct {
	Data     []Session `json:"data"`
	Count    int       `json:"count"`
	Page     int       `json:"page"`
	PageSize int       `json:"pageSize"`
}

func (c *UmamiClient) GetSessions(
	websiteID, startDate, endDate, search string, page, pageSize int,
) (*SessionList, error) {
	params := map[string]string{
		"startAt": startDate,
		"endAt":   endDate,
	}
	if search != "" {
		params["search"] = search
	}
	if page > 0 {
		params["page"] = fmt.Sprintf("%d", page)
	}
	if pageSize > 0 {
		params["pageSize"] = fmt.Sprintf("%d", pageSize)
	}

	data, err := c.doRequest(fmt.Sprintf("%s/%s/sessions", c.websitesPath(), websiteID), params)
	if err != nil {
		return nil, err
	}

	var result SessionList
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

type SessionStats struct {
	PageViews int `json:"pageviews"`
	Visitors  int `json:"visitors"`
	Visits    int `json:"visits"`
	Countries int `json:"countries"`
	Events    int `json:"events"`
}

type valueField struct {
	Value int `json:"value"`
}

func (c *UmamiClient) GetSessionStats(websiteID, startDate, endDate string) (*SessionStats, error) {
	params := map[string]string{
		"startAt": startDate,
		"endAt":   endDate,
	}

	data, err := c.doRequest(fmt.Sprintf("%s/%s/sessions/stats", c.websitesPath(), websiteID), params)
	if err != nil {
		return nil, err
	}

	var raw struct {
		PageViews valueField `json:"pageviews"`
		Visitors  valueField `json:"visitors"`
		Visits    valueField `json:"visits"`
		Countries valueField `json:"countries"`
		Events    valueField `json:"events"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	return &SessionStats{
		PageViews: raw.PageViews.Value,
		Visitors:  raw.Visitors.Value,
		Visits:    raw.Visits.Value,
		Countries: raw.Countries.Value,
		Events:    raw.Events.Value,
	}, nil
}

type SessionActivity struct {
	CreatedAt      time.Time `json:"createdAt"`
	URLPath        string    `json:"urlPath"`
	URLQuery       string    `json:"urlQuery"`
	ReferrerDomain string    `json:"referrerDomain"`
	EventID        string    `json:"eventId"`
	EventType      int       `json:"eventType"`
	EventName      string    `json:"eventName"`
	VisitID        string    `json:"visitId"`
	HasData        bool      `json:"hasData"`
}

func (c *UmamiClient) GetSessionActivity(websiteID, sessionID, startDate, endDate string) ([]SessionActivity, error) {
	if startDate == "" {
		startDate = "0"
	}
	if endDate == "" {
		endDate = fmt.Sprintf("%d", time.Now().UnixMilli())
	}
	params := map[string]string{
		"startAt": startDate,
		"endAt":   endDate,
	}

	data, err := c.doRequest(
		fmt.Sprintf("%s/%s/sessions/%s/activity", c.websitesPath(), websiteID, sessionID), params,
	)
	if err != nil {
		return nil, err
	}

	var activity []SessionActivity
	if err := json.Unmarshal(data, &activity); err != nil {
		return nil, err
	}
	return activity, nil
}
