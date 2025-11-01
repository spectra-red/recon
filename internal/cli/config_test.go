package cli

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitConfig_Defaults(t *testing.T) {
	// Reset viper between tests
	viper.Reset()

	cfg, err := InitConfig("")
	require.NoError(t, err)
	assert.NotNil(t, cfg)

	// Check defaults
	assert.Equal(t, "http://localhost:3000", cfg.API.URL)
	assert.Equal(t, 30*time.Second, cfg.API.Timeout)
	assert.Equal(t, "json", cfg.Output.Format)
	assert.True(t, cfg.Output.Color)
	assert.Empty(t, cfg.Scanner.PublicKey)
	assert.Empty(t, cfg.Scanner.PrivateKey)
}

func TestInitConfig_FromFile(t *testing.T) {
	// Reset viper between tests
	viper.Reset()

	// Create a temporary config file
	tmpDir := t.TempDir()
	cfgFile := filepath.Join(tmpDir, ".spectra.yaml")

	configContent := `
api:
  url: http://api.example.com:8080
  timeout: 60s

scanner:
  public_key: test-public-key
  private_key: test-private-key

output:
  format: yaml
  color: false
`

	err := os.WriteFile(cfgFile, []byte(configContent), 0644)
	require.NoError(t, err)

	cfg, err := InitConfig(cfgFile)
	require.NoError(t, err)
	assert.NotNil(t, cfg)

	// Check values from file
	assert.Equal(t, "http://api.example.com:8080", cfg.API.URL)
	assert.Equal(t, 60*time.Second, cfg.API.Timeout)
	assert.Equal(t, "yaml", cfg.Output.Format)
	assert.False(t, cfg.Output.Color)
	assert.Equal(t, "test-public-key", cfg.Scanner.PublicKey)
	assert.Equal(t, "test-private-key", cfg.Scanner.PrivateKey)
}

func TestInitConfig_EnvVarsOverride(t *testing.T) {
	// Reset viper between tests
	viper.Reset()

	// Set environment variables
	os.Setenv("SPECTRA_API_URL", "http://env.example.com")
	os.Setenv("SPECTRA_OUTPUT_FORMAT", "table")
	defer func() {
		os.Unsetenv("SPECTRA_API_URL")
		os.Unsetenv("SPECTRA_OUTPUT_FORMAT")
	}()

	cfg, err := InitConfig("")
	require.NoError(t, err)
	assert.NotNil(t, cfg)

	// Environment variables should override defaults
	assert.Equal(t, "http://env.example.com", cfg.API.URL)
	assert.Equal(t, "table", cfg.Output.Format)
}

func TestValidateConfig_Valid(t *testing.T) {
	cfg := &Config{
		API: APIConfig{
			URL:     "http://localhost:3000",
			Timeout: 30 * time.Second,
		},
		Output: OutputConfig{
			Format: "json",
			Color:  true,
		},
	}

	err := ValidateConfig(cfg)
	assert.NoError(t, err)
}

func TestValidateConfig_InvalidURL(t *testing.T) {
	cfg := &Config{
		API: APIConfig{
			URL:     "",
			Timeout: 30 * time.Second,
		},
		Output: OutputConfig{
			Format: "json",
		},
	}

	err := ValidateConfig(cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "api.url cannot be empty")
}

func TestValidateConfig_InvalidTimeout(t *testing.T) {
	cfg := &Config{
		API: APIConfig{
			URL:     "http://localhost:3000",
			Timeout: -1 * time.Second,
		},
		Output: OutputConfig{
			Format: "json",
		},
	}

	err := ValidateConfig(cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "api.timeout must be positive")
}

func TestValidateConfig_InvalidOutputFormat(t *testing.T) {
	cfg := &Config{
		API: APIConfig{
			URL:     "http://localhost:3000",
			Timeout: 30 * time.Second,
		},
		Output: OutputConfig{
			Format: "invalid",
		},
	}

	err := ValidateConfig(cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid output format")
}

func TestGetterFunctions(t *testing.T) {
	// Reset viper between tests
	viper.Reset()

	// Set some values
	viper.Set("api.url", "http://test.example.com")
	viper.Set("api.timeout", "45s")
	viper.Set("output.format", "yaml")
	viper.Set("output.color", false)
	viper.Set("scanner.public_key", "pub-key")
	viper.Set("scanner.private_key", "priv-key")

	// Test getters
	assert.Equal(t, "http://test.example.com", GetAPIURL())
	assert.Equal(t, 45*time.Second, GetAPITimeout())
	assert.Equal(t, "yaml", GetOutputFormat())
	assert.False(t, GetOutputColor())
	assert.Equal(t, "pub-key", GetScannerPublicKey())
	assert.Equal(t, "priv-key", GetScannerPrivateKey())
}

func TestConfigPrecedence(t *testing.T) {
	// This test verifies the precedence: env vars > config file > defaults
	viper.Reset()

	// Create a temporary config file
	tmpDir := t.TempDir()
	cfgFile := filepath.Join(tmpDir, ".spectra.yaml")

	configContent := `
api:
  url: http://file.example.com
  timeout: 60s
`

	err := os.WriteFile(cfgFile, []byte(configContent), 0644)
	require.NoError(t, err)

	// Set environment variable (should override file)
	os.Setenv("SPECTRA_API_URL", "http://env.example.com")
	defer os.Unsetenv("SPECTRA_API_URL")

	cfg, err := InitConfig(cfgFile)
	require.NoError(t, err)

	// Environment variable should win
	assert.Equal(t, "http://env.example.com", cfg.API.URL)
	// File value should be used where no env var is set
	assert.Equal(t, 60*time.Second, cfg.API.Timeout)
}
