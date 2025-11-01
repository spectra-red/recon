package workflows

import (
	"context"
	"fmt"
	"time"

	restate "github.com/restatedev/sdk-go"
	"github.com/spectra-red/recon/internal/enrichment"
	"github.com/surrealdb/surrealdb.go"
)

// EnrichCPEWorkflow handles CPE matching and vulnerability correlation
type EnrichCPEWorkflow struct {
	db        *surrealdb.DB
	nvdClient *enrichment.NVDClient
}

// NewEnrichCPEWorkflow creates a new EnrichCPEWorkflow instance
func NewEnrichCPEWorkflow(db *surrealdb.DB, nvdAPIKey string) *EnrichCPEWorkflow {
	return &EnrichCPEWorkflow{
		db:        db,
		nvdClient: enrichment.NewNVDClient(nvdAPIKey),
	}
}

// ServiceName returns the Restate service name
func (w *EnrichCPEWorkflow) ServiceName() string {
	return "EnrichCPEWorkflow"
}

// EnrichCPERequest represents the request to the CPE enrichment workflow
type EnrichCPERequest struct {
	Services []enrichment.ServiceInfo `json:"services"` // Services to enrich
	BatchID  string                   `json:"batch_id"` // Optional batch identifier for tracking
}

// EnrichCPEResponse represents the response from the CPE enrichment workflow
type EnrichCPEResponse struct {
	BatchID            string `json:"batch_id"`
	ServicesProcessed  int    `json:"services_processed"`
	CPEsGenerated      int    `json:"cpes_generated"`
	VulnsFound         int    `json:"vulns_found"`
	RelationshipsCreated int  `json:"relationships_created"`
}

// Run executes the CPE enrichment workflow with durable steps
// This workflow is idempotent and can be safely retried
func (w *EnrichCPEWorkflow) Run(ctx restate.Context, req EnrichCPERequest) (EnrichCPEResponse, error) {
	// Step 1: Generate CPE identifiers from service data
	serviceCPEs, err := restate.Run[map[string][]enrichment.CPEIdentifier](ctx, func(ctx restate.RunContext) (map[string][]enrichment.CPEIdentifier, error) {
		cpes := enrichment.GenerateCPEBatch(req.Services)
		return cpes, nil
	})
	if err != nil {
		return EnrichCPEResponse{}, fmt.Errorf("failed to generate CPEs: %w", err)
	}

	cpeCount := 0
	for _, cpes := range serviceCPEs {
		cpeCount += len(cpes)
	}

	// Step 2: Query NVD for vulnerabilities (with rate limiting)
	// We collect all unique CPE strings
	uniqueCPEs := make(map[string]bool)
	for _, cpes := range serviceCPEs {
		for _, cpe := range cpes {
			uniqueCPEs[cpe.CPE] = true
		}
	}

	cpeList := make([]string, 0, len(uniqueCPEs))
	for cpe := range uniqueCPEs {
		cpeList = append(cpeList, cpe)
	}

	cvesByCPE, err := restate.Run[map[string][]enrichment.CVEItem](ctx, func(ctx restate.RunContext) (map[string][]enrichment.CVEItem, error) {
		return w.nvdClient.QueryByCPEBatch(context.Background(), cpeList)
	})
	if err != nil {
		return EnrichCPEResponse{}, fmt.Errorf("failed to query NVD: %w", err)
	}

	// Step 3: Match services to CVEs
	matches, err := restate.Run[[]enrichment.VulnMatch](ctx, func(ctx restate.RunContext) ([]enrichment.VulnMatch, error) {
		allMatches := enrichment.MatchServicesToCVEs(serviceCPEs, cvesByCPE)
		// Deduplicate matches
		deduped := enrichment.DeduplicateMatches(allMatches)
		return deduped, nil
	})
	if err != nil {
		return EnrichCPEResponse{}, fmt.Errorf("failed to match CVEs: %w", err)
	}

	// Step 4: Create vulnerability nodes in SurrealDB
	vulnCount, err := restate.Run[int](ctx, func(ctx restate.RunContext) (int, error) {
		return w.createVulnNodes(cvesByCPE)
	})
	if err != nil {
		return EnrichCPEResponse{}, fmt.Errorf("failed to create vulnerability nodes: %w", err)
	}

	// Step 5: Update service records with CPE identifiers
	_, err = restate.Run[int](ctx, func(ctx restate.RunContext) (int, error) {
		return w.updateServiceCPEs(serviceCPEs)
	})
	if err != nil {
		return EnrichCPEResponse{}, fmt.Errorf("failed to update service CPEs: %w", err)
	}

	// Step 6: Create AFFECTED_BY relationships
	relationshipsCreated, err := restate.Run[int](ctx, func(ctx restate.RunContext) (int, error) {
		return w.createAffectedByRelationships(matches)
	})
	if err != nil {
		return EnrichCPEResponse{}, fmt.Errorf("failed to create relationships: %w", err)
	}

	return EnrichCPEResponse{
		BatchID:              req.BatchID,
		ServicesProcessed:    len(req.Services),
		CPEsGenerated:        cpeCount,
		VulnsFound:           vulnCount,
		RelationshipsCreated: relationshipsCreated,
	}, nil
}

// createVulnNodes creates vulnerability nodes in SurrealDB
// Returns the count of vulnerabilities created
func (w *EnrichCPEWorkflow) createVulnNodes(cvesByCPE map[string][]enrichment.CVEItem) (int, error) {
	ctx := context.Background()
	now := time.Now().UTC()
	count := 0

	// Collect unique CVEs (same CVE may appear in multiple CPE results)
	uniqueCVEs := make(map[string]enrichment.CVEItem)
	for _, cves := range cvesByCPE {
		for _, cve := range cves {
			uniqueCVEs[cve.CVEID] = cve
		}
	}

	for _, cve := range uniqueCVEs {
		// Create vuln node (idempotent upsert)
		query := `
			LET $vuln_id = type::thing('vuln', $cve_id);
			CREATE $vuln_id CONTENT {
				cve_id: $cve_id,
				cvss: $cvss,
				severity: $severity,
				kev_flag: false,
				first_seen: $now,
				last_updated: $now
			} ON DUPLICATE KEY UPDATE {
				cvss: $cvss,
				severity: $severity,
				last_updated: $now
			};
		`

		_, err := surrealdb.Query[interface{}](ctx, w.db, query, map[string]interface{}{
			"cve_id":   cve.CVEID,
			"cvss":     cve.CVSS,
			"severity": cve.Severity,
			"now":      now,
		})

		if err != nil {
			return count, fmt.Errorf("failed to create vuln node %s: %w", cve.CVEID, err)
		}

		// Create vuln_doc node for RAG (if description exists)
		if cve.Description != "" {
			docQuery := `
				LET $doc_id = type::thing('vuln_doc', $cve_id);
				CREATE $doc_id CONTENT {
					cve_id: $cve_id,
					title: $title,
					summary: $summary,
					cvss: $cvss,
					epss: 0.0,
					cpe: $cpe,
					exploit_refs: $refs,
					embedding: [],
					published_date: $published,
					last_modified: $modified
				} ON DUPLICATE KEY UPDATE {
					summary: $summary,
					cvss: $cvss,
					cpe: $cpe,
					exploit_refs: $refs,
					last_modified: $modified
				};
			`

			// Use CVE ID as title if not available
			title := cve.CVEID

			_, err := surrealdb.Query[interface{}](ctx, w.db, docQuery, map[string]interface{}{
				"cve_id":    cve.CVEID,
				"title":     title,
				"summary":   cve.Description,
				"cvss":      cve.CVSS,
				"cpe":       cve.CPEs,
				"refs":      cve.References,
				"published": cve.Published,
				"modified":  cve.Modified,
			})

			if err != nil {
				// Log error but don't fail the entire process
				// vuln_doc is for RAG, not critical for basic vulnerability tracking
				continue
			}
		}

		count++
	}

	return count, nil
}

// updateServiceCPEs updates service records with generated CPE identifiers
func (w *EnrichCPEWorkflow) updateServiceCPEs(serviceCPEs map[string][]enrichment.CPEIdentifier) (int, error) {
	ctx := context.Background()
	count := 0

	for serviceID, cpes := range serviceCPEs {
		// Extract CPE strings
		cpeStrings := make([]string, len(cpes))
		for i, cpe := range cpes {
			cpeStrings[i] = cpe.CPE
		}

		// Update service record with CPE array
		query := `
			UPDATE $service_id MERGE {
				cpe: $cpe,
				last_seen: $now
			};
		`

		_, err := surrealdb.Query[interface{}](ctx, w.db, query, map[string]interface{}{
			"service_id": serviceID,
			"cpe":        cpeStrings,
			"now":        time.Now().UTC(),
		})

		if err != nil {
			return count, fmt.Errorf("failed to update service %s: %w", serviceID, err)
		}

		count++
	}

	return count, nil
}

// createAffectedByRelationships creates AFFECTED_BY edges between services and vulnerabilities
func (w *EnrichCPEWorkflow) createAffectedByRelationships(matches []enrichment.VulnMatch) (int, error) {
	ctx := context.Background()
	now := time.Now().UTC()
	count := 0

	for _, match := range matches {
		// Create AFFECTED_BY relationship (idempotent)
		query := `
			LET $service_id = $sid;
			LET $vuln_id = type::thing('vuln', $cve_id);
			RELATE $service_id->AFFECTED_BY->$vuln_id CONTENT {
				confidence: 1.0,
				first_detected: $now,
				last_confirmed: $now
			} ON DUPLICATE KEY UPDATE {
				last_confirmed: $now
			};
		`

		_, err := surrealdb.Query[interface{}](ctx, w.db, query, map[string]interface{}{
			"sid":    match.ServiceID,
			"cve_id": match.CVE,
			"now":    now,
		})

		if err != nil {
			return count, fmt.Errorf("failed to create AFFECTED_BY edge %s->%s: %w", match.ServiceID, match.CVE, err)
		}

		count++
	}

	return count, nil
}

// GetServicesByFilter retrieves services from SurrealDB for batch processing
func GetServicesByFilter(db *surrealdb.DB, filter ServiceFilter) ([]enrichment.ServiceInfo, error) {
	ctx := context.Background()

	// Build query based on filter
	query := `
		SELECT
			id,
			name,
			product,
			version,
			meta::tb(id) as table_name
		FROM service
		WHERE 1=1
	`

	params := make(map[string]interface{})

	// Add filters
	if filter.MinLastSeen != nil {
		query += " AND last_seen >= $min_last_seen"
		params["min_last_seen"] = filter.MinLastSeen
	}

	if filter.OnlyMissingCPE {
		query += " AND (cpe IS NONE OR array::len(cpe) = 0)"
	}

	// Add limit
	if filter.Limit > 0 {
		query += " LIMIT $limit"
		params["limit"] = filter.Limit
	} else {
		query += " LIMIT 100" // Default limit
	}

	// Execute query
	type ServiceRow struct {
		ID      string `json:"id"`
		Name    string `json:"name"`
		Product string `json:"product"`
		Version string `json:"version"`
	}

	result, err := surrealdb.Query[[]ServiceRow](ctx, db, query, params)
	if err != nil {
		return nil, fmt.Errorf("failed to query services: %w", err)
	}

	// Extract services from result
	var services []enrichment.ServiceInfo
	if result != nil && len(*result) > 0 {
		rows := (*result)[0].Result
		for _, row := range rows {
			services = append(services, enrichment.ServiceInfo{
				ID:      row.ID,
				Name:    row.Name,
				Product: row.Product,
				Version: row.Version,
				Banner:  "", // We'd need to fetch banners separately if needed
			})
		}
	}

	return services, nil
}

// ServiceFilter defines filters for retrieving services
type ServiceFilter struct {
	MinLastSeen    *time.Time // Only services seen after this time
	OnlyMissingCPE bool       // Only services without CPE identifiers
	Limit          int        // Maximum number of services to retrieve
}
