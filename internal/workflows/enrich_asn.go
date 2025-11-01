package workflows

import (
	"context"
	"fmt"
	"strings"
	"time"

	restate "github.com/restatedev/sdk-go"
	"github.com/spectra-red/recon/internal/enrichment"
	"github.com/surrealdb/surrealdb.go"
)

// EnrichASNWorkflow handles ASN enrichment for IP addresses
type EnrichASNWorkflow struct {
	db        *surrealdb.DB
	asnClient enrichment.ASNClient
}

// NewEnrichASNWorkflow creates a new EnrichASNWorkflow instance
func NewEnrichASNWorkflow(db *surrealdb.DB, asnClient enrichment.ASNClient) *EnrichASNWorkflow {
	return &EnrichASNWorkflow{
		db:        db,
		asnClient: asnClient,
	}
}

// ServiceName returns the Restate service name
func (w *EnrichASNWorkflow) ServiceName() string {
	return "EnrichASNWorkflow"
}

// EnrichASNRequest represents the request to enrich ASN data
type EnrichASNRequest struct {
	IPs       []string `json:"ips"`        // IP addresses to enrich (batch)
	JobID     string   `json:"job_id"`     // Optional job ID for tracking
	ForceRefresh bool  `json:"force_refresh"` // Force re-lookup even if cached
}

// EnrichASNResponse represents the response from ASN enrichment
type EnrichASNResponse struct {
	TotalIPs      int                       `json:"total_ips"`
	EnrichedIPs   int                       `json:"enriched_ips"`
	CachedIPs     int                       `json:"cached_ips"`
	FailedIPs     int                       `json:"failed_ips"`
	FailedIPsList []string                  `json:"failed_ips_list,omitempty"`
	ASNData       map[string]*enrichment.ASNInfo `json:"asn_data"`
}

// HostASNData represents the ASN data to update in the database
type HostASNData struct {
	IP      string
	ASN     int
	Country string
}

// Run executes the ASN enrichment workflow with durable steps
func (w *EnrichASNWorkflow) Run(ctx restate.Context, req EnrichASNRequest) (EnrichASNResponse, error) {
	// Validate request
	if len(req.IPs) == 0 {
		return EnrichASNResponse{}, fmt.Errorf("no IPs provided")
	}

	// Limit batch size to prevent overwhelming the workflow
	maxBatchSize := 100
	if len(req.IPs) > maxBatchSize {
		return EnrichASNResponse{}, fmt.Errorf("batch size exceeds maximum of %d (got %d)", maxBatchSize, len(req.IPs))
	}

	response := EnrichASNResponse{
		TotalIPs:      len(req.IPs),
		ASNData:       make(map[string]*enrichment.ASNInfo),
		FailedIPsList: make([]string, 0),
	}

	// Step 1: Check which IPs need enrichment (filter already enriched hosts)
	ipsToEnrich, err := restate.Run[[]string](ctx, func(ctx restate.RunContext) ([]string, error) {
		if req.ForceRefresh {
			// Force refresh all IPs
			return req.IPs, nil
		}
		return w.filterIPsNeedingEnrichment(req.IPs)
	})
	if err != nil {
		return response, fmt.Errorf("failed to filter IPs: %w", err)
	}

	// If no IPs need enrichment, return early
	if len(ipsToEnrich) == 0 {
		response.CachedIPs = len(req.IPs)
		return response, nil
	}

	// Step 2: Lookup ASN data (external API call - durable)
	asnLookupResults, err := restate.Run[map[string]*enrichment.ASNInfo](ctx, func(ctx restate.RunContext) (map[string]*enrichment.ASNInfo, error) {
		// Use background context for external API call (not the Restate context)
		apiCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		return w.asnClient.LookupBatch(apiCtx, ipsToEnrich)
	})
	if err != nil {
		return response, fmt.Errorf("failed to lookup ASN data: %w", err)
	}

	// Track results
	response.ASNData = asnLookupResults
	response.EnrichedIPs = len(asnLookupResults)
	response.CachedIPs = response.TotalIPs - len(ipsToEnrich)
	response.FailedIPs = len(ipsToEnrich) - len(asnLookupResults)

	// Identify failed IPs
	for _, ip := range ipsToEnrich {
		if _, ok := asnLookupResults[ip]; !ok {
			response.FailedIPsList = append(response.FailedIPsList, ip)
		}
	}

	// Step 3: Update SurrealDB host records with ASN data
	_, err = restate.Run[int](ctx, func(ctx restate.RunContext) (int, error) {
		return w.updateHostASNData(asnLookupResults)
	})
	if err != nil {
		return response, fmt.Errorf("failed to update host ASN data: %w", err)
	}

	// Step 4: Create or update ASN nodes and edges
	_, err = restate.Run[int](ctx, func(ctx restate.RunContext) (int, error) {
		return w.upsertASNNodesAndEdges(asnLookupResults)
	})
	if err != nil {
		return response, fmt.Errorf("failed to upsert ASN nodes: %w", err)
	}

	return response, nil
}

// filterIPsNeedingEnrichment queries the database to find IPs that don't have ASN data
func (w *EnrichASNWorkflow) filterIPsNeedingEnrichment(ips []string) ([]string, error) {
	ctx := context.Background()
	var ipsToEnrich []string

	// Query each IP to check if it has ASN data
	for _, ip := range ips {
		query := `SELECT asn FROM type::thing('host', $host_id) LIMIT 1;`
		result, err := surrealdb.Query[[]map[string]interface{}](ctx, w.db, query, map[string]interface{}{
			"host_id": strings.ReplaceAll(ip, ".", "_"),
		})

		// If query fails or host doesn't exist, add to enrich list
		if err != nil || result == nil || len(*result) == 0 {
			ipsToEnrich = append(ipsToEnrich, ip)
			continue
		}

		// Check if ASN field is set
		hosts := (*result)[0].Result
		if len(hosts) == 0 {
			ipsToEnrich = append(ipsToEnrich, ip)
			continue
		}

		// Check if asn field exists and is not nil/zero
		host := hosts[0]
		asnValue, hasASN := host["asn"]
		if !hasASN || asnValue == nil || asnValue == 0 {
			ipsToEnrich = append(ipsToEnrich, ip)
		}
	}

	return ipsToEnrich, nil
}

// updateHostASNData updates host records in SurrealDB with ASN information
func (w *EnrichASNWorkflow) updateHostASNData(asnData map[string]*enrichment.ASNInfo) (int, error) {
	ctx := context.Background()
	updated := 0

	for ip, info := range asnData {
		hostID := strings.ReplaceAll(ip, ".", "_")

		// Update host with ASN data
		updateQuery := `
			UPDATE type::thing('host', $host_id) MERGE {
				asn: $asn,
				country: $country
			};
		`

		_, err := surrealdb.Query[interface{}](ctx, w.db, updateQuery, map[string]interface{}{
			"host_id": hostID,
			"asn":     info.Number,
			"country": info.Country,
		})

		if err != nil {
			// Log error but continue with other hosts
			continue
		}

		updated++
	}

	return updated, nil
}

// upsertASNNodesAndEdges creates ASN nodes and IN_ASN edges in the graph
func (w *EnrichASNWorkflow) upsertASNNodesAndEdges(asnData map[string]*enrichment.ASNInfo) (int, error) {
	ctx := context.Background()
	created := 0

	// Group by ASN to avoid duplicate upserts
	asnMap := make(map[int]*enrichment.ASNInfo)
	hostsByASN := make(map[int][]string)

	for ip, info := range asnData {
		asnMap[info.Number] = info
		hostsByASN[info.Number] = append(hostsByASN[info.Number], ip)
	}

	// Upsert ASN nodes
	for asnNum, info := range asnMap {
		upsertASNQuery := `
			LET $asn_id = type::thing('asn', $asn_number);
			CREATE $asn_id CONTENT {
				number: $asn_number,
				org: $org,
				country: $country
			} ON DUPLICATE KEY UPDATE {
				org: $org,
				country: $country
			};
		`

		_, err := surrealdb.Query[interface{}](ctx, w.db, upsertASNQuery, map[string]interface{}{
			"asn_number": asnNum,
			"org":        info.Org,
			"country":    info.Country,
		})

		if err != nil {
			continue
		}

		// Create IN_ASN edges for all hosts in this ASN
		for _, ip := range hostsByASN[asnNum] {
			hostID := strings.ReplaceAll(ip, ".", "_")

			relateQuery := `
				LET $host_id = type::thing('host', $host_encoded);
				LET $asn_id = type::thing('asn', $asn_number);
				RELATE $host_id->IN_ASN->$asn_id;
			`

			_, err := surrealdb.Query[interface{}](ctx, w.db, relateQuery, map[string]interface{}{
				"host_encoded": hostID,
				"asn_number":   asnNum,
			})

			if err != nil {
				continue
			}

			created++
		}
	}

	return created, nil
}
