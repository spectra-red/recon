package workflows

import (
	"testing"

	"github.com/spectra-red/recon/internal/enrichment"
)

func TestEnrichCPEWorkflow_ServiceName(t *testing.T) {
	workflow := &EnrichCPEWorkflow{}

	name := workflow.ServiceName()
	if name != "EnrichCPEWorkflow" {
		t.Errorf("ServiceName() = %v, want EnrichCPEWorkflow", name)
	}
}

func TestNewEnrichCPEWorkflow(t *testing.T) {
	// Test without API key
	workflow := NewEnrichCPEWorkflow(nil, "")
	if workflow == nil {
		t.Fatal("NewEnrichCPEWorkflow() returned nil")
	}

	if workflow.nvdClient == nil {
		t.Error("nvdClient is nil")
	}

	// Test with API key
	workflow = NewEnrichCPEWorkflow(nil, "test-key")
	if workflow == nil {
		t.Fatal("NewEnrichCPEWorkflow() returned nil")
	}

	if workflow.nvdClient == nil {
		t.Error("nvdClient is nil")
	}
}

func TestEnrichCPERequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		req     EnrichCPERequest
		wantErr bool
	}{
		{
			name: "valid request",
			req: EnrichCPERequest{
				Services: []enrichment.ServiceInfo{
					{
						ID:      "service:test",
						Product: "nginx",
						Version: "1.24.0",
					},
				},
				BatchID: "batch-001",
			},
			wantErr: false,
		},
		{
			name: "empty services",
			req: EnrichCPERequest{
				Services: []enrichment.ServiceInfo{},
				BatchID:  "batch-002",
			},
			wantErr: false, // Empty is allowed, workflow will just process zero services
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Basic validation - ensure fields are accessible
			if tt.req.BatchID == "" && tt.name == "valid request" {
				t.Error("BatchID should not be empty for valid request")
			}

			if len(tt.req.Services) == 0 && tt.name == "valid request" {
				t.Error("Services should not be empty for valid request")
			}
		})
	}
}

func TestServiceFilter(t *testing.T) {
	filter := ServiceFilter{
		MinLastSeen:    nil,
		OnlyMissingCPE: true,
		Limit:          50,
	}

	if filter.Limit != 50 {
		t.Errorf("Limit = %d, want 50", filter.Limit)
	}

	if !filter.OnlyMissingCPE {
		t.Error("OnlyMissingCPE should be true")
	}
}

// Note: Full integration tests require a real SurrealDB instance and Restate runtime
// These would be implemented in a separate integration test suite
// Here we test the basic structure and logic

func TestEnrichCPEResponse_Structure(t *testing.T) {
	resp := EnrichCPEResponse{
		BatchID:              "batch-001",
		ServicesProcessed:    10,
		CPEsGenerated:        15,
		VulnsFound:           5,
		RelationshipsCreated: 12,
	}

	if resp.BatchID != "batch-001" {
		t.Errorf("BatchID = %v, want batch-001", resp.BatchID)
	}

	if resp.ServicesProcessed != 10 {
		t.Errorf("ServicesProcessed = %d, want 10", resp.ServicesProcessed)
	}

	if resp.CPEsGenerated != 15 {
		t.Errorf("CPEsGenerated = %d, want 15", resp.CPEsGenerated)
	}

	if resp.VulnsFound != 5 {
		t.Errorf("VulnsFound = %d, want 5", resp.VulnsFound)
	}

	if resp.RelationshipsCreated != 12 {
		t.Errorf("RelationshipsCreated = %d, want 12", resp.RelationshipsCreated)
	}
}

// Mock test for CPE generation workflow step
func TestWorkflow_CPEGeneration(t *testing.T) {
	services := []enrichment.ServiceInfo{
		{
			ID:      "service:1",
			Product: "nginx",
			Version: "1.24.0",
		},
		{
			ID:      "service:2",
			Product: "apache",
			Version: "2.4.57",
		},
		{
			ID:      "service:3",
			Product: "",
			Version: "",
			Banner:  "SSH-2.0-OpenSSH_9.0",
		},
	}

	// This simulates what happens in the workflow's first step
	serviceCPEs := enrichment.GenerateCPEBatch(services)

	// Verify results
	if len(serviceCPEs) < 2 {
		t.Errorf("Expected at least 2 services with CPEs, got %d", len(serviceCPEs))
	}

	// Check service 1
	if cpes, exists := serviceCPEs["service:1"]; exists {
		if len(cpes) == 0 {
			t.Error("service:1 should have CPEs")
		}
		if cpes[0].Product != "nginx" {
			t.Errorf("service:1 product = %v, want nginx", cpes[0].Product)
		}
	} else {
		t.Error("service:1 should be in results")
	}

	// Check service 3 (banner parsing)
	if cpes, exists := serviceCPEs["service:3"]; exists {
		if len(cpes) == 0 {
			t.Error("service:3 should have CPEs from banner")
		}
		if cpes[0].Product != "openssh" {
			t.Errorf("service:3 product = %v, want openssh", cpes[0].Product)
		}
	} else {
		t.Error("service:3 should be in results (parsed from banner)")
	}
}

// Mock test for CVE matching workflow step
func TestWorkflow_CVEMatching(t *testing.T) {
	serviceCPEs := map[string][]enrichment.CPEIdentifier{
		"service:1": {
			{
				Vendor:  "nginx",
				Product: "nginx",
				Version: "1.24.0",
				CPE:     "cpe:2.3:a:nginx:nginx:1.24.0:*:*:*:*:*:*:*",
			},
		},
	}

	cvesByCPE := map[string][]enrichment.CVEItem{
		"cpe:2.3:a:nginx:nginx:1.24.0:*:*:*:*:*:*:*": {
			{
				CVEID:    "CVE-2023-1001",
				CVSS:     7.5,
				Severity: "HIGH",
			},
		},
	}

	// This simulates the matching step in the workflow
	matches := enrichment.MatchServicesToCVEs(serviceCPEs, cvesByCPE)

	if len(matches) != 1 {
		t.Fatalf("Expected 1 match, got %d", len(matches))
	}

	match := matches[0]
	if match.ServiceID != "service:1" {
		t.Errorf("ServiceID = %v, want service:1", match.ServiceID)
	}

	if match.CVE != "CVE-2023-1001" {
		t.Errorf("CVE = %v, want CVE-2023-1001", match.CVE)
	}

	if match.CVSS != 7.5 {
		t.Errorf("CVSS = %v, want 7.5", match.CVSS)
	}
}

// Test deduplication in the workflow
func TestWorkflow_Deduplication(t *testing.T) {
	matches := []enrichment.VulnMatch{
		{ServiceID: "s1", CVE: "CVE-1", CVSS: 9.8, Severity: "CRITICAL"},
		{ServiceID: "s1", CVE: "CVE-1", CVSS: 9.8, Severity: "CRITICAL"}, // Duplicate
		{ServiceID: "s2", CVE: "CVE-2", CVSS: 7.5, Severity: "HIGH"},
	}

	// This simulates the deduplication step
	deduplicated := enrichment.DeduplicateMatches(matches)

	if len(deduplicated) != 2 {
		t.Errorf("Expected 2 unique matches after deduplication, got %d", len(deduplicated))
	}
}

// Test batch processing logic
func TestWorkflow_BatchProcessing(t *testing.T) {
	// Create a large batch of services
	services := make([]enrichment.ServiceInfo, 100)
	for i := 0; i < 100; i++ {
		services[i] = enrichment.ServiceInfo{
			ID:      "service:" + string(rune(i)),
			Product: "nginx",
			Version: "1.24.0",
		}
	}

	// Generate CPEs (simulates workflow step)
	serviceCPEs := enrichment.GenerateCPEBatch(services)

	// Verify all services processed
	if len(serviceCPEs) != 100 {
		t.Errorf("Expected 100 services with CPEs, got %d", len(serviceCPEs))
	}

	// Collect unique CPEs
	uniqueCPEs := make(map[string]bool)
	for _, cpes := range serviceCPEs {
		for _, cpe := range cpes {
			uniqueCPEs[cpe.CPE] = true
		}
	}

	// Since all services are the same, we should have 1 unique CPE
	if len(uniqueCPEs) != 1 {
		t.Errorf("Expected 1 unique CPE for identical services, got %d", len(uniqueCPEs))
	}
}

// Integration test placeholder
// In a real integration test, this would:
// 1. Set up a test SurrealDB instance
// 2. Create test service records
// 3. Run the workflow with Restate test runtime
// 4. Verify database state after workflow completion
func TestEnrichCPEWorkflow_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// TODO: Implement full integration test
	// This requires:
	// - Test SurrealDB instance (could use testcontainers)
	// - Restate test runtime
	// - Mock NVD API responses (to avoid rate limits)

	t.Skip("Integration test not yet implemented")
}
