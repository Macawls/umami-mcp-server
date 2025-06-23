package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type UmamiClient struct {
	baseURL    string
	username   string
	password   string
	token      string
	httpClient *http.Client
}

func NewUmamiClient(baseURL, username, password string) *UmamiClient {
	return &UmamiClient{
		baseURL:    baseURL,
		username:   username,
		password:   password,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *UmamiClient) Authenticate() error {
	payload := map[string]string{
		"username": c.username,
		"password": c.password,
	}

	data, _ := json.Marshal(payload)
	resp, err := c.httpClient.Post(
		c.baseURL+"/api/auth/login",
		"application/json",
		bytes.NewReader(data),
	)
	if err != nil {
		return fmt.Errorf("authentication request failed: %w", err)
	}
	defer resp.Body.Close()

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
func (c *UmamiClient) doRequest(method, path string, params map[string]string) ([]byte, error) {
	req, err := http.NewRequest(method, c.baseURL+path, nil)
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

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

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

func (c *UmamiClient) GetWebsites() ([]Website, error) {
	data, err := c.doRequest("GET", "/api/websites", nil)
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
	PageViews  ValueChange `json:"pageviews"`
	Visitors   ValueChange `json:"visitors"`
	Bounces    ValueChange `json:"bounces"`
	TotalTime  ValueChange `json:"totaltime"`
}

type ValueChange struct {
	Value  int `json:"value"`
	Change int `json:"change"`
}

func (c *UmamiClient) GetStats(websiteID, startDate, endDate string) (*Stats, error) {
	params := map[string]string{
		"startAt": startDate,
		"endAt":   endDate,
	}

	data, err := c.doRequest("GET", fmt.Sprintf("/api/websites/%s/stats", websiteID), params)
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

	data, err := c.doRequest("GET", fmt.Sprintf("/api/websites/%s/pageviews", websiteID), params)
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
	params := map[string]string{
		"startAt": startDate,
		"endAt":   endDate,
		"type":    metricType,
		"limit":   fmt.Sprintf("%d", limit),
	}

	data, err := c.doRequest("GET", fmt.Sprintf("/api/websites/%s/metrics", websiteID), params)
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
	data, err := c.doRequest("GET", fmt.Sprintf("/api/websites/%s/active", websiteID), nil)
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