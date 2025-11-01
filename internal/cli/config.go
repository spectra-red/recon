package cli

import (
	"crypto/ed25519"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/viper"
)

// Config holds all configuration for the CLI
type Config struct {
	API      APIConfig      `mapstructure:"api"`
	Scanner  ScannerConfig  `mapstructure:"scanner"`
	Output   OutputConfig   `mapstructure:"output"`
}

// APIConfig holds API-related configuration
type APIConfig struct {
	URL     string        `mapstructure:"url"`
	Timeout time.Duration `mapstructure:"timeout"`
}

// ScannerConfig holds scanner authentication configuration
type ScannerConfig struct {
	PublicKey  string `mapstructure:"public_key"`
	PrivateKey string `mapstructure:"private_key"`
}

// OutputConfig holds output formatting configuration
type OutputConfig struct {
	Format string `mapstructure:"format"`
	Color  bool   `mapstructure:"color"`
}

// InitConfig initializes configuration from file, environment variables, and flags
// Configuration precedence: flags > env vars > config file > defaults
func InitConfig(cfgFile string) (*Config, error) {
	// Set defaults
	setDefaults()

	// If a config file is specified, use it
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		// Search for config in standard locations
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("unable to find home directory: %w", err)
		}

		// Add config search paths (in order of precedence)
		viper.AddConfigPath(".")                            // Current directory
		viper.AddConfigPath(filepath.Join(home, ".spectra")) // ~/.spectra/
		viper.AddConfigPath("/etc/spectra")                 // /etc/spectra/

		viper.SetConfigName(".spectra")
		viper.SetConfigType("yaml")
	}

	// Read environment variables
	viper.SetEnvPrefix("SPECTRA")
	viper.AutomaticEnv()

	// Bind environment variables to config keys explicitly
	viper.BindEnv("api.url", "SPECTRA_API_URL")
	viper.BindEnv("api.timeout", "SPECTRA_API_TIMEOUT")
	viper.BindEnv("output.format", "SPECTRA_OUTPUT_FORMAT")
	viper.BindEnv("output.color", "SPECTRA_OUTPUT_COLOR")
	viper.BindEnv("scanner.public_key", "SPECTRA_SCANNER_PUBLIC_KEY")
	viper.BindEnv("scanner.private_key", "SPECTRA_SCANNER_PRIVATE_KEY")

	// Read config file if it exists
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			// Config file was found but another error was produced
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
		// Config file not found; ignore error and use defaults
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("unable to decode config: %w", err)
	}

	return &config, nil
}

// setDefaults sets default values for all configuration options
func setDefaults() {
	// API defaults
	viper.SetDefault("api.url", "http://localhost:3000")
	viper.SetDefault("api.timeout", "30s")

	// Scanner defaults
	viper.SetDefault("scanner.public_key", "")
	viper.SetDefault("scanner.private_key", "")

	// Output defaults
	viper.SetDefault("output.format", "json")
	viper.SetDefault("output.color", true)
}

// GetAPIURL returns the configured API URL
func GetAPIURL() string {
	return viper.GetString("api.url")
}

// GetAPITimeout returns the configured API timeout
func GetAPITimeout() time.Duration {
	return viper.GetDuration("api.timeout")
}

// GetOutputFormat returns the configured output format
func GetOutputFormat() string {
	return viper.GetString("output.format")
}

// GetOutputColor returns whether color output is enabled
func GetOutputColor() bool {
	return viper.GetBool("output.color")
}

// GetScannerPublicKey returns the scanner's public key
func GetScannerPublicKey() string {
	return viper.GetString("scanner.public_key")
}

// GetScannerPrivateKey returns the scanner's private key
func GetScannerPrivateKey() string {
	return viper.GetString("scanner.private_key")
}

// ValidateConfig validates the configuration
func ValidateConfig(cfg *Config) error {
	// Validate API URL
	if cfg.API.URL == "" {
		return fmt.Errorf("api.url cannot be empty")
	}

	// Validate API timeout
	if cfg.API.Timeout <= 0 {
		return fmt.Errorf("api.timeout must be positive")
	}

	// Validate output format
	validFormats := map[string]bool{
		"json":  true,
		"yaml":  true,
		"table": true,
	}
	if !validFormats[cfg.Output.Format] {
		return fmt.Errorf("invalid output format: %s (must be json, yaml, or table)", cfg.Output.Format)
	}

	return nil
}

// GetPrivateKeyBytes decodes and returns the Ed25519 private key bytes
func GetPrivateKeyBytes() ([]byte, error) {
	privKeyStr := GetScannerPrivateKey()
	if privKeyStr == "" {
		return nil, fmt.Errorf("no private key configured (run 'spectra keys generate')")
	}

	// Import encoding/base64 at the top of the file
	return decodeBase64Key(privKeyStr, "private key")
}

// GetPublicKeyBytes decodes and returns the Ed25519 public key bytes
func GetPublicKeyBytes() ([]byte, error) {
	pubKeyStr := GetScannerPublicKey()
	if pubKeyStr == "" {
		return nil, fmt.Errorf("no public key configured")
	}

	return decodeBase64Key(pubKeyStr, "public key")
}

// decodeBase64Key decodes a base64-encoded key
func decodeBase64Key(encoded, keyType string) ([]byte, error) {
	keyBytes, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("failed to decode %s: %w", keyType, err)
	}
	return keyBytes, nil
}

// GetPrivateKey returns the Ed25519 private key
func GetPrivateKey() (ed25519.PrivateKey, error) {
	keyBytes, err := GetPrivateKeyBytes()
	if err != nil {
		return nil, err
	}

	if len(keyBytes) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("invalid private key size: expected %d, got %d",
			ed25519.PrivateKeySize, len(keyBytes))
	}

	return ed25519.PrivateKey(keyBytes), nil
}

// GetPublicKey returns the Ed25519 public key
func GetPublicKey() (ed25519.PublicKey, error) {
	keyBytes, err := GetPublicKeyBytes()
	if err != nil {
		return nil, err
	}

	if len(keyBytes) != ed25519.PublicKeySize {
		return nil, fmt.Errorf("invalid public key size: expected %d, got %d",
			ed25519.PublicKeySize, len(keyBytes))
	}

	return ed25519.PublicKey(keyBytes), nil
}
