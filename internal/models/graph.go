package models

import "time"

// GraphQueryType represents the type of graph query to perform
type GraphQueryType string

const (
	QueryByASN      GraphQueryType = "by_asn"
	QueryByLocation GraphQueryType = "by_location"
	QueryByVuln     GraphQueryType = "by_vuln"
	QueryByService  GraphQueryType = "by_service"
)

// GraphQueryRequest represents the request for a graph traversal query
type GraphQueryRequest struct {
	QueryType GraphQueryType `json:"query_type" validate:"required,oneof=by_asn by_location by_vuln by_service"`

	// ASN query parameters
	ASN *int `json:"asn,omitempty"`

	// Location query parameters
	City    string `json:"city,omitempty"`
	Region  string `json:"region,omitempty"`
	Country string `json:"country,omitempty"`

	// Vulnerability query parameters
	CVE string `json:"cve,omitempty"`

	// Service query parameters
	Product string `json:"product,omitempty"`
	Service string `json:"service,omitempty"`

	// Pagination parameters
	Limit  int `json:"limit,omitempty"`  // Default: 100, Max: 1000
	Offset int `json:"offset,omitempty"` // Default: 0
}

// GraphQueryResponse represents the response from a graph traversal query
type GraphQueryResponse struct {
	Results    []HostResult       `json:"results"`
	Pagination PaginationMetadata `json:"pagination"`
	QueryTime  float64            `json:"query_time_ms"`
}

// HostResult represents a host returned from a graph query
type HostResult struct {
	ID        string    `json:"id"`
	IP        string    `json:"ip"`
	ASN       int       `json:"asn,omitempty"`
	City      string    `json:"city,omitempty"`
	Region    string    `json:"region,omitempty"`
	Country   string    `json:"country,omitempty"`
	Ports     []Port    `json:"ports,omitempty"`
	Services  []Service `json:"services,omitempty"`
	LastSeen  time.Time `json:"last_seen"`
	FirstSeen time.Time `json:"first_seen,omitempty"`
}

// Port represents a port on a host
type Port struct {
	ID       string `json:"id"`
	Number   int    `json:"number"`
	Protocol string `json:"protocol"` // tcp, udp
	State    string `json:"state"`    // open, closed, filtered
}

// Service represents a service running on a port
type Service struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Product string `json:"product,omitempty"`
	Version string `json:"version,omitempty"`
	CPE     string `json:"cpe,omitempty"`
}

// PaginationMetadata represents pagination information
type PaginationMetadata struct {
	Limit      int  `json:"limit"`
	Offset     int  `json:"offset"`
	Total      int  `json:"total"`
	HasMore    bool `json:"has_more"`
	NextOffset int  `json:"next_offset,omitempty"`
}

// Validate validates the GraphQueryRequest
func (r *GraphQueryRequest) Validate() error {
	// Validate query type
	switch r.QueryType {
	case QueryByASN:
		if r.ASN == nil {
			return ErrMissingASN
		}
	case QueryByLocation:
		if r.City == "" && r.Region == "" && r.Country == "" {
			return ErrMissingLocation
		}
	case QueryByVuln:
		if r.CVE == "" {
			return ErrMissingCVE
		}
	case QueryByService:
		if r.Product == "" && r.Service == "" {
			return ErrMissingService
		}
	default:
		return ErrInvalidQueryType
	}

	// Validate and set pagination defaults
	if r.Limit <= 0 {
		r.Limit = DefaultLimit
	}
	if r.Limit > MaxLimit {
		r.Limit = MaxLimit
	}
	if r.Offset < 0 {
		r.Offset = 0
	}

	return nil
}

// Pagination constants
const (
	DefaultLimit = 100
	MaxLimit     = 1000
)

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Field + ": " + e.Message
}

// Validation errors
var (
	ErrInvalidQueryType = &ValidationError{Field: "query_type", Message: "invalid query type"}
	ErrMissingASN       = &ValidationError{Field: "asn", Message: "asn is required for by_asn queries"}
	ErrMissingLocation  = &ValidationError{Field: "location", Message: "at least one of city, region, or country is required"}
	ErrMissingCVE       = &ValidationError{Field: "cve", Message: "cve is required for by_vuln queries"}
	ErrMissingService   = &ValidationError{Field: "service", Message: "product or service is required for by_service queries"}
)
