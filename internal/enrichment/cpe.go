package enrichment

import (
	"crypto/sha256"
	"fmt"
	"regexp"
	"strings"
)

// CPEIdentifier represents a Common Platform Enumeration identifier
type CPEIdentifier struct {
	Vendor  string `json:"vendor"`
	Product string `json:"product"`
	Version string `json:"version"`
	CPE     string `json:"cpe"` // Full CPE 2.3 string
}

// ServiceInfo represents service data for CPE generation
type ServiceInfo struct {
	ID       string `json:"id"`        // Service record ID
	Name     string `json:"name"`      // Service name (http, ssh, etc.)
	Product  string `json:"product"`   // Product name (nginx, openssh, etc.)
	Version  string `json:"version"`   // Version string
	Banner   string `json:"banner"`    // Raw banner text
}

// BannerPattern represents a regex pattern for parsing service banners
type BannerPattern struct {
	Regex   *regexp.Regexp
	Vendor  string // Fixed vendor name (if not captured)
	Product string // Fixed product name (if not captured)
}

// Common banner patterns for CPE generation
var bannerPatterns = []BannerPattern{
	// SSH patterns
	{
		Regex:   regexp.MustCompile(`SSH-[\d.]+-OpenSSH[_-]([\d.p]+)`),
		Vendor:  "openbsd",
		Product: "openssh",
	},
	{
		Regex:   regexp.MustCompile(`SSH-[\d.]+-Cisco-[\d.]+-(.+)`),
		Vendor:  "cisco",
		Product: "ssh",
	},

	// HTTP Server patterns
	{
		Regex:   regexp.MustCompile(`nginx/([\d.]+)`),
		Vendor:  "nginx",
		Product: "nginx",
	},
	{
		Regex:   regexp.MustCompile(`Apache/([\d.]+)`),
		Vendor:  "apache",
		Product: "http_server",
	},
	{
		Regex:   regexp.MustCompile(`Microsoft-IIS/([\d.]+)`),
		Vendor:  "microsoft",
		Product: "iis",
	},
	{
		Regex:   regexp.MustCompile(`lighttpd/([\d.]+)`),
		Vendor:  "lighttpd",
		Product: "lighttpd",
	},
	{
		Regex:   regexp.MustCompile(`Caddy\s+v?([\d.]+)`),
		Vendor:  "caddyserver",
		Product: "caddy",
	},

	// Database patterns
	{
		Regex:   regexp.MustCompile(`MySQL/([\d.]+)`),
		Vendor:  "mysql",
		Product: "mysql",
	},
	{
		Regex:   regexp.MustCompile(`PostgreSQL\s+([\d.]+)`),
		Vendor:  "postgresql",
		Product: "postgresql",
	},
	{
		Regex:   regexp.MustCompile(`MariaDB-([\d.]+)`),
		Vendor:  "mariadb",
		Product: "mariadb",
	},
	{
		Regex:   regexp.MustCompile(`MongoDB\s+([\d.]+)`),
		Vendor:  "mongodb",
		Product: "mongodb",
	},
	{
		Regex:   regexp.MustCompile(`Redis\s+server\s+v=([\d.]+)`),
		Vendor:  "redis",
		Product: "redis",
	},

	// Application servers
	{
		Regex:   regexp.MustCompile(`Tomcat/([\d.]+)`),
		Vendor:  "apache",
		Product: "tomcat",
	},
	{
		Regex:   regexp.MustCompile(`Jetty\(?([\d.]+)`),
		Vendor:  "eclipse",
		Product: "jetty",
	},

	// FTP patterns
	{
		Regex:   regexp.MustCompile(`ProFTPD\s+([\d.]+)`),
		Vendor:  "proftpd",
		Product: "proftpd",
	},
	{
		Regex:   regexp.MustCompile(`vsftpd\s+([\d.]+)`),
		Vendor:  "vsftpd_project",
		Product: "vsftpd",
	},

	// DNS patterns
	{
		Regex:   regexp.MustCompile(`BIND\s+([\d.]+)`),
		Vendor:  "isc",
		Product: "bind",
	},
	{
		Regex:   regexp.MustCompile(`dnsmasq-([\d.]+)`),
		Vendor:  "thekelleys",
		Product: "dnsmasq",
	},

	// Mail servers
	{
		Regex:   regexp.MustCompile(`Postfix\s+([\d.]+)`),
		Vendor:  "postfix",
		Product: "postfix",
	},
	{
		Regex:   regexp.MustCompile(`Exim\s+([\d.]+)`),
		Vendor:  "exim",
		Product: "exim",
	},
	{
		Regex:   regexp.MustCompile(`Sendmail/([\d.]+)`),
		Vendor:  "sendmail",
		Product: "sendmail",
	},

	// Proxy/Cache servers
	{
		Regex:   regexp.MustCompile(`squid/([\d.]+)`),
		Vendor:  "squid-cache",
		Product: "squid",
	},
	{
		Regex:   regexp.MustCompile(`Varnish/([\d.]+)`),
		Vendor:  "varnish-cache",
		Product: "varnish",
	},
}

// ProductVendorMap provides vendor mapping for products when not in banner
var ProductVendorMap = map[string]string{
	"nginx":      "nginx",
	"apache":     "apache",
	"openssh":    "openbsd",
	"mysql":      "mysql",
	"postgresql": "postgresql",
	"mariadb":    "mariadb",
	"mongodb":    "mongodb",
	"redis":      "redis",
	"elasticsearch": "elastic",
	"kibana":     "elastic",
	"logstash":   "elastic",
	"php":        "php",
	"python":     "python",
	"node":       "nodejs",
	"tomcat":     "apache",
	"jetty":      "eclipse",
	"iis":        "microsoft",
	"openssl":    "openssl",
	"bind":       "isc",
	"postfix":    "postfix",
	"dovecot":    "dovecot",
}

// ParseBanner extracts product and version information from a service banner
func ParseBanner(banner string) (product string, version string, vendor string) {
	if banner == "" {
		return "", "", ""
	}

	// Try each pattern
	for _, pattern := range bannerPatterns {
		matches := pattern.Regex.FindStringSubmatch(banner)
		if len(matches) >= 2 {
			return pattern.Product, matches[1], pattern.Vendor
		}
	}

	return "", "", ""
}

// GenerateCPE creates a CPE 2.3 identifier from service information
func GenerateCPE(service ServiceInfo) []CPEIdentifier {
	var cpes []CPEIdentifier

	// Strategy 1: Use existing product/version from service record
	if service.Product != "" && service.Version != "" {
		vendor := normalizeVendor(service.Product)
		cpe := formatCPE23(vendor, service.Product, service.Version)
		cpes = append(cpes, CPEIdentifier{
			Vendor:  vendor,
			Product: service.Product,
			Version: service.Version,
			CPE:     cpe,
		})
	}

	// Strategy 2: Parse banner if available
	if service.Banner != "" {
		product, version, vendor := ParseBanner(service.Banner)
		if product != "" && version != "" {
			cpe := formatCPE23(vendor, product, version)
			// Only add if different from strategy 1
			if !containsCPE(cpes, cpe) {
				cpes = append(cpes, CPEIdentifier{
					Vendor:  vendor,
					Product: product,
					Version: version,
					CPE:     cpe,
				})
			}
		}
	}

	// Strategy 3: Generate fuzzy CPE without version (for broader matching)
	if service.Product != "" && service.Version == "" {
		vendor := normalizeVendor(service.Product)
		cpe := formatCPE23(vendor, service.Product, "*")
		if !containsCPE(cpes, cpe) {
			cpes = append(cpes, CPEIdentifier{
				Vendor:  vendor,
				Product: service.Product,
				Version: "*",
				CPE:     cpe,
			})
		}
	}

	return cpes
}

// GenerateCPEBatch generates CPEs for multiple services
func GenerateCPEBatch(services []ServiceInfo) map[string][]CPEIdentifier {
	result := make(map[string][]CPEIdentifier)

	for _, service := range services {
		cpes := GenerateCPE(service)
		if len(cpes) > 0 {
			result[service.ID] = cpes
		}
	}

	return result
}

// normalizeVendor attempts to determine the vendor from the product name
func normalizeVendor(product string) string {
	// Normalize product name (lowercase, remove special chars)
	normalized := strings.ToLower(strings.TrimSpace(product))

	// Check product-to-vendor map
	if vendor, exists := ProductVendorMap[normalized]; exists {
		return vendor
	}

	// Default: use product name as vendor
	return normalized
}

// formatCPE23 creates a CPE 2.3 formatted string
// Format: cpe:2.3:a:vendor:product:version:*:*:*:*:*:*:*
func formatCPE23(vendor, product, version string) string {
	// Normalize components
	vendor = normalizeCPEComponent(vendor)
	product = normalizeCPEComponent(product)
	version = normalizeCPEComponent(version)

	// CPE 2.3 format for applications (part = 'a')
	return fmt.Sprintf("cpe:2.3:a:%s:%s:%s:*:*:*:*:*:*:*", vendor, product, version)
}

// normalizeCPEComponent normalizes a CPE component according to CPE 2.3 spec
func normalizeCPEComponent(s string) string {
	if s == "" || s == "*" {
		return "*"
	}

	// Lowercase
	s = strings.ToLower(s)

	// Replace spaces with underscores
	s = strings.ReplaceAll(s, " ", "_")

	// Remove/replace special characters that aren't allowed in CPE
	s = regexp.MustCompile(`[^a-z0-9._\-]`).ReplaceAllString(s, "")

	return s
}

// containsCPE checks if a CPE string already exists in a slice
func containsCPE(cpes []CPEIdentifier, cpe string) bool {
	for _, c := range cpes {
		if c.CPE == cpe {
			return true
		}
	}
	return false
}

// GenerateServiceFingerprint creates a unique fingerprint for service deduplication
func GenerateServiceFingerprint(name, product, version string) string {
	// Create a stable hash from service attributes
	data := fmt.Sprintf("%s|%s|%s",
		strings.ToLower(strings.TrimSpace(name)),
		strings.ToLower(strings.TrimSpace(product)),
		strings.ToLower(strings.TrimSpace(version)),
	)

	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("%x", hash)
}

// MatchesVersionRange checks if a version matches a version range specification
// Supports basic version comparisons for vulnerability matching
func MatchesVersionRange(version, rangeSpec string) bool {
	// If no range specified or wildcard, match everything
	if rangeSpec == "" || rangeSpec == "*" {
		return true
	}

	// Exact match
	if version == rangeSpec {
		return true
	}

	// TODO: Implement proper version range parsing (e.g., ">=1.2.0,<2.0.0")
	// For MVP, we use exact matching only
	// Future: Use github.com/hashicorp/go-version for semantic version comparison

	return false
}

// ExtractVersionComponents parses a version string into major, minor, patch components
func ExtractVersionComponents(version string) (major, minor, patch string) {
	// Remove common prefixes
	version = strings.TrimPrefix(version, "v")
	version = strings.TrimPrefix(version, "V")

	// Remove build metadata (anything after + or -)
	if idx := strings.IndexAny(version, "+-"); idx != -1 {
		version = version[:idx]
	}

	// Split on dots
	parts := strings.Split(version, ".")

	if len(parts) > 0 {
		major = parts[0]
	}
	if len(parts) > 1 {
		minor = parts[1]
	}
	if len(parts) > 2 {
		patch = parts[2]
	}

	return major, minor, patch
}
