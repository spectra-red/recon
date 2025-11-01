package db

import (
	"context"
	"fmt"
	"time"

	"github.com/spectra-red/recon/internal/models"
	"github.com/surrealdb/surrealdb.go"
	"go.uber.org/zap"
)

// GraphQueryExecutor handles graph traversal queries against SurrealDB
type GraphQueryExecutor struct {
	db     *surrealdb.DB
	logger *zap.Logger
}

// NewGraphQueryExecutor creates a new graph query executor
func NewGraphQueryExecutor(db *surrealdb.DB, logger *zap.Logger) *GraphQueryExecutor {
	return &GraphQueryExecutor{
		db:     db,
		logger: logger,
	}
}

// ExecuteGraphQuery executes a graph traversal query based on the query type
func (e *GraphQueryExecutor) ExecuteGraphQuery(ctx context.Context, req models.GraphQueryRequest) (*models.GraphQueryResponse, error) {
	startTime := time.Now()

	// Validate request
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	// Add timeout to context if not already set
	_, hasDeadline := ctx.Deadline()
	if !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
	}

	// Execute query based on type
	var results []models.HostResult
	var total int
	var err error

	switch req.QueryType {
	case models.QueryByASN:
		results, total, err = e.queryByASN(ctx, *req.ASN, req.Limit, req.Offset)
	case models.QueryByLocation:
		results, total, err = e.queryByLocation(ctx, req.City, req.Region, req.Country, req.Limit, req.Offset)
	case models.QueryByVuln:
		results, total, err = e.queryByVuln(ctx, req.CVE, req.Limit, req.Offset)
	case models.QueryByService:
		results, total, err = e.queryByService(ctx, req.Product, req.Service, req.Limit, req.Offset)
	default:
		return nil, fmt.Errorf("unsupported query type: %s", req.QueryType)
	}

	if err != nil {
		return nil, err
	}

	// Calculate query time
	queryTime := time.Since(startTime).Seconds() * 1000 // Convert to milliseconds

	// Log slow queries
	if queryTime > 1000 {
		e.logger.Warn("slow query detected",
			zap.String("query_type", string(req.QueryType)),
			zap.Float64("query_time_ms", queryTime),
			zap.Int("result_count", len(results)))
	}

	// Build pagination metadata
	hasMore := total > (req.Offset + len(results))
	nextOffset := 0
	if hasMore {
		nextOffset = req.Offset + req.Limit
	}

	return &models.GraphQueryResponse{
		Results: results,
		Pagination: models.PaginationMetadata{
			Limit:      req.Limit,
			Offset:     req.Offset,
			Total:      total,
			HasMore:    hasMore,
			NextOffset: nextOffset,
		},
		QueryTime: queryTime,
	}, nil
}

// queryByASN returns all hosts in a given ASN
func (e *GraphQueryExecutor) queryByASN(ctx context.Context, asn, limit, offset int) ([]models.HostResult, int, error) {
	e.logger.Debug("executing ASN query",
		zap.Int("asn", asn),
		zap.Int("limit", limit),
		zap.Int("offset", offset))

	query := `
		SELECT
			id,
			ip,
			asn,
			city,
			region,
			country,
			last_seen,
			first_seen
		FROM host
		WHERE asn = $asn
		ORDER BY last_seen DESC
		LIMIT $limit
		START $offset
	`

	params := map[string]interface{}{
		"asn":    asn,
		"limit":  limit,
		"offset": offset,
	}

	result, err := surrealdb.Query[[]models.HostResult](ctx, e.db, query, params)
	if err != nil {
		e.logger.Error("failed to execute ASN query",
			zap.Error(err),
			zap.Int("asn", asn))
		return nil, 0, fmt.Errorf("failed to query by ASN: %w", err)
	}

	hosts := extractHostResults(result)
	total := len(hosts) // Simplified: use result count as total

	return hosts, total, nil
}

// queryByLocation returns all hosts in a given location
func (e *GraphQueryExecutor) queryByLocation(ctx context.Context, city, region, country string, limit, offset int) ([]models.HostResult, int, error) {
	e.logger.Debug("executing location query",
		zap.String("city", city),
		zap.String("region", region),
		zap.String("country", country))

	var whereClause string
	params := map[string]interface{}{
		"limit":  limit,
		"offset": offset,
	}

	if city != "" {
		whereClause = "WHERE city = $city"
		params["city"] = city
	} else if region != "" {
		whereClause = "WHERE region = $region"
		params["region"] = region
	} else if country != "" {
		whereClause = "WHERE country = $country"
		params["country"] = country
	}

	query := fmt.Sprintf(`
		SELECT
			id,
			ip,
			asn,
			city,
			region,
			country,
			last_seen,
			first_seen
		FROM host
		%s
		ORDER BY last_seen DESC
		LIMIT $limit
		START $offset
	`, whereClause)

	result, err := surrealdb.Query[[]models.HostResult](ctx, e.db, query, params)
	if err != nil {
		e.logger.Error("failed to execute location query", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to query by location: %w", err)
	}

	hosts := extractHostResults(result)
	total := len(hosts)

	return hosts, total, nil
}

// queryByVuln returns all hosts affected by a given vulnerability
func (e *GraphQueryExecutor) queryByVuln(ctx context.Context, cve string, limit, offset int) ([]models.HostResult, int, error) {
	e.logger.Debug("executing vulnerability query",
		zap.String("cve", cve))

	query := `
		SELECT
			id,
			ip,
			asn,
			city,
			region,
			country,
			last_seen,
			first_seen
		FROM host
		WHERE id IN (
			SELECT VALUE <-HAS<-port<-RUNS<-service<-AFFECTED_BY<-vuln.id
			FROM vuln
			WHERE cve = $cve
		)
		LIMIT $limit
		START $offset
	`

	params := map[string]interface{}{
		"cve":    cve,
		"limit":  limit,
		"offset": offset,
	}

	result, err := surrealdb.Query[[]models.HostResult](ctx, e.db, query, params)
	if err != nil {
		e.logger.Error("failed to execute vulnerability query", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to query by vulnerability: %w", err)
	}

	hosts := extractHostResults(result)
	total := len(hosts)

	return hosts, total, nil
}

// queryByService returns all hosts running a given service
func (e *GraphQueryExecutor) queryByService(ctx context.Context, product, serviceName string, limit, offset int) ([]models.HostResult, int, error) {
	e.logger.Debug("executing service query",
		zap.String("product", product),
		zap.String("service", serviceName))

	var whereClause string
	params := map[string]interface{}{
		"limit":  limit,
		"offset": offset,
	}

	if product != "" {
		whereClause = "WHERE product = $product"
		params["product"] = product
	} else {
		whereClause = "WHERE name = $service"
		params["service"] = serviceName
	}

	query := fmt.Sprintf(`
		SELECT
			id,
			ip,
			asn,
			city,
			region,
			country,
			last_seen,
			first_seen
		FROM host
		WHERE id IN (
			SELECT VALUE <-HAS<-port<-RUNS<-service.id
			FROM service
			%s
		)
		LIMIT $limit
		START $offset
	`, whereClause)

	result, err := surrealdb.Query[[]models.HostResult](ctx, e.db, query, params)
	if err != nil {
		e.logger.Error("failed to execute service query", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to query by service: %w", err)
	}

	hosts := extractHostResults(result)
	total := len(hosts)

	return hosts, total, nil
}

// extractHostResults extracts host results from SurrealDB query response
func extractHostResults(results *[]surrealdb.QueryResult[[]models.HostResult]) []models.HostResult {
	if results == nil || len(*results) == 0 {
		return []models.HostResult{}
	}

	// Get the first query result
	queryResult := (*results)[0]
	if queryResult.Error != nil || queryResult.Result == nil {
		return []models.HostResult{}
	}

	return queryResult.Result
}
