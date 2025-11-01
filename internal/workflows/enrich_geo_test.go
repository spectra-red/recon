package workflows

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/spectra-red/recon/internal/enrichment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/surrealdb/surrealdb.go"
	"go.uber.org/zap"
)

// TestEnrichGeoWorkflow_NewWorkflow tests workflow creation
func TestEnrichGeoWorkflow_NewWorkflow(t *testing.T) {
	db := &surrealdb.DB{} // Mock DB
	geoClient := &enrichment.GeoIPClient{}
	logger, _ := zap.NewDevelopment()

	workflow := NewEnrichGeoWorkflow(db, geoClient, logger)
	assert.NotNil(t, workflow)
	assert.Equal(t, "EnrichGeoWorkflow", workflow.ServiceName())
}

// TestEnrichGeoWorkflow_Integration tests the workflow with a real database
func TestEnrichGeoWorkflow_Integration(t *testing.T) {
	// Skip integration test if SurrealDB not available
	if os.Getenv("SKIP_INTEGRATION") != "" {
		t.Skip("Skipping integration test")
	}

	// Skip if GeoIP MMDB file not available
	mmdbPath := getTestMMDBPath()
	if mmdbPath == "" {
		t.Skip("No GeoIP MMDB file available for testing (set GEOIP_MMDB_PATH environment variable)")
	}

	// Connect to test SurrealDB instance
	db, err := setupTestDB(t)
	if err != nil {
		t.Skipf("SurrealDB not available: %v", err)
	}
	defer db.Close(context.Background())

	// Create GeoIP client
	geoClient, err := enrichment.NewGeoIPClient(enrichment.GeoIPConfig{
		MMDBPath: mmdbPath,
	})
	require.NoError(t, err)
	defer geoClient.Close()

	// Create logger
	logger, _ := zap.NewDevelopment()

	// Create workflow
	workflow := NewEnrichGeoWorkflow(db, geoClient, logger)

	t.Run("enrich valid IPs", func(t *testing.T) {
		// First, ensure hosts exist in the database
		ctx := context.Background()
		testIPs := []string{"8.8.8.8", "1.1.1.1"}

		// Create host records
		for _, ip := range testIPs {
			createHostQuery := `
				CREATE type::thing('host', $host_id) CONTENT {
					ip: $ip,
					last_seen: $now
				} ON DUPLICATE KEY UPDATE {
					last_seen: $now
				};
			`
			_, err := surrealdb.Query[interface{}](ctx, db, createHostQuery, map[string]interface{}{
				"host_id": ip,
				"ip":      ip,
				"now":     time.Now().UTC(),
			})
			require.NoError(t, err)
		}

		// Test GeoIP lookup directly
		geoData, err := workflow.lookupGeoIP(testIPs)
		require.NoError(t, err)
		assert.NotEmpty(t, geoData)

		// Test creating geographic nodes
		nodeResult, err := workflow.createGeoNodes(geoData)
		require.NoError(t, err)
		assert.Greater(t, nodeResult.CountriesCreated, 0)

		// Test creating relationships
		relResult, err := workflow.createGeoRelationships(geoData)
		require.NoError(t, err)
		assert.Greater(t, relResult.HostCityLinks, 0)

		// Test updating host records
		err = workflow.updateHostRecords(geoData)
		require.NoError(t, err)

		// Verify host records were updated
		for _, ip := range testIPs {
			verifyQuery := `SELECT * FROM type::thing('host', $host_id);`
			result, err := surrealdb.Query[[]interface{}](ctx, db, verifyQuery, map[string]interface{}{
				"host_id": ip,
			})
			require.NoError(t, err)
			assert.NotNil(t, result)
		}
	})
}

// TestEnrichGeoWorkflow_LookupGeoIP tests the GeoIP lookup step
func TestEnrichGeoWorkflow_LookupGeoIP(t *testing.T) {
	mmdbPath := getTestMMDBPath()
	if mmdbPath == "" {
		t.Skip("No GeoIP MMDB file available for testing")
	}

	geoClient, err := enrichment.NewGeoIPClient(enrichment.GeoIPConfig{
		MMDBPath: mmdbPath,
	})
	require.NoError(t, err)
	defer geoClient.Close()

	db := &surrealdb.DB{} // Mock DB
	logger, _ := zap.NewDevelopment()
	workflow := NewEnrichGeoWorkflow(db, geoClient, logger)

	t.Run("valid IPs", func(t *testing.T) {
		ips := []string{"8.8.8.8", "1.1.1.1"}
		results, err := workflow.lookupGeoIP(ips)
		require.NoError(t, err)
		assert.NotEmpty(t, results)
		assert.Contains(t, results, "8.8.8.8")
		assert.Contains(t, results, "1.1.1.1")
	})

	t.Run("empty IP list", func(t *testing.T) {
		results, err := workflow.lookupGeoIP([]string{})
		assert.Error(t, err)
		assert.Empty(t, results)
	})

	t.Run("invalid IPs", func(t *testing.T) {
		ips := []string{"invalid", "not-an-ip"}
		results, err := workflow.lookupGeoIP(ips)
		assert.Error(t, err) // Should fail when no successful lookups
		assert.Empty(t, results)
	})
}

// TestEnrichGeoWorkflow_CreateGeoNodes tests geographic node creation
func TestEnrichGeoWorkflow_CreateGeoNodes(t *testing.T) {
	if os.Getenv("SKIP_INTEGRATION") != "" {
		t.Skip("Skipping integration test")
	}

	db, err := setupTestDB(t)
	if err != nil {
		t.Skipf("SurrealDB not available: %v", err)
	}
	defer db.Close(context.Background())

	logger, _ := zap.NewDevelopment()
	workflow := NewEnrichGeoWorkflow(db, nil, logger)

	geoData := map[string]*enrichment.GeoIPInfo{
		"8.8.8.8": {
			IP:        "8.8.8.8",
			City:      "Mountain View",
			Region:    "California",
			Country:   "United States",
			CountryCC: "US",
			Latitude:  37.4056,
			Longitude: -122.0775,
		},
		"1.1.1.1": {
			IP:        "1.1.1.1",
			City:      "Los Angeles",
			Region:    "California",
			Country:   "United States",
			CountryCC: "US",
			Latitude:  34.0522,
			Longitude: -118.2437,
		},
	}

	result, err := workflow.createGeoNodes(geoData)
	require.NoError(t, err)

	// Should create 1 country (US), 1 region (California), 2 cities
	assert.Equal(t, 1, result.CountriesCreated)
	assert.Equal(t, 1, result.RegionsCreated)
	assert.Equal(t, 2, result.CitiesCreated)
}

// TestEnrichGeoWorkflow_CreateGeoRelationships tests relationship creation
func TestEnrichGeoWorkflow_CreateGeoRelationships(t *testing.T) {
	if os.Getenv("SKIP_INTEGRATION") != "" {
		t.Skip("Skipping integration test")
	}

	db, err := setupTestDB(t)
	if err != nil {
		t.Skipf("SurrealDB not available: %v", err)
	}
	defer db.Close(context.Background())

	logger, _ := zap.NewDevelopment()
	workflow := NewEnrichGeoWorkflow(db, nil, logger)

	// First create the nodes
	geoData := map[string]*enrichment.GeoIPInfo{
		"8.8.8.8": {
			IP:        "8.8.8.8",
			City:      "Mountain View",
			Region:    "California",
			Country:   "United States",
			CountryCC: "US",
			Latitude:  37.4056,
			Longitude: -122.0775,
		},
	}

	// Create host node
	ctx := context.Background()
	createHostQuery := `
		CREATE type::thing('host', $host_id) CONTENT {
			ip: $ip,
			last_seen: $now
		};
	`
	_, err = surrealdb.Query[interface{}](ctx, db, createHostQuery, map[string]interface{}{
		"host_id": "8_8_8_8",
		"ip":      "8.8.8.8",
		"now":     time.Now().UTC(),
	})
	require.NoError(t, err)

	// Create geographic nodes
	_, err = workflow.createGeoNodes(geoData)
	require.NoError(t, err)

	// Create relationships
	result, err := workflow.createGeoRelationships(geoData)
	require.NoError(t, err)

	assert.Equal(t, 1, result.HostCityLinks)
	assert.Equal(t, 1, result.CityRegionLinks)
	assert.Equal(t, 1, result.RegionCountryLinks)
}

// TestEnrichGeoWorkflow_UpdateHostRecords tests host record updates
func TestEnrichGeoWorkflow_UpdateHostRecords(t *testing.T) {
	if os.Getenv("SKIP_INTEGRATION") != "" {
		t.Skip("Skipping integration test")
	}

	db, err := setupTestDB(t)
	if err != nil {
		t.Skipf("SurrealDB not available: %v", err)
	}
	defer db.Close(context.Background())

	logger, _ := zap.NewDevelopment()
	workflow := NewEnrichGeoWorkflow(db, nil, logger)

	// Create a host
	ctx := context.Background()
	createHostQuery := `
		CREATE type::thing('host', $host_id) CONTENT {
			ip: $ip,
			last_seen: $now
		};
	`
	_, err = surrealdb.Query[interface{}](ctx, db, createHostQuery, map[string]interface{}{
		"host_id": "8_8_8_8",
		"ip":      "8.8.8.8",
		"now":     time.Now().UTC(),
	})
	require.NoError(t, err)

	// Update with geo data
	geoData := map[string]*enrichment.GeoIPInfo{
		"8.8.8.8": {
			IP:        "8.8.8.8",
			City:      "Mountain View",
			Region:    "California",
			Country:   "United States",
			CountryCC: "US",
		},
	}

	err = workflow.updateHostRecords(geoData)
	require.NoError(t, err)

	// Verify update
	verifyQuery := `SELECT * FROM type::thing('host', $host_id);`
	result, err := surrealdb.Query[[]interface{}](ctx, db, verifyQuery, map[string]interface{}{
		"host_id": "8_8_8_8",
	})
	require.NoError(t, err)
	assert.NotNil(t, result)
}

// setupTestDB creates a test SurrealDB connection
func setupTestDB(t *testing.T) (*surrealdb.DB, error) {
	surrealURL := os.Getenv("SURREALDB_URL")
	if surrealURL == "" {
		surrealURL = "ws://localhost:8000/rpc"
	}

	db, err := surrealdb.New(surrealURL)
	if err != nil {
		return nil, err
	}

	// Use test namespace/database
	if _, err := db.SignIn(context.Background(), surrealdb.Auth{
		Username: "root",
		Password: "root",
	}); err != nil {
		db.Close(context.Background())
		return nil, err
	}

	testNS := "test_" + t.Name()
	testDB := "test_geo"

	if err := db.Use(context.Background(), testNS, testDB); err != nil {
		db.Close(context.Background())
		return nil, err
	}

	// Clean up test data when done
	t.Cleanup(func() {
		// Remove test namespace
		ctx := context.Background()
		surrealdb.Query[interface{}](ctx, db, "REMOVE NAMESPACE "+testNS, nil)
	})

	return db, nil
}

// getTestMMDBPath returns the path to a test MMDB file
func getTestMMDBPath() string {
	if path := os.Getenv("GEOIP_MMDB_PATH"); path != "" {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	commonPaths := []string{
		"/usr/share/GeoIP/GeoLite2-City.mmdb",
		"/var/lib/GeoIP/GeoLite2-City.mmdb",
		"./GeoLite2-City.mmdb",
	}

	for _, path := range commonPaths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return ""
}
