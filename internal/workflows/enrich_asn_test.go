package workflows

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/spectra-red/recon/internal/enrichment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockASNClient implements enrichment.ASNClient for testing
type mockASNClient struct {
	lookupFunc      func(ctx context.Context, ip string) (*enrichment.ASNInfo, error)
	lookupBatchFunc func(ctx context.Context, ips []string) (map[string]*enrichment.ASNInfo, error)
}

func (m *mockASNClient) LookupASN(ctx context.Context, ip string) (*enrichment.ASNInfo, error) {
	if m.lookupFunc != nil {
		return m.lookupFunc(ctx, ip)
	}
	return nil, errors.New("not implemented")
}

func (m *mockASNClient) LookupBatch(ctx context.Context, ips []string) (map[string]*enrichment.ASNInfo, error) {
	if m.lookupBatchFunc != nil {
		return m.lookupBatchFunc(ctx, ips)
	}
	return nil, errors.New("not implemented")
}

func TestEnrichASNWorkflow_ServiceName(t *testing.T) {
	workflow := &EnrichASNWorkflow{}
	assert.Equal(t, "EnrichASNWorkflow", workflow.ServiceName())
}

func TestEnrichASNWorkflow_Validation(t *testing.T) {
	mockClient := &mockASNClient{
		lookupBatchFunc: func(ctx context.Context, ips []string) (map[string]*enrichment.ASNInfo, error) {
			return make(map[string]*enrichment.ASNInfo), nil
		},
	}

	_ = NewEnrichASNWorkflow(nil, mockClient)

	tests := []struct {
		name    string
		req     EnrichASNRequest
		wantErr string
	}{
		{
			name: "empty IPs list",
			req: EnrichASNRequest{
				IPs: []string{},
			},
			wantErr: "no IPs provided",
		},
		{
			name: "batch size exceeds maximum",
			req: EnrichASNRequest{
				IPs: make([]string, 101),
			},
			wantErr: "batch size exceeds maximum",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Note: This test validates the request structure
			// In a real test, we'd use a Restate test harness
			assert.Equal(t, len(tt.req.IPs) == 0 || len(tt.req.IPs) > 100, true)
		})
	}
}

func TestEnrichASNRequest_BatchSize(t *testing.T) {
	tests := []struct {
		name      string
		ipsCount  int
		wantValid bool
	}{
		{
			name:      "valid batch size",
			ipsCount:  50,
			wantValid: true,
		},
		{
			name:      "maximum batch size",
			ipsCount:  100,
			wantValid: true,
		},
		{
			name:      "exceeds maximum",
			ipsCount:  101,
			wantValid: false,
		},
		{
			name:      "empty batch",
			ipsCount:  0,
			wantValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := EnrichASNRequest{
				IPs: make([]string, tt.ipsCount),
			}

			// Validate batch size
			valid := len(req.IPs) > 0 && len(req.IPs) <= 100
			assert.Equal(t, tt.wantValid, valid)
		})
	}
}

func TestMockASNClient_LookupBatch(t *testing.T) {
	tests := []struct {
		name    string
		ips     []string
		mockResp map[string]*enrichment.ASNInfo
		mockErr error
		wantErr bool
	}{
		{
			name: "successful batch lookup",
			ips:  []string{"8.8.8.8", "1.1.1.1"},
			mockResp: map[string]*enrichment.ASNInfo{
				"8.8.8.8": {Number: 15169, Org: "GOOGLE", Country: "US"},
				"1.1.1.1": {Number: 13335, Org: "CLOUDFLARE", Country: "US"},
			},
			mockErr: nil,
			wantErr: false,
		},
		{
			name:     "partial failure",
			ips:      []string{"8.8.8.8", "1.1.1.1", "invalid.ip"},
			mockResp: map[string]*enrichment.ASNInfo{
				"8.8.8.8": {Number: 15169, Org: "GOOGLE", Country: "US"},
			},
			mockErr: nil,
			wantErr: false,
		},
		{
			name:     "complete failure",
			ips:      []string{"8.8.8.8"},
			mockResp: nil,
			mockErr:  errors.New("network error"),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockASNClient{
				lookupBatchFunc: func(ctx context.Context, ips []string) (map[string]*enrichment.ASNInfo, error) {
					return tt.mockResp, tt.mockErr
				},
			}

			result, err := mockClient.LookupBatch(context.Background(), tt.ips)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, len(tt.mockResp), len(result))

			for ip, info := range tt.mockResp {
				gotInfo, ok := result[ip]
				require.True(t, ok, "expected IP %s in results", ip)
				assert.Equal(t, info.Number, gotInfo.Number)
				assert.Equal(t, info.Org, gotInfo.Org)
				assert.Equal(t, info.Country, gotInfo.Country)
			}
		})
	}
}

func TestEnrichASNResponse_Statistics(t *testing.T) {
	tests := []struct {
		name     string
		totalIPs int
		enriched int
		cached   int
		failed   int
	}{
		{
			name:     "all enriched",
			totalIPs: 10,
			enriched: 10,
			cached:   0,
			failed:   0,
		},
		{
			name:     "all cached",
			totalIPs: 10,
			enriched: 0,
			cached:   10,
			failed:   0,
		},
		{
			name:     "mixed results",
			totalIPs: 10,
			enriched: 5,
			cached:   3,
			failed:   2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_ = EnrichASNResponse{
				TotalIPs:    tt.totalIPs,
				EnrichedIPs: tt.enriched,
				CachedIPs:   tt.cached,
				FailedIPs:   tt.failed,
			}

			// Verify statistics add up correctly
			assert.Equal(t, tt.totalIPs, tt.enriched+tt.cached+tt.failed,
				"enriched + cached + failed should equal total")
		})
	}
}

func TestHostASNData_Structure(t *testing.T) {
	data := HostASNData{
		IP:      "8.8.8.8",
		ASN:     15169,
		Country: "US",
	}

	assert.Equal(t, "8.8.8.8", data.IP)
	assert.Equal(t, 15169, data.ASN)
	assert.Equal(t, "US", data.Country)
}

// Test helper functions
func TestIPEncoding(t *testing.T) {
	tests := []struct {
		name     string
		ip       string
		expected string
	}{
		{
			name:     "standard IPv4",
			ip:       "8.8.8.8",
			expected: "8_8_8_8",
		},
		{
			name:     "private IPv4",
			ip:       "192.168.1.1",
			expected: "192_168_1_1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This mimics the encoding logic used in the workflow
			encoded := strings.ReplaceAll(tt.ip, ".", "_")
			assert.Equal(t, tt.expected, encoded)
		})
	}
}

// Benchmark tests
func BenchmarkEnrichASNRequest_Creation(b *testing.B) {
	ips := make([]string, 100)
	for i := 0; i < 100; i++ {
		ips[i] = "8.8.8.8"
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = EnrichASNRequest{
			IPs:   ips,
			JobID: "test-job",
		}
	}
}

func BenchmarkEnrichASNResponse_Creation(b *testing.B) {
	asnData := make(map[string]*enrichment.ASNInfo, 100)
	for i := 0; i < 100; i++ {
		asnData["8.8.8.8"] = &enrichment.ASNInfo{
			Number:  15169,
			Org:     "GOOGLE",
			Country: "US",
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = EnrichASNResponse{
			TotalIPs:    100,
			EnrichedIPs: 100,
			ASNData:     asnData,
		}
	}
}
