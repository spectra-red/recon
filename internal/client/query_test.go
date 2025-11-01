package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/spectra-red/recon/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewQueryClient(t *testing.T) {
	client := NewQueryClient("http://localhost:3000")
	assert.NotNil(t, client)
	assert.Equal(t, "http://localhost:3000", client.baseURL)
	assert.Equal(t, 30*time.Second, client.httpClient.Timeout)
}

func TestNewQueryClientWithTimeout(t *testing.T) {
	timeout := 10 * time.Second
	client := NewQueryClientWithTimeout("http://localhost:3000", timeout)
	assert.NotNil(t, client)
	assert.Equal(t, timeout, client.httpClient.Timeout)
}

func TestQueryHost_Success(t *testing.T) {
	// Create mock response
	mockResponse := &models.HostQueryResponse{
		IP:        "1.2.3.4",
		ASN:       15169,
		City:      "Mountain View",
		Country:   "United States",
		FirstSeen: time.Now().Add(-24 * time.Hour),
		LastSeen:  time.Now(),
		Ports: []models.PortDetail{
			{
				Number:   80,
				Protocol: "tcp",
				Services: []models.ServiceDetail{
					{
						Name:    "http",
						Product: "nginx",
						Version: "1.25.1",
					},
				},
			},
		},
	}

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/query/host/1.2.3.4", r.URL.Path)
		assert.Equal(t, "depth=2", r.URL.RawQuery)
		assert.Equal(t, "application/json", r.Header.Get("Accept"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	// Create client
	client := NewQueryClient(server.URL)

	// Execute query
	ctx := context.Background()
	result, err := client.QueryHost(ctx, "1.2.3.4", 2)

	// Assertions
	require.NoError(t, err)
	assert.Equal(t, "1.2.3.4", result.IP)
	assert.Equal(t, 15169, result.ASN)
	assert.Equal(t, "Mountain View", result.City)
	assert.Len(t, result.Ports, 1)
	assert.Equal(t, 80, result.Ports[0].Number)
}

func TestQueryHost_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":"host not found"}`))
	}))
	defer server.Close()

	client := NewQueryClient(server.URL)
	ctx := context.Background()
	result, err := client.QueryHost(ctx, "1.2.3.4", 2)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "404")
}

func TestGraphQuery_ByASN(t *testing.T) {
	asn := 15169
	mockResponse := &models.GraphQueryResponse{
		Results: []models.HostResult{
			{
				IP:      "1.2.3.4",
				ASN:     15169,
				City:    "Mountain View",
				Country: "United States",
			},
			{
				IP:      "5.6.7.8",
				ASN:     15169,
				City:    "San Francisco",
				Country: "United States",
			},
		},
		Pagination: models.PaginationMetadata{
			Limit:   100,
			Offset:  0,
			Total:   2,
			HasMore: false,
		},
		QueryTime: 123.45,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/query/graph", r.URL.Path)
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		var req models.GraphQueryRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		assert.Equal(t, models.QueryByASN, req.QueryType)
		assert.NotNil(t, req.ASN)
		assert.Equal(t, 15169, *req.ASN)
		assert.Equal(t, 100, req.Limit)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	client := NewQueryClient(server.URL)
	req := GraphQueryByASN(asn, 100, 0)

	ctx := context.Background()
	result, err := client.GraphQuery(ctx, req)

	require.NoError(t, err)
	assert.Len(t, result.Results, 2)
	assert.Equal(t, "1.2.3.4", result.Results[0].IP)
	assert.Equal(t, 15169, result.Results[0].ASN)
	assert.Equal(t, 123.45, result.QueryTime)
}

func TestGraphQuery_ByLocation(t *testing.T) {
	mockResponse := &models.GraphQueryResponse{
		Results: []models.HostResult{
			{
				IP:      "1.2.3.4",
				City:    "Paris",
				Country: "France",
			},
		},
		Pagination: models.PaginationMetadata{
			Limit:   100,
			Offset:  0,
			Total:   1,
			HasMore: false,
		},
		QueryTime: 98.76,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req models.GraphQueryRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		assert.Equal(t, models.QueryByLocation, req.QueryType)
		assert.Equal(t, "Paris", req.City)
		assert.Equal(t, "France", req.Country)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	client := NewQueryClient(server.URL)
	req := GraphQueryByLocation("Paris", "", "France", 100, 0)

	ctx := context.Background()
	result, err := client.GraphQuery(ctx, req)

	require.NoError(t, err)
	assert.Len(t, result.Results, 1)
	assert.Equal(t, "Paris", result.Results[0].City)
}

func TestGraphQuery_ByVuln(t *testing.T) {
	mockResponse := &models.GraphQueryResponse{
		Results: []models.HostResult{
			{
				IP:  "1.2.3.4",
				ASN: 15169,
			},
		},
		Pagination: models.PaginationMetadata{
			Limit:   50,
			Offset:  0,
			Total:   1,
			HasMore: false,
		},
		QueryTime: 150.0,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req models.GraphQueryRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		assert.Equal(t, models.QueryByVuln, req.QueryType)
		assert.Equal(t, "CVE-2024-1234", req.CVE)
		assert.Equal(t, 50, req.Limit)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	client := NewQueryClient(server.URL)
	req := GraphQueryByVuln("CVE-2024-1234", 50, 0)

	ctx := context.Background()
	result, err := client.GraphQuery(ctx, req)

	require.NoError(t, err)
	assert.Len(t, result.Results, 1)
	assert.Equal(t, 150.0, result.QueryTime)
}

func TestGraphQuery_ByService(t *testing.T) {
	mockResponse := &models.GraphQueryResponse{
		Results: []models.HostResult{
			{
				IP: "1.2.3.4",
				Services: []models.Service{
					{
						Name:    "http",
						Product: "nginx",
						Version: "1.25.1",
					},
				},
			},
		},
		Pagination: models.PaginationMetadata{
			Limit:   100,
			Offset:  0,
			Total:   1,
			HasMore: false,
		},
		QueryTime: 87.5,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req models.GraphQueryRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		assert.Equal(t, models.QueryByService, req.QueryType)
		assert.Equal(t, "nginx", req.Product)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	client := NewQueryClient(server.URL)
	req := GraphQueryByService("nginx", "", 100, 0)

	ctx := context.Background()
	result, err := client.GraphQuery(ctx, req)

	require.NoError(t, err)
	assert.Len(t, result.Results, 1)
	assert.Len(t, result.Results[0].Services, 1)
	assert.Equal(t, "nginx", result.Results[0].Services[0].Product)
}

func TestSimilarQuery_Success(t *testing.T) {
	mockResponse := &models.SimilarResponse{
		Query: "nginx remote code execution",
		Results: []models.VulnResult{
			{
				CVEID:   "CVE-2024-1234",
				Title:   "Nginx Buffer Overflow",
				Summary: "Remote code execution vulnerability in nginx",
				CVSS:    9.8,
				Score:   0.95,
			},
			{
				CVEID:   "CVE-2024-5678",
				Title:   "Nginx Authentication Bypass",
				Summary: "Authentication bypass leading to RCE",
				CVSS:    8.1,
				Score:   0.87,
			},
		},
		Count:     2,
		Timestamp: time.Now().Format(time.RFC3339),
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/query/similar", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		var req models.SimilarRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		assert.Equal(t, "nginx remote code execution", req.Query)
		assert.NotNil(t, req.K)
		assert.Equal(t, 10, *req.K)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	client := NewQueryClient(server.URL)
	req := NewSimilarRequest("nginx remote code execution", 10)

	ctx := context.Background()
	result, err := client.SimilarQuery(ctx, req)

	require.NoError(t, err)
	assert.Equal(t, "nginx remote code execution", result.Query)
	assert.Equal(t, 2, result.Count)
	assert.Len(t, result.Results, 2)
	assert.Equal(t, "CVE-2024-1234", result.Results[0].CVEID)
	assert.Equal(t, 0.95, result.Results[0].Score)
}

func TestSimilarQuery_ValidationError(t *testing.T) {
	client := NewQueryClient("http://localhost:3000")

	// Empty query
	req := NewSimilarRequest("", 10)
	ctx := context.Background()
	result, err := client.SimilarQuery(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "invalid request")
}

func TestNewSimilarRequest_DefaultK(t *testing.T) {
	req := NewSimilarRequest("test query", 0)
	assert.Equal(t, "test query", req.Query)
	assert.NotNil(t, req.K)
	assert.Equal(t, models.DefaultK, *req.K)
}

func TestGraphQueryHelpers(t *testing.T) {
	t.Run("GraphQueryByASN", func(t *testing.T) {
		req := GraphQueryByASN(15169, 100, 0)
		assert.Equal(t, models.QueryByASN, req.QueryType)
		assert.NotNil(t, req.ASN)
		assert.Equal(t, 15169, *req.ASN)
		assert.Equal(t, 100, req.Limit)
		assert.Equal(t, 0, req.Offset)
	})

	t.Run("GraphQueryByLocation", func(t *testing.T) {
		req := GraphQueryByLocation("Paris", "Ile-de-France", "France", 50, 10)
		assert.Equal(t, models.QueryByLocation, req.QueryType)
		assert.Equal(t, "Paris", req.City)
		assert.Equal(t, "Ile-de-France", req.Region)
		assert.Equal(t, "France", req.Country)
		assert.Equal(t, 50, req.Limit)
		assert.Equal(t, 10, req.Offset)
	})

	t.Run("GraphQueryByVuln", func(t *testing.T) {
		req := GraphQueryByVuln("CVE-2024-1234", 200, 50)
		assert.Equal(t, models.QueryByVuln, req.QueryType)
		assert.Equal(t, "CVE-2024-1234", req.CVE)
		assert.Equal(t, 200, req.Limit)
		assert.Equal(t, 50, req.Offset)
	})

	t.Run("GraphQueryByService", func(t *testing.T) {
		req := GraphQueryByService("nginx", "http", 75, 25)
		assert.Equal(t, models.QueryByService, req.QueryType)
		assert.Equal(t, "nginx", req.Product)
		assert.Equal(t, "http", req.Service)
		assert.Equal(t, 75, req.Limit)
		assert.Equal(t, 25, req.Offset)
	})
}

func TestQueryClient_Timeout(t *testing.T) {
	// Create a slow server that takes longer than the timeout
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Create client with short timeout
	client := NewQueryClientWithTimeout(server.URL, 100*time.Millisecond)

	ctx := context.Background()
	result, err := client.QueryHost(ctx, "1.2.3.4", 2)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "deadline exceeded")
}
