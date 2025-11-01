package db

import (
	"context"
	"fmt"
	"time"

	"github.com/spectra-red/recon/internal/models"
	"github.com/surrealdb/surrealdb.go"
	"go.uber.org/zap"
)

// QueryHost retrieves host information with graph traversal based on depth
// Depth levels:
//
//	0: Host only
//	1: Host + Ports
//	2: Host + Ports + Services (default)
//	3: Host + Ports + Services + Vulnerabilities
//	4-5: Extended relationships
func QueryHost(ctx context.Context, db *surrealdb.DB, logger *zap.Logger, ip string, depth int) (*models.HostQueryResponse, error) {
	// Validate depth
	if !models.ValidateDepth(depth) {
		return nil, fmt.Errorf("invalid depth: %d (must be 0-5)", depth)
	}

	// Build the query based on depth
	query := buildHostQuery(ip, depth)

	logger.Debug("executing host query",
		zap.String("ip", ip),
		zap.Int("depth", depth),
		zap.String("query", query))

	// Execute query using the SurrealDB Query function
	// Note: The result structure from SurrealDB varies based on the query
	result, err := surrealdb.Query[map[string]interface{}](ctx, db, query, map[string]interface{}{
		"ip": ip,
	})
	if err != nil {
		logger.Error("query execution failed",
			zap.Error(err),
			zap.String("ip", ip))
		return nil, fmt.Errorf("failed to query host: %w", err)
	}

	// Check if query result is empty
	if result == nil || len(*result) == 0 {
		logger.Debug("host not found",
			zap.String("ip", ip))
		return nil, nil
	}

	// Get the first query result
	queryResult := (*result)[0]
	if queryResult.Error != nil {
		logger.Error("query returned error",
			zap.Error(queryResult.Error),
			zap.String("ip", ip))
		return nil, fmt.Errorf("query error: %w", queryResult.Error)
	}

	if queryResult.Result == nil {
		logger.Debug("host not found",
			zap.String("ip", ip))
		return nil, nil
	}

	// Parse result based on depth
	response, err := parseHostQueryResult(queryResult.Result, depth, logger)
	if err != nil {
		logger.Error("failed to parse query result",
			zap.Error(err),
			zap.String("ip", ip))
		return nil, fmt.Errorf("failed to parse result: %w", err)
	}

	// Check if host was found
	if response == nil {
		logger.Debug("host not found",
			zap.String("ip", ip))
		return nil, nil
	}

	return response, nil
}

// buildHostQuery constructs the SurrealDB query based on depth
func buildHostQuery(ip string, depth int) string {
	// Base query - always get host
	query := `SELECT * FROM host WHERE ip = $ip`

	// Add FETCH clauses based on depth
	if depth >= 1 {
		// Depth 1: Include ports
		query = `SELECT *,
			->HAS->port.* AS ports
		FROM host WHERE ip = $ip`
	}

	if depth >= 2 {
		// Depth 2: Include ports and services
		query = `SELECT *,
			->HAS->port.* AS ports,
			->HAS->port->RUNS->service.* AS services
		FROM host WHERE ip = $ip`
	}

	if depth >= 3 {
		// Depth 3: Include ports, services, and vulnerabilities
		query = `SELECT *,
			->HAS->port.* AS ports,
			->HAS->port->RUNS->service.* AS services,
			->HAS->port->RUNS->service->AFFECTED_BY->vuln.* AS vulns
		FROM host WHERE ip = $ip`
	}

	if depth >= 4 {
		// Depth 4+: Include extended relationships (geographic, ASN)
		query = `SELECT *,
			->HAS->port.* AS ports,
			->HAS->port->RUNS->service.* AS services,
			->HAS->port->RUNS->service->AFFECTED_BY->vuln.* AS vulns,
			->IN_CITY->city.* AS city_detail,
			->IN_ASN->asn.* AS asn_detail
		FROM host WHERE ip = $ip`
	}

	return query + " LIMIT 1;"
}

// parseHostQueryResult parses the SurrealDB result into HostQueryResponse
func parseHostQueryResult(result map[string]interface{}, depth int, logger *zap.Logger) (*models.HostQueryResponse, error) {
	// The surrealdb.Query[T] function returns the structured result directly
	// Check if we have any data
	if result == nil || len(result) == 0 {
		// No host found
		return nil, nil
	}

	// Use result directly as hostData
	hostData := result

	// Parse host fields
	response := &models.HostQueryResponse{
		IP: getStringField(hostData, "ip"),
	}

	// Parse optional host fields
	if asn, ok := getIntField(hostData, "asn"); ok {
		response.ASN = asn
	}
	if city, ok := hostData["city"].(string); ok {
		response.City = city
	}
	if region, ok := hostData["region"].(string); ok {
		response.Region = region
	}
	if country, ok := hostData["country"].(string); ok {
		response.Country = country
	}
	if cloudRegion, ok := hostData["cloud_region"].(string); ok {
		response.CloudRegion = cloudRegion
	}

	// Parse timestamps
	if firstSeen, err := parseTimeField(hostData, "first_seen"); err == nil {
		response.FirstSeen = firstSeen
	}
	if lastSeen, err := parseTimeField(hostData, "last_seen"); err == nil {
		response.LastSeen = lastSeen
	}

	// Parse depth-specific fields
	if depth >= 1 {
		// Parse ports
		if ports, ok := hostData["ports"].([]interface{}); ok {
			response.Ports = parsePorts(ports, depth, logger)
		}
	}

	if depth >= 2 {
		// Parse services
		if services, ok := hostData["services"].([]interface{}); ok {
			response.Services = parseServices(services, depth, logger)
		}
	}

	if depth >= 3 {
		// Parse vulnerabilities
		if vulns, ok := hostData["vulns"].([]interface{}); ok {
			response.Vulns = parseVulns(vulns, logger)
		}
	}

	return response, nil
}

// parsePorts extracts port information from query result
func parsePorts(portsData []interface{}, depth int, logger *zap.Logger) []models.PortDetail {
	ports := make([]models.PortDetail, 0, len(portsData))

	for _, portItem := range portsData {
		portMap, ok := portItem.(map[string]interface{})
		if !ok {
			logger.Warn("invalid port data type", zap.Any("port", portItem))
			continue
		}

		port := models.PortDetail{
			Protocol: getStringField(portMap, "protocol"),
		}

		if number, ok := getIntField(portMap, "number"); ok {
			port.Number = number
		}
		if transport, ok := portMap["transport"].(string); ok {
			port.Transport = transport
		}
		if firstSeen, err := parseTimeField(portMap, "first_seen"); err == nil {
			port.FirstSeen = firstSeen
		}
		if lastSeen, err := parseTimeField(portMap, "last_seen"); err == nil {
			port.LastSeen = lastSeen
		}

		ports = append(ports, port)
	}

	return ports
}

// parseServices extracts service information from query result
func parseServices(servicesData []interface{}, depth int, logger *zap.Logger) []models.ServiceDetail {
	services := make([]models.ServiceDetail, 0, len(servicesData))

	for _, serviceItem := range servicesData {
		serviceMap, ok := serviceItem.(map[string]interface{})
		if !ok {
			logger.Warn("invalid service data type", zap.Any("service", serviceItem))
			continue
		}

		service := models.ServiceDetail{
			Name:    getStringField(serviceMap, "name"),
			Product: getStringField(serviceMap, "product"),
			Version: getStringField(serviceMap, "version"),
		}

		// Parse CPE array
		if cpeData, ok := serviceMap["cpe"].([]interface{}); ok {
			cpe := make([]string, 0, len(cpeData))
			for _, c := range cpeData {
				if cpeStr, ok := c.(string); ok {
					cpe = append(cpe, cpeStr)
				}
			}
			service.CPE = cpe
		}

		if firstSeen, err := parseTimeField(serviceMap, "first_seen"); err == nil {
			service.FirstSeen = firstSeen
		}
		if lastSeen, err := parseTimeField(serviceMap, "last_seen"); err == nil {
			service.LastSeen = lastSeen
		}

		services = append(services, service)
	}

	return services
}

// parseVulns extracts vulnerability information from query result
func parseVulns(vulnsData []interface{}, logger *zap.Logger) []models.VulnDetail {
	vulns := make([]models.VulnDetail, 0, len(vulnsData))

	for _, vulnItem := range vulnsData {
		vulnMap, ok := vulnItem.(map[string]interface{})
		if !ok {
			logger.Warn("invalid vuln data type", zap.Any("vuln", vulnItem))
			continue
		}

		vuln := models.VulnDetail{
			CVEID:    getStringField(vulnMap, "cve_id"),
			Severity: getStringField(vulnMap, "severity"),
		}

		if cvss, ok := getFloatField(vulnMap, "cvss"); ok {
			vuln.CVSS = cvss
		}
		if kevFlag, ok := vulnMap["kev_flag"].(bool); ok {
			vuln.KEVFlag = kevFlag
		}
		if firstSeen, err := parseTimeField(vulnMap, "first_seen"); err == nil {
			vuln.FirstSeen = firstSeen
		}

		vulns = append(vulns, vuln)
	}

	return vulns
}

// Helper functions for type conversion

func getStringField(data map[string]interface{}, key string) string {
	if val, ok := data[key].(string); ok {
		return val
	}
	return ""
}

func getIntField(data map[string]interface{}, key string) (int, bool) {
	switch val := data[key].(type) {
	case int:
		return val, true
	case int64:
		return int(val), true
	case float64:
		return int(val), true
	}
	return 0, false
}

func getFloatField(data map[string]interface{}, key string) (float64, bool) {
	switch val := data[key].(type) {
	case float64:
		return val, true
	case float32:
		return float64(val), true
	case int:
		return float64(val), true
	case int64:
		return float64(val), true
	}
	return 0, false
}

func parseTimeField(data map[string]interface{}, key string) (time.Time, error) {
	val, ok := data[key]
	if !ok {
		return time.Time{}, fmt.Errorf("field %s not found", key)
	}

	switch t := val.(type) {
	case time.Time:
		return t, nil
	case string:
		// Try parsing ISO8601
		parsed, err := time.Parse(time.RFC3339, t)
		if err != nil {
			return time.Time{}, fmt.Errorf("failed to parse time: %w", err)
		}
		return parsed, nil
	default:
		return time.Time{}, fmt.Errorf("unsupported time type: %T", val)
	}
}
