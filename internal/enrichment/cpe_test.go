package enrichment

import (
	"testing"
)

func TestParseBanner(t *testing.T) {
	tests := []struct {
		name            string
		banner          string
		wantProduct     string
		wantVersion     string
		wantVendor      string
	}{
		{
			name:        "OpenSSH standard",
			banner:      "SSH-2.0-OpenSSH_9.0",
			wantProduct: "openssh",
			wantVersion: "9.0",
			wantVendor:  "openbsd",
		},
		{
			name:        "OpenSSH with patch",
			banner:      "SSH-2.0-OpenSSH_9.0p1",
			wantProduct: "openssh",
			wantVersion: "9.0p1",
			wantVendor:  "openbsd",
		},
		{
			name:        "nginx",
			banner:      "nginx/1.24.0",
			wantProduct: "nginx",
			wantVersion: "1.24.0",
			wantVendor:  "nginx",
		},
		{
			name:        "Apache",
			banner:      "Apache/2.4.57 (Unix)",
			wantProduct: "http_server",
			wantVersion: "2.4.57",
			wantVendor:  "apache",
		},
		{
			name:        "MySQL",
			banner:      "MySQL/8.0.35",
			wantProduct: "mysql",
			wantVersion: "8.0.35",
			wantVendor:  "mysql",
		},
		{
			name:        "PostgreSQL",
			banner:      "PostgreSQL 15.4",
			wantProduct: "postgresql",
			wantVersion: "15.4",
			wantVendor:  "postgresql",
		},
		{
			name:        "Redis",
			banner:      "Redis server v=7.0.12",
			wantProduct: "redis",
			wantVersion: "7.0.12",
			wantVendor:  "redis",
		},
		{
			name:        "Microsoft IIS",
			banner:      "Microsoft-IIS/10.0",
			wantProduct: "iis",
			wantVersion: "10.0",
			wantVendor:  "microsoft",
		},
		{
			name:        "empty banner",
			banner:      "",
			wantProduct: "",
			wantVersion: "",
			wantVendor:  "",
		},
		{
			name:        "unknown format",
			banner:      "Some Unknown Server",
			wantProduct: "",
			wantVersion: "",
			wantVendor:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotProduct, gotVersion, gotVendor := ParseBanner(tt.banner)

			if gotProduct != tt.wantProduct {
				t.Errorf("ParseBanner() product = %v, want %v", gotProduct, tt.wantProduct)
			}
			if gotVersion != tt.wantVersion {
				t.Errorf("ParseBanner() version = %v, want %v", gotVersion, tt.wantVersion)
			}
			if gotVendor != tt.wantVendor {
				t.Errorf("ParseBanner() vendor = %v, want %v", gotVendor, tt.wantVendor)
			}
		})
	}
}

func TestGenerateCPE(t *testing.T) {
	tests := []struct {
		name        string
		service     ServiceInfo
		wantCPEsLen int
		wantCPE     string // First expected CPE
	}{
		{
			name: "nginx with version",
			service: ServiceInfo{
				ID:      "svc1",
				Name:    "http",
				Product: "nginx",
				Version: "1.24.0",
				Banner:  "",
			},
			wantCPEsLen: 1,
			wantCPE:     "cpe:2.3:a:nginx:nginx:1.24.0:*:*:*:*:*:*:*",
		},
		{
			name: "openssh from banner",
			service: ServiceInfo{
				ID:      "svc2",
				Name:    "ssh",
				Product: "",
				Version: "",
				Banner:  "SSH-2.0-OpenSSH_9.0p1",
			},
			wantCPEsLen: 1,
			wantCPE:     "cpe:2.3:a:openbsd:openssh:9.0p1:*:*:*:*:*:*:*",
		},
		{
			name: "apache with banner and product",
			service: ServiceInfo{
				ID:      "svc3",
				Name:    "http",
				Product: "apache",
				Version: "2.4.57",
				Banner:  "Apache/2.4.57 (Unix)",
			},
			wantCPEsLen: 2, // One from product, one from banner
			wantCPE:     "cpe:2.3:a:apache:apache:2.4.57:*:*:*:*:*:*:*",
		},
		{
			name: "product without version",
			service: ServiceInfo{
				ID:      "svc4",
				Name:    "http",
				Product: "nginx",
				Version: "",
				Banner:  "",
			},
			wantCPEsLen: 1,
			wantCPE:     "cpe:2.3:a:nginx:nginx:*:*:*:*:*:*:*:*",
		},
		{
			name: "no product or banner",
			service: ServiceInfo{
				ID:      "svc5",
				Name:    "unknown",
				Product: "",
				Version: "",
				Banner:  "",
			},
			wantCPEsLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateCPE(tt.service)

			if len(got) != tt.wantCPEsLen {
				t.Errorf("GenerateCPE() returned %d CPEs, want %d", len(got), tt.wantCPEsLen)
			}

			if tt.wantCPEsLen > 0 && len(got) > 0 {
				if got[0].CPE != tt.wantCPE {
					t.Errorf("GenerateCPE() first CPE = %v, want %v", got[0].CPE, tt.wantCPE)
				}
			}
		})
	}
}

func TestFormatCPE23(t *testing.T) {
	tests := []struct {
		name    string
		vendor  string
		product string
		version string
		want    string
	}{
		{
			name:    "standard CPE",
			vendor:  "nginx",
			product: "nginx",
			version: "1.24.0",
			want:    "cpe:2.3:a:nginx:nginx:1.24.0:*:*:*:*:*:*:*",
		},
		{
			name:    "CPE with spaces in product",
			vendor:  "apache",
			product: "http server",
			version: "2.4.57",
			want:    "cpe:2.3:a:apache:http_server:2.4.57:*:*:*:*:*:*:*",
		},
		{
			name:    "wildcard version",
			vendor:  "mysql",
			product: "mysql",
			version: "*",
			want:    "cpe:2.3:a:mysql:mysql:*:*:*:*:*:*:*:*",
		},
		{
			name:    "version with patch",
			vendor:  "openbsd",
			product: "openssh",
			version: "9.0p1",
			want:    "cpe:2.3:a:openbsd:openssh:9.0p1:*:*:*:*:*:*:*",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatCPE23(tt.vendor, tt.product, tt.version)
			if got != tt.want {
				t.Errorf("formatCPE23() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNormalizeCPEComponent(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "lowercase conversion",
			input: "Nginx",
			want:  "nginx",
		},
		{
			name:  "space to underscore",
			input: "http server",
			want:  "http_server",
		},
		{
			name:  "remove special chars",
			input: "product@#$name",
			want:  "productname",
		},
		{
			name:  "preserve dots and dashes",
			input: "1.2.3-beta",
			want:  "1.2.3-beta",
		},
		{
			name:  "wildcard preserved",
			input: "*",
			want:  "*",
		},
		{
			name:  "empty string",
			input: "",
			want:  "*",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeCPEComponent(tt.input)
			if got != tt.want {
				t.Errorf("normalizeCPEComponent() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGenerateServiceFingerprint(t *testing.T) {
	tests := []struct {
		name    string
		svcName string
		product string
		version string
		want    string // We'll check it's not empty and consistent
	}{
		{
			name:    "standard service",
			svcName: "http",
			product: "nginx",
			version: "1.24.0",
		},
		{
			name:    "case insensitive",
			svcName: "HTTP",
			product: "NGINX",
			version: "1.24.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateServiceFingerprint(tt.svcName, tt.product, tt.version)

			// Check not empty
			if got == "" {
				t.Error("GenerateServiceFingerprint() returned empty string")
			}

			// Check consistent (same input = same output)
			got2 := GenerateServiceFingerprint(tt.svcName, tt.product, tt.version)
			if got != got2 {
				t.Errorf("GenerateServiceFingerprint() not consistent: %v != %v", got, got2)
			}

			// Check length (SHA256 hex = 64 chars)
			if len(got) != 64 {
				t.Errorf("GenerateServiceFingerprint() length = %v, want 64", len(got))
			}
		})
	}

	// Test that case doesn't matter
	fp1 := GenerateServiceFingerprint("http", "nginx", "1.24.0")
	fp2 := GenerateServiceFingerprint("HTTP", "NGINX", "1.24.0")
	if fp1 != fp2 {
		t.Error("GenerateServiceFingerprint() should be case-insensitive")
	}
}

func TestExtractVersionComponents(t *testing.T) {
	tests := []struct {
		name      string
		version   string
		wantMajor string
		wantMinor string
		wantPatch string
	}{
		{
			name:      "standard semver",
			version:   "1.2.3",
			wantMajor: "1",
			wantMinor: "2",
			wantPatch: "3",
		},
		{
			name:      "with v prefix",
			version:   "v2.4.57",
			wantMajor: "2",
			wantMinor: "4",
			wantPatch: "57",
		},
		{
			name:      "with build metadata",
			version:   "9.0.1+ubuntu",
			wantMajor: "9",
			wantMinor: "0",
			wantPatch: "1",
		},
		{
			name:      "major.minor only",
			version:   "8.0",
			wantMajor: "8",
			wantMinor: "0",
			wantPatch: "",
		},
		{
			name:      "major only",
			version:   "15",
			wantMajor: "15",
			wantMinor: "",
			wantPatch: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMajor, gotMinor, gotPatch := ExtractVersionComponents(tt.version)

			if gotMajor != tt.wantMajor {
				t.Errorf("ExtractVersionComponents() major = %v, want %v", gotMajor, tt.wantMajor)
			}
			if gotMinor != tt.wantMinor {
				t.Errorf("ExtractVersionComponents() minor = %v, want %v", gotMinor, tt.wantMinor)
			}
			if gotPatch != tt.wantPatch {
				t.Errorf("ExtractVersionComponents() patch = %v, want %v", gotPatch, tt.wantPatch)
			}
		})
	}
}

func TestGenerateCPEBatch(t *testing.T) {
	services := []ServiceInfo{
		{
			ID:      "svc1",
			Product: "nginx",
			Version: "1.24.0",
		},
		{
			ID:      "svc2",
			Product: "openssh",
			Version: "9.0p1",
		},
		{
			ID:      "svc3",
			Product: "",
			Version: "",
			Banner:  "", // No data, should be skipped
		},
	}

	result := GenerateCPEBatch(services)

	// Should have 2 entries (svc3 has no data)
	if len(result) != 2 {
		t.Errorf("GenerateCPEBatch() returned %d entries, want 2", len(result))
	}

	// Check svc1
	if cpes, exists := result["svc1"]; !exists {
		t.Error("GenerateCPEBatch() missing svc1")
	} else if len(cpes) == 0 {
		t.Error("GenerateCPEBatch() svc1 has no CPEs")
	}

	// Check svc2
	if cpes, exists := result["svc2"]; !exists {
		t.Error("GenerateCPEBatch() missing svc2")
	} else if len(cpes) == 0 {
		t.Error("GenerateCPEBatch() svc2 has no CPEs")
	}

	// Check svc3 is not present
	if _, exists := result["svc3"]; exists {
		t.Error("GenerateCPEBatch() should not include svc3 (no data)")
	}
}
