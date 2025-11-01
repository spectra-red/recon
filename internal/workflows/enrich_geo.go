package workflows

import (
	"context"
	"fmt"
	"strings"
	"time"

	restate "github.com/restatedev/sdk-go"
	"github.com/spectra-red/recon/internal/enrichment"
	"github.com/surrealdb/surrealdb.go"
	"go.uber.org/zap"
)

// EnrichGeoWorkflow handles GeoIP enrichment for IP addresses
type EnrichGeoWorkflow struct {
	db        *surrealdb.DB
	geoClient *enrichment.GeoIPClient
	logger    *zap.Logger
}

// NewEnrichGeoWorkflow creates a new GeoIP enrichment workflow
func NewEnrichGeoWorkflow(db *surrealdb.DB, geoClient *enrichment.GeoIPClient, logger *zap.Logger) *EnrichGeoWorkflow {
	if logger == nil {
		logger, _ = zap.NewProduction()
	}

	return &EnrichGeoWorkflow{
		db:        db,
		geoClient: geoClient,
		logger:    logger,
	}
}

// ServiceName returns the Restate service name
func (w *EnrichGeoWorkflow) ServiceName() string {
	return "EnrichGeoWorkflow"
}

// EnrichGeoRequest represents the request to enrich IPs with geographic data
type EnrichGeoRequest struct {
	IPs []string `json:"ips"` // Batch of IP addresses to enrich
}

// EnrichGeoResponse represents the response from the enrichment workflow
type EnrichGeoResponse struct {
	Enriched int      `json:"enriched"` // Number of IPs successfully enriched
	Failed   int      `json:"failed"`   // Number of IPs that failed enrichment
	Errors   []string `json:"errors,omitempty"`
}

// GeoNodeResult holds the result of creating geographic nodes
type GeoNodeResult struct {
	CountriesCreated int
	RegionsCreated   int
	CitiesCreated    int
}

// RelationshipResult holds the result of creating geographic relationships
type RelationshipResult struct {
	HostCityLinks   int
	CityRegionLinks int
	RegionCountryLinks int
}

// Run executes the GeoIP enrichment workflow with durable steps
func (w *EnrichGeoWorkflow) Run(ctx restate.Context, req EnrichGeoRequest) (EnrichGeoResponse, error) {
	if len(req.IPs) == 0 {
		return EnrichGeoResponse{}, fmt.Errorf("no IPs provided for enrichment")
	}

	w.logger.Info("starting GeoIP enrichment workflow",
		zap.Int("ip_count", len(req.IPs)))

	// Step 1: Lookup GeoIP data for all IPs
	geoData, err := restate.Run(ctx, func(ctx restate.RunContext) (map[string]*enrichment.GeoIPInfo, error) {
		return w.lookupGeoIP(req.IPs)
	})
	if err != nil {
		w.logger.Error("GeoIP lookup failed",
			zap.Error(err),
			zap.Int("ip_count", len(req.IPs)))
		return EnrichGeoResponse{
			Failed: len(req.IPs),
			Errors: []string{fmt.Sprintf("GeoIP lookup failed: %v", err)},
		}, err
	}

	w.logger.Info("GeoIP lookup completed",
		zap.Int("successful", len(geoData)),
		zap.Int("failed", len(req.IPs)-len(geoData)))

	// Step 2: Create geographic nodes (city, region, country)
	_, err = restate.Run(ctx, func(ctx restate.RunContext) (GeoNodeResult, error) {
		return w.createGeoNodes(geoData)
	})
	if err != nil {
		w.logger.Error("failed to create geographic nodes", zap.Error(err))
		return EnrichGeoResponse{
			Failed: len(req.IPs),
			Errors: []string{fmt.Sprintf("Failed to create geographic nodes: %v", err)},
		}, err
	}

	// Step 3: Create geographic relationships
	_, err = restate.Run(ctx, func(ctx restate.RunContext) (RelationshipResult, error) {
		return w.createGeoRelationships(geoData)
	})
	if err != nil {
		w.logger.Error("failed to create geographic relationships", zap.Error(err))
		return EnrichGeoResponse{
			Failed: len(req.IPs),
			Errors: []string{fmt.Sprintf("Failed to create geographic relationships: %v", err)},
		}, err
	}

	// Step 4: Update host records with geographic data
	_, err = restate.Run(ctx, func(ctx restate.RunContext) (restate.Void, error) {
		return restate.Void{}, w.updateHostRecords(geoData)
	})
	if err != nil {
		w.logger.Error("failed to update host records", zap.Error(err))
		return EnrichGeoResponse{
			Enriched: len(geoData),
			Failed:   len(req.IPs) - len(geoData),
			Errors:   []string{fmt.Sprintf("Failed to update host records: %v", err)},
		}, err
	}

	w.logger.Info("GeoIP enrichment workflow completed",
		zap.Int("enriched", len(geoData)),
		zap.Int("failed", len(req.IPs)-len(geoData)))

	return EnrichGeoResponse{
		Enriched: len(geoData),
		Failed:   len(req.IPs) - len(geoData),
	}, nil
}

// lookupGeoIP performs batch GeoIP lookup using the GeoIP client
func (w *EnrichGeoWorkflow) lookupGeoIP(ips []string) (map[string]*enrichment.GeoIPInfo, error) {
	if w.geoClient == nil {
		return nil, fmt.Errorf("GeoIP client not initialized")
	}

	w.logger.Info("performing GeoIP lookup", zap.Int("ip_count", len(ips)))

	results, err := w.geoClient.LookupBatch(ips)
	if err != nil {
		return nil, fmt.Errorf("batch GeoIP lookup failed: %w", err)
	}

	w.logger.Info("GeoIP lookup completed",
		zap.Int("successful", len(results)),
		zap.Int("total", len(ips)))

	return results, nil
}

// createGeoNodes creates city, region, and country nodes in SurrealDB
// Uses idempotent upserts with ON DUPLICATE KEY
func (w *EnrichGeoWorkflow) createGeoNodes(geoData map[string]*enrichment.GeoIPInfo) (GeoNodeResult, error) {
	ctx := context.Background()
	result := GeoNodeResult{}

	// Track unique geographic entities
	countries := make(map[string]*enrichment.GeoIPInfo)
	regions := make(map[string]*enrichment.GeoIPInfo)
	cities := make(map[string]*enrichment.GeoIPInfo)

	// Collect unique entities
	for _, info := range geoData {
		if info.CountryCC != "" {
			countries[info.CountryCC] = info
		}
		if info.Region != "" {
			regionKey := fmt.Sprintf("%s:%s", info.CountryCC, info.Region)
			regions[regionKey] = info
		}
		if info.City != "" {
			cityKey := fmt.Sprintf("%s:%s:%s", info.CountryCC, info.Region, info.City)
			cities[cityKey] = info
		}
	}

	w.logger.Info("creating geographic nodes",
		zap.Int("countries", len(countries)),
		zap.Int("regions", len(regions)),
		zap.Int("cities", len(cities)))

	// Create country nodes
	for cc, info := range countries {
		query := `
			LET $country_id = type::thing('country', $cc);
			CREATE $country_id CONTENT {
				cc: $cc,
				name: $name
			} ON DUPLICATE KEY UPDATE {
				name: $name
			};
		`
		_, err := surrealdb.Query[interface{}](ctx, w.db, query, map[string]interface{}{
			"cc":   cc,
			"name": info.Country,
		})
		if err != nil {
			w.logger.Error("failed to create country node",
				zap.String("country", cc),
				zap.Error(err))
			continue
		}
		result.CountriesCreated++
	}

	// Create region nodes
	for regionKey, info := range regions {
		// Generate a safe region ID
		regionID := strings.ReplaceAll(regionKey, ":", "_")

		query := `
			LET $region_id = type::thing('region', $region_id);
			CREATE $region_id CONTENT {
				name: $name,
				cc: $cc,
				code: $code
			} ON DUPLICATE KEY UPDATE {
				name: $name
			};
		`
		_, err := surrealdb.Query[interface{}](ctx, w.db, query, map[string]interface{}{
			"region_id": regionID,
			"name":      info.Region,
			"cc":        info.CountryCC,
			"code":      "", // Region code not available from MaxMind
		})
		if err != nil {
			w.logger.Error("failed to create region node",
				zap.String("region", regionKey),
				zap.Error(err))
			continue
		}
		result.RegionsCreated++
	}

	// Create city nodes
	for cityKey, info := range cities {
		// Generate a safe city ID
		cityID := strings.ReplaceAll(cityKey, ":", "_")

		query := `
			LET $city_id = type::thing('city', $city_id);
			CREATE $city_id CONTENT {
				name: $name,
				cc: $cc,
				lat: $lat,
				lon: $lon
			} ON DUPLICATE KEY UPDATE {
				name: $name,
				lat: $lat,
				lon: $lon
			};
		`
		_, err := surrealdb.Query[interface{}](ctx, w.db, query, map[string]interface{}{
			"city_id": cityID,
			"name":    info.City,
			"cc":      info.CountryCC,
			"lat":     info.Latitude,
			"lon":     info.Longitude,
		})
		if err != nil {
			w.logger.Error("failed to create city node",
				zap.String("city", cityKey),
				zap.Error(err))
			continue
		}
		result.CitiesCreated++
	}

	w.logger.Info("geographic nodes created",
		zap.Int("countries", result.CountriesCreated),
		zap.Int("regions", result.RegionsCreated),
		zap.Int("cities", result.CitiesCreated))

	return result, nil
}

// createGeoRelationships creates LOCATED_IN relationships between geographic entities
// host -> IN_CITY -> city -> IN_REGION -> region -> IN_COUNTRY -> country
func (w *EnrichGeoWorkflow) createGeoRelationships(geoData map[string]*enrichment.GeoIPInfo) (RelationshipResult, error) {
	ctx := context.Background()
	result := RelationshipResult{}

	for ip, info := range geoData {
		// Create host -> IN_CITY -> city relationship
		if info.City != "" {
			cityID := strings.ReplaceAll(fmt.Sprintf("%s:%s:%s", info.CountryCC, info.Region, info.City), ":", "_")
			hostID := strings.ReplaceAll(ip, ".", "_")

			query := `
				LET $host_id = type::thing('host', $host_id);
				LET $city_id = type::thing('city', $city_id);
				RELATE $host_id->IN_CITY->$city_id;
			`
			_, err := surrealdb.Query[interface{}](ctx, w.db, query, map[string]interface{}{
				"host_id": hostID,
				"city_id": cityID,
			})
			if err != nil {
				w.logger.Error("failed to create host->city relationship",
					zap.String("ip", ip),
					zap.String("city", info.City),
					zap.Error(err))
			} else {
				result.HostCityLinks++
			}
		}

		// Create city -> IN_REGION -> region relationship
		if info.City != "" && info.Region != "" {
			cityID := strings.ReplaceAll(fmt.Sprintf("%s:%s:%s", info.CountryCC, info.Region, info.City), ":", "_")
			regionID := strings.ReplaceAll(fmt.Sprintf("%s:%s", info.CountryCC, info.Region), ":", "_")

			query := `
				LET $city_id = type::thing('city', $city_id);
				LET $region_id = type::thing('region', $region_id);
				RELATE $city_id->IN_REGION->$region_id;
			`
			_, err := surrealdb.Query[interface{}](ctx, w.db, query, map[string]interface{}{
				"city_id":   cityID,
				"region_id": regionID,
			})
			if err != nil {
				w.logger.Error("failed to create city->region relationship",
					zap.String("city", info.City),
					zap.String("region", info.Region),
					zap.Error(err))
			} else {
				result.CityRegionLinks++
			}
		}

		// Create region -> IN_COUNTRY -> country relationship
		if info.Region != "" && info.CountryCC != "" {
			regionID := strings.ReplaceAll(fmt.Sprintf("%s:%s", info.CountryCC, info.Region), ":", "_")

			query := `
				LET $region_id = type::thing('region', $region_id);
				LET $country_id = type::thing('country', $cc);
				RELATE $region_id->IN_COUNTRY->$country_id;
			`
			_, err := surrealdb.Query[interface{}](ctx, w.db, query, map[string]interface{}{
				"region_id": regionID,
				"cc":        info.CountryCC,
			})
			if err != nil {
				w.logger.Error("failed to create region->country relationship",
					zap.String("region", info.Region),
					zap.String("country", info.CountryCC),
					zap.Error(err))
			} else {
				result.RegionCountryLinks++
			}
		}
	}

	w.logger.Info("geographic relationships created",
		zap.Int("host_city", result.HostCityLinks),
		zap.Int("city_region", result.CityRegionLinks),
		zap.Int("region_country", result.RegionCountryLinks))

	return result, nil
}

// updateHostRecords updates host records with city, region, and country fields
func (w *EnrichGeoWorkflow) updateHostRecords(geoData map[string]*enrichment.GeoIPInfo) error {
	ctx := context.Background()
	now := time.Now().UTC()

	for ip, info := range geoData {
		hostID := strings.ReplaceAll(ip, ".", "_")

		query := `
			UPDATE type::thing('host', $host_id) MERGE {
				city: $city,
				region: $region,
				country: $country,
				last_seen: $now
			};
		`
		_, err := surrealdb.Query[interface{}](ctx, w.db, query, map[string]interface{}{
			"host_id": hostID,
			"city":    info.City,
			"region":  info.Region,
			"country": info.Country,
			"now":     now,
		})
		if err != nil {
			w.logger.Error("failed to update host record",
				zap.String("ip", ip),
				zap.Error(err))
			return fmt.Errorf("failed to update host %s: %w", ip, err)
		}
	}

	w.logger.Info("host records updated",
		zap.Int("count", len(geoData)))

	return nil
}
