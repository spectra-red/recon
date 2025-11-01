package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/spectra-red/recon/internal/models"
)

// QueryClient handles API queries to the Spectra-Red backend
type QueryClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewQueryClient creates a new query client
func NewQueryClient(baseURL string) *QueryClient {
	return &QueryClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// NewQueryClientWithTimeout creates a query client with a custom timeout
func NewQueryClientWithTimeout(baseURL string, timeout time.Duration) *QueryClient {
	return &QueryClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// QueryHost queries host information by IP address
func (c *QueryClient) QueryHost(ctx context.Context, ip string, depth int) (*models.HostQueryResponse, error) {
	url := fmt.Sprintf("%s/v1/query/host/%s?depth=%d", c.baseURL, ip, depth)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var result models.HostQueryResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// GraphQuery executes a graph traversal query
func (c *QueryClient) GraphQuery(ctx context.Context, req *models.GraphQueryRequest) (*models.GraphQueryResponse, error) {
	url := fmt.Sprintf("%s/v1/query/graph", c.baseURL)

	// Validate request
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Marshal request body
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var result models.GraphQueryResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// SimilarQuery performs a vector similarity search
func (c *QueryClient) SimilarQuery(ctx context.Context, req *models.SimilarRequest) (*models.SimilarResponse, error) {
	url := fmt.Sprintf("%s/v1/query/similar", c.baseURL)

	// Validate request
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Marshal request body
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var result models.SimilarResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// HostQueryOptions contains options for host queries
type HostQueryOptions struct {
	IP    string
	Depth int
}

// GraphQueryByASN creates a graph query by ASN
func GraphQueryByASN(asn int, limit, offset int) *models.GraphQueryRequest {
	return &models.GraphQueryRequest{
		QueryType: models.QueryByASN,
		ASN:       &asn,
		Limit:     limit,
		Offset:    offset,
	}
}

// GraphQueryByLocation creates a graph query by location
func GraphQueryByLocation(city, region, country string, limit, offset int) *models.GraphQueryRequest {
	return &models.GraphQueryRequest{
		QueryType: models.QueryByLocation,
		City:      city,
		Region:    region,
		Country:   country,
		Limit:     limit,
		Offset:    offset,
	}
}

// GraphQueryByVuln creates a graph query by vulnerability
func GraphQueryByVuln(cve string, limit, offset int) *models.GraphQueryRequest {
	return &models.GraphQueryRequest{
		QueryType: models.QueryByVuln,
		CVE:       cve,
		Limit:     limit,
		Offset:    offset,
	}
}

// GraphQueryByService creates a graph query by service
func GraphQueryByService(product, service string, limit, offset int) *models.GraphQueryRequest {
	return &models.GraphQueryRequest{
		QueryType: models.QueryByService,
		Product:   product,
		Service:   service,
		Limit:     limit,
		Offset:    offset,
	}
}

// NewSimilarRequest creates a similarity search request
func NewSimilarRequest(query string, k int) *models.SimilarRequest {
	if k <= 0 {
		k = models.DefaultK
	}
	return &models.SimilarRequest{
		Query: query,
		K:     &k,
	}
}
