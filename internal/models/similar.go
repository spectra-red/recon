package models

// SimilarRequest represents a request to search for similar vulnerability documents
type SimilarRequest struct {
	// Query is the natural language query string
	Query string `json:"query"`

	// K is the number of results to return (optional, default 10)
	K *int `json:"k,omitempty"`
}

// SimilarResponse represents the response from a similarity search
type SimilarResponse struct {
	// Query is the original query string
	Query string `json:"query"`

	// Results is the list of similar vulnerability documents
	Results []VulnResult `json:"results"`

	// Count is the number of results returned
	Count int `json:"count"`

	// Timestamp is when the search was performed
	Timestamp string `json:"timestamp"`
}

// VulnResult represents a single vulnerability search result with similarity score
type VulnResult struct {
	// CVEID is the CVE identifier
	CVEID string `json:"cve_id"`

	// Title is the vulnerability title
	Title string `json:"title"`

	// Summary is the vulnerability description/summary
	Summary string `json:"summary"`

	// CVSS is the CVSS score
	CVSS float64 `json:"cvss,omitempty"`

	// CPE is the list of affected CPEs
	CPE []string `json:"cpe,omitempty"`

	// PublishedDate is when the vulnerability was published
	PublishedDate string `json:"published_date,omitempty"`

	// Score is the similarity score (0.0 to 1.0, higher is more similar)
	Score float64 `json:"score"`
}

// ErrorResponse represents an error response from the API
type ErrorResponse struct {
	// Error is the error message
	Error string `json:"error"`

	// Code is the error code
	Code string `json:"code,omitempty"`

	// Details provides additional error context
	Details string `json:"details,omitempty"`

	// Timestamp is when the error occurred
	Timestamp string `json:"timestamp"`
}

// Validate validates a SimilarRequest
func (r *SimilarRequest) Validate() error {
	if r.Query == "" {
		return ErrEmptyQuery
	}

	if len(r.Query) > MaxQueryLength {
		return ErrQueryTooLong
	}

	// Validate K if provided
	if r.K != nil {
		if *r.K < 1 {
			return ErrInvalidK
		}
		if *r.K > MaxK {
			return ErrKTooLarge
		}
	}

	return nil
}

// GetK returns the K value or the default if not set
func (r *SimilarRequest) GetK() int {
	if r.K == nil {
		return DefaultK
	}
	return *r.K
}

// Constants for validation and defaults
const (
	// DefaultK is the default number of results to return
	DefaultK = 10

	// MaxK is the maximum number of results allowed
	MaxK = 50

	// MaxQueryLength is the maximum query string length
	MaxQueryLength = 500
)

// Error types for validation
var (
	// ErrEmptyQuery indicates the query is empty
	ErrEmptyQuery = &ValidationError{Field: "query", Message: "query cannot be empty"}

	// ErrQueryTooLong indicates the query exceeds max length
	ErrQueryTooLong = &ValidationError{Field: "query", Message: "query exceeds maximum length"}

	// ErrInvalidK indicates K is invalid
	ErrInvalidK = &ValidationError{Field: "k", Message: "k must be greater than 0"}

	// ErrKTooLarge indicates K exceeds maximum
	ErrKTooLarge = &ValidationError{Field: "k", Message: "k exceeds maximum allowed value"}
)
