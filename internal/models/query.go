package models

import (
	"time"
)

// HostQueryResponse represents the complete response for a host query
type HostQueryResponse struct {
	IP          string          `json:"ip"`
	ASN         int             `json:"asn,omitempty"`
	City        string          `json:"city,omitempty"`
	Region      string          `json:"region,omitempty"`
	Country     string          `json:"country,omitempty"`
	CloudRegion string          `json:"cloud_region,omitempty"`
	FirstSeen   time.Time       `json:"first_seen"`
	LastSeen    time.Time       `json:"last_seen"`
	Ports       []PortDetail    `json:"ports,omitempty"`
	Services    []ServiceDetail `json:"services,omitempty"`
	Vulns       []VulnDetail    `json:"vulnerabilities,omitempty"`
}

// PortDetail represents a port with its relationships
type PortDetail struct {
	Number    int             `json:"number"`
	Protocol  string          `json:"protocol"`
	Transport string          `json:"transport,omitempty"`
	FirstSeen time.Time       `json:"first_seen"`
	LastSeen  time.Time       `json:"last_seen"`
	Services  []ServiceDetail `json:"services,omitempty"`
}

// ServiceDetail represents a service with its metadata
type ServiceDetail struct {
	Name       string       `json:"name"`
	Product    string       `json:"product,omitempty"`
	Version    string       `json:"version,omitempty"`
	CPE        []string     `json:"cpe,omitempty"`
	Confidence float64      `json:"confidence,omitempty"`
	FirstSeen  time.Time    `json:"first_seen"`
	LastSeen   time.Time    `json:"last_seen"`
	Vulns      []VulnDetail `json:"vulnerabilities,omitempty"`
}

// VulnDetail represents vulnerability information
type VulnDetail struct {
	CVEID      string    `json:"cve_id"`
	CVSS       float64   `json:"cvss"`
	Severity   string    `json:"severity"`
	KEVFlag    bool      `json:"kev_flag"`
	Confidence float64   `json:"confidence,omitempty"`
	FirstSeen  time.Time `json:"first_detected"`
}

// QueryDepth represents the valid depth levels for graph traversal
type QueryDepth int

const (
	// DepthHostOnly returns only host information (depth 0)
	DepthHostOnly QueryDepth = 0
	// DepthWithPorts returns host + ports (depth 1)
	DepthWithPorts QueryDepth = 1
	// DepthWithServices returns host + ports + services (depth 2, default)
	DepthWithServices QueryDepth = 2
	// DepthWithVulns returns host + ports + services + vulnerabilities (depth 3)
	DepthWithVulns QueryDepth = 3
	// DepthMaximum is the maximum allowed depth (5)
	DepthMaximum QueryDepth = 5
)

// ValidateDepth checks if the depth is within acceptable range
func ValidateDepth(depth int) bool {
	return depth >= 0 && depth <= int(DepthMaximum)
}

// DefaultDepth returns the default query depth
func DefaultDepth() QueryDepth {
	return DepthWithServices
}
