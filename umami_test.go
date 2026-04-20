package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUmamiClient_Authenticate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/auth/login" {
			t.Errorf("Expected path /api/auth/login, got %s", r.URL.Path)
		}

		var req map[string]string
		_ = json.NewDecoder(r.Body).Decode(&req)

		if req["username"] != "testuser" || req["password"] != "testpass" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"token": "test-token-123"})
	}))
	defer server.Close()

	client := NewUmamiClient(server.URL, "testuser", "testpass")

	err := client.Authenticate()
	if err != nil {
		t.Fatalf("Authentication failed: %v", err)
	}

	if client.token != "test-token-123" {
		t.Errorf("Expected token test-token-123, got %s", client.token)
	}
}

func TestUmamiClient_APIKey_NoLogin(t *testing.T) {
	loginCalled := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/auth/login" {
			loginCalled = true
			w.WriteHeader(http.StatusNotFound)
			return
		}

		if r.URL.Path != "/v1/websites" {
			t.Errorf("Expected path /v1/websites, got %s", r.URL.Path)
		}
		if r.Header.Get("x-umami-api-key") != "cloud-key" {
			t.Errorf("Expected x-umami-api-key=cloud-key, got %s", r.Header.Get("x-umami-api-key"))
		}
		if r.Header.Get("Authorization") != "" {
			t.Errorf("Expected no Authorization header, got %s", r.Header.Get("Authorization"))
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{
					"id":        "cloud-site-1",
					"name":      "Cloud Site",
					"domain":    "cloud.example.com",
					"createdAt": "2025-01-01T00:00:00Z",
				},
			},
		})
	}))
	defer server.Close()

	client := NewUmamiClientWithAPIKey(server.URL, "cloud-key")

	if err := client.Authenticate(); err != nil {
		t.Fatalf("Authenticate should be a no-op for API key mode, got: %v", err)
	}
	if loginCalled {
		t.Error("Expected no login request in API key mode")
	}

	websites, err := client.GetWebsites(false)
	if err != nil {
		t.Fatalf("GetWebsites failed: %v", err)
	}
	if len(websites) != 1 || websites[0].ID != "cloud-site-1" {
		t.Errorf("Unexpected websites: %+v", websites)
	}
}

func TestUmamiClient_GetWebsites(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/websites" {
			t.Errorf("Expected path /api/websites, got %s", r.URL.Path)
		}

		if r.Header.Get("Authorization") != "Bearer test-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{
					"id":        "test-id-1",
					"name":      "Test Site 1",
					"domain":    "test1.com",
					"createdAt": "2025-01-01T00:00:00Z",
				},
				{
					"id":        "test-id-2",
					"name":      "Test Site 2",
					"domain":    "test2.com",
					"createdAt": "2025-01-02T00:00:00Z",
				},
			},
		})
	}))
	defer server.Close()

	client := &UmamiClient{
		baseURL:    server.URL,
		token:      "test-token",
		httpClient: &http.Client{},
	}

	websites, err := client.GetWebsites(false)
	if err != nil {
		t.Fatalf("GetWebsites failed: %v", err)
	}

	if len(websites) != 2 {
		t.Errorf("Expected 2 websites, got %d", len(websites))
	}

	if websites[0].ID != "test-id-1" {
		t.Errorf("Expected first website ID test-id-1, got %s", websites[0].ID)
	}
}

func TestUmamiClient_GetWebsites_IncludeTeams(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/websites" {
			t.Errorf("Expected path /api/websites, got %s", r.URL.Path)
		}

		includeTeams := r.URL.Query().Get("includeTeams")
		if includeTeams != "true" {
			t.Errorf("Expected includeTeams=true, got %s", includeTeams)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{
					"id":        "test-id-1",
					"name":      "Team Site",
					"domain":    "team.com",
					"createdAt": "2025-01-01T00:00:00Z",
				},
			},
		})
	}))
	defer server.Close()

	client := &UmamiClient{
		baseURL:    server.URL,
		token:      "test-token",
		httpClient: &http.Client{},
	}

	websites, err := client.GetWebsites(true)
	if err != nil {
		t.Fatalf("GetWebsites with includeTeams failed: %v", err)
	}

	if len(websites) != 1 {
		t.Errorf("Expected 1 website, got %d", len(websites))
	}

	if websites[0].Name != "Team Site" {
		t.Errorf("Expected website name 'Team Site', got %s", websites[0].Name)
	}
}

func TestUmamiClient_GetWebsites_TeamID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/teams/my-team-123/websites" {
			t.Errorf("Expected path /api/teams/my-team-123/websites, got %s", r.URL.Path)
		}

		if r.URL.Query().Get("includeTeams") != "" {
			t.Errorf("Expected no includeTeams param when teamID is set")
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{
					"id":        "team-site-1",
					"name":      "Team Website",
					"domain":    "team.example.com",
					"createdAt": "2025-01-01T00:00:00Z",
				},
			},
		})
	}))
	defer server.Close()

	client := &UmamiClient{
		baseURL:    server.URL,
		token:      "test-token",
		teamID:     "my-team-123",
		httpClient: &http.Client{},
	}

	websites, err := client.GetWebsites(true)
	if err != nil {
		t.Fatalf("GetWebsites with teamID failed: %v", err)
	}

	if len(websites) != 1 {
		t.Errorf("Expected 1 website, got %d", len(websites))
	}

	if websites[0].ID != "team-site-1" {
		t.Errorf("Expected website ID 'team-site-1', got %s", websites[0].ID)
	}
}

func TestUmamiClient_GetStats(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/websites/test-website-id/stats" {
			t.Errorf("Expected path /api/websites/test-website-id/stats, got %s", r.URL.Path)
		}

		startAt := r.URL.Query().Get("startAt")
		endAt := r.URL.Query().Get("endAt")

		if startAt != "1234567890" || endAt != "1234567899" {
			t.Errorf("Invalid date params: startAt=%s, endAt=%s", startAt, endAt)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"pageviews": 15171,
			"visitors":  4415,
			"visits":    5680,
			"bounces":   3567,
			"totaltime": 809968,
			"comparison": map[string]int{
				"pageviews": 38675,
				"visitors":  12000,
				"visits":    15000,
				"bounces":   8000,
				"totaltime": 2000000,
			},
		})
	}))
	defer server.Close()

	client := &UmamiClient{
		baseURL:    server.URL,
		token:      "test-token",
		httpClient: &http.Client{},
	}

	stats, err := client.GetStats("test-website-id", "1234567890", "1234567899")
	if err != nil {
		t.Fatalf("GetStats failed: %v", err)
	}

	if stats.PageViews != 15171 {
		t.Errorf("Expected 15171 pageviews, got %d", stats.PageViews)
	}

	if stats.Visitors != 4415 {
		t.Errorf("Expected 4415 visitors, got %d", stats.Visitors)
	}

	if stats.Visits != 5680 {
		t.Errorf("Expected 5680 visits, got %d", stats.Visits)
	}

	if stats.Bounces != 3567 {
		t.Errorf("Expected 3567 bounces, got %d", stats.Bounces)
	}

	if stats.TotalTime != 809968 {
		t.Errorf("Expected 809968 totaltime, got %d", stats.TotalTime)
	}

	if stats.Comparison == nil {
		t.Fatal("Expected comparison to be present")
	}

	if stats.Comparison.PageViews != 38675 {
		t.Errorf("Expected comparison pageviews 38675, got %d", stats.Comparison.PageViews)
	}
}

func TestUmamiClient_GetStats_WithoutComparison(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"pageviews": 100,
			"visitors":  50,
			"visits":    60,
			"bounces":   20,
			"totaltime": 5000,
		})
	}))
	defer server.Close()

	client := &UmamiClient{
		baseURL:    server.URL,
		token:      "test-token",
		httpClient: &http.Client{},
	}

	stats, err := client.GetStats("test-website-id", "1234567890", "1234567899")
	if err != nil {
		t.Fatalf("GetStats failed: %v", err)
	}

	if stats.PageViews != 100 {
		t.Errorf("Expected 100 pageviews, got %d", stats.PageViews)
	}

	if stats.Comparison != nil {
		t.Error("Expected comparison to be nil when not present in response")
	}
}

func TestUmamiClient_GetMetrics_UrlToPathMapping(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/websites/test-website-id/metrics" {
			t.Errorf("Expected metrics path, got %s", r.URL.Path)
		}

		metricType := r.URL.Query().Get("type")
		if metricType != "path" {
			t.Errorf("Expected type=path (mapped from url), got %s", metricType)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[
			{"x": "/blog/post1", "y": 150},
			{"x": "/about", "y": 80}
		]`))
	}))
	defer server.Close()

	client := &UmamiClient{
		baseURL:    server.URL,
		token:      "test-token",
		httpClient: &http.Client{},
	}

	metrics, err := client.GetMetrics("test-website-id", "1234567890", "1234567899", "url", 10)
	if err != nil {
		t.Fatalf("GetMetrics failed: %v", err)
	}

	if len(metrics) != 2 {
		t.Errorf("Expected 2 metrics, got %d", len(metrics))
	}
}

func TestUmamiClient_GetMetrics_DirectArray(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/websites/test-website-id/metrics" {
			t.Errorf("Expected metrics path, got %s", r.URL.Path)
		}

		metricType := r.URL.Query().Get("type")
		if metricType != "path" {
			t.Errorf("Expected type=path, got %s", metricType)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[
			{"x": "/blog/post1", "y": 150},
			{"x": "/blog/post2", "y": 120},
			{"x": "/about", "y": 80}
		]`))
	}))
	defer server.Close()

	client := &UmamiClient{
		baseURL:    server.URL,
		token:      "test-token",
		httpClient: &http.Client{},
	}

	metrics, err := client.GetMetrics("test-website-id", "1234567890", "1234567899", "path", 10)
	if err != nil {
		t.Fatalf("GetMetrics failed: %v", err)
	}

	if len(metrics) != 3 {
		t.Errorf("Expected 3 metrics, got %d", len(metrics))
	}

	if metrics[0].X != "/blog/post1" || metrics[0].Y != 150 {
		t.Errorf("Expected /blog/post1 with 150 views, got %s with %d", metrics[0].X, metrics[0].Y)
	}
}

func TestUmamiClient_GetMetrics_WrappedInData(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"data": [
				{"x": "/blog/post1", "y": 150},
				{"x": "/blog/post2", "y": 120}
			]
		}`))
	}))
	defer server.Close()

	client := &UmamiClient{
		baseURL:    server.URL,
		token:      "test-token",
		httpClient: &http.Client{},
	}

	metrics, err := client.GetMetrics("test-website-id", "1234567890", "1234567899", "path", 10)
	if err == nil {
		t.Error("Expected error for wrapped data format, got nil")
	}

	if len(metrics) != 0 {
		t.Errorf("Expected 0 metrics on error, got %d", len(metrics))
	}
}

func TestUmamiClient_GetPageViews_ObjectResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/websites/test-website-id/pageviews" {
			t.Errorf("Expected pageviews path, got %s", r.URL.Path)
		}

		unit := r.URL.Query().Get("unit")
		if unit != "day" {
			t.Errorf("Expected unit=day, got %s", unit)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"pageviews": [
				{"t": "2025-01-01", "y": 100},
				{"t": "2025-01-02", "y": 150},
				{"t": "2025-01-03", "y": 200}
			],
			"sessions": []
		}`))
	}))
	defer server.Close()

	client := &UmamiClient{
		baseURL:    server.URL,
		token:      "test-token",
		httpClient: &http.Client{},
	}

	pageviews, err := client.GetPageViews("test-website-id", "1234567890", "1234567899", "day")
	if err != nil {
		t.Fatalf("GetPageViews failed: %v", err)
	}

	if len(pageviews) != 3 {
		t.Errorf("Expected 3 pageviews, got %d", len(pageviews))
	}

	if pageviews[0].T != "2025-01-01" || pageviews[0].Y != 100 {
		t.Errorf("Expected 2025-01-01 with 100 views, got %s with %d", pageviews[0].T, pageviews[0].Y)
	}
}

func TestUmamiClient_GetActive_SingleValue(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/websites/test-website-id/active" {
			t.Errorf("Expected active path, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"x": 10}`))
	}))
	defer server.Close()

	client := &UmamiClient{
		baseURL:    server.URL,
		token:      "test-token",
		httpClient: &http.Client{},
	}

	active, err := client.GetActive("test-website-id")
	if err != nil {
		t.Fatalf("GetActive failed: %v", err)
	}

	if len(active) != 1 {
		t.Errorf("Expected 1 active metric, got %d", len(active))
	}

	if active[0].X != "10" || active[0].Y != 10 {
		t.Errorf("Expected 10 active users, got %s with %d", active[0].X, active[0].Y)
	}
}

func TestUmamiClient_ErrorHandling(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		response   string
		expectErr  bool
	}{
		{
			name:       "404 Not Found",
			statusCode: 404,
			response:   "Not Found",
			expectErr:  true,
		},
		{
			name:       "500 Server Error",
			statusCode: 500,
			response:   "Internal Server Error",
			expectErr:  true,
		},
		{
			name:       "Invalid JSON",
			statusCode: 200,
			response:   "{invalid json",
			expectErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				_, _ = w.Write([]byte(tt.response))
			}))
			defer server.Close()

			client := &UmamiClient{
				baseURL:    server.URL,
				token:      "test-token",
				httpClient: &http.Client{},
			}

			_, err := client.GetWebsites(false)
			if (err != nil) != tt.expectErr {
				t.Errorf("Expected error=%v, got error=%v", tt.expectErr, err != nil)
			}
		})
	}
}
