package config

import (
	"errors"
	"fmt"
	"iscrie/utils"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Default values
const (
	DefaultBatchSize     = 1   // Default batch Size
	MaxBatchSize         = 100 // Max allowed batch_size
	DefaultRetryTimeout  = 10  // Timeout by default per second
	DefaultRetryAttempts = 3   // number of retries attempt
)

// AuthConfig defines the authentication configuration
type AuthConfig struct {
	Type        string `mapstructure:"type"`
	UserToken   string `mapstructure:"user_token"`
	PassToken   string `mapstructure:"pass_token"`
	AccessToken string `mapstructure:"access_token"`
	HeaderName  string `mapstructure:"header_name"`
	HeaderValue string `mapstructure:"header_value"`
}

// ProxyConfig defines proxy configuration
type ProxyConfig struct {
	Enabled  bool   `mapstructure:"enabled"`
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}

// Config represents the application's configuration
type Config struct {
	General struct {
		RootPath  string `mapstructure:"root_path"`
		LogPath   string `mapstructure:"log_path"`
		LogLevel  string `mapstructure:"log_level"`
		BatchSize int    `mapstructure:"batch_size"`
	} `mapstructure:"general"`
	Nexus struct {
		URL            string `mapstructure:"url"`
		Repository     string `mapstructure:"repository"`
		RepositoryType string `mapstructure:"repository_type"`
		ForceReplace   bool   `mapstructure:"force_replace"`
	} `mapstructure:"nexus"`
	Retry RetryConfig `mapstructure:"retry"`
	Proxy ProxyConfig `mapstructure:"proxy"`
	Auth  AuthConfig  `mapstructure:"auth"`
}

type RetryConfig struct {
	RetryAttempts int `mapstructure:"retry_attempts"`
	Timeout       int `mapstructure:"timeout"`
}

// SupportedRepositoryTypes defines all repository types currently supported by Iscrie.
var SupportedRepositoryTypes = []string{"maven2", "raw"}

// IsValidRepositoryType checks if the given repository type is supported.
func IsValidRepositoryType(repoType string) bool {
	for _, supportedType := range SupportedRepositoryTypes {
		if repoType == supportedType {
			return true
		}
	}
	return false
}

// LoadConfig charges TOML configuration from a file
func LoadConfig(configPath string) (*Config, error) {
	setDefaults()

	if configPath == "" {
		defaultPath, _ := filepath.Abs("./iscrie.toml")
		configPath = defaultPath
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, utils.LogAndReturnError("configuration file not found at path: %s", configPath)
	}

	viper.SetConfigFile(configPath)
	viper.SetConfigType("toml")

	if err := viper.ReadInConfig(); err != nil {
		return nil, utils.LogAndReturnError("failed to read configuration file: %w", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, utils.LogAndReturnError("failed to parse configuration: %w", err)
	}

	if err := validateConfig(&cfg); err != nil {
		return nil, err
	}

	paths, err := utils.ConvertPathsToAbsolute(cfg.General.RootPath, cfg.General.LogPath)
	if err != nil {
		fmt.Printf("Error converting paths to absolute: %v\n", err)
		return nil, err
	}
	cfg.General.RootPath = paths[0]
	cfg.General.LogPath = paths[1]

	return &cfg, nil
}

// setDefaults centralize default values
func setDefaults() {
	viper.SetDefault("general.log_level", "info")
	viper.SetDefault("general.batch_size", DefaultBatchSize)
	viper.SetDefault("retry.retry_attempts", DefaultRetryAttempts)
	viper.SetDefault("retry.timeout", DefaultRetryTimeout)
	viper.SetDefault("proxy.enabled", false)
	viper.SetDefault("nexus.repository_type", "raw")
	viper.SetDefault("nexus.force_replace", false)
	fmt.Println("Default configuration values applied.")
}

// validateConfig validates configuration to ensure coherence
func validateConfig(cfg *Config) error {
	if cfg.General.RootPath == "" {
		return errors.New("missing required field: general.root_path")
	}
	if cfg.General.BatchSize <= 0 || cfg.General.BatchSize > MaxBatchSize {
		return utils.LogAndReturnError("general.batch_size must be between 1 and %d", MaxBatchSize)
	}
	if cfg.Nexus.URL == "" {
		return errors.New("missing required field: nexus.url")
	}
	if cfg.Nexus.Repository == "" {
		return errors.New("missing required field: nexus.repository")
	}

	switch cfg.Nexus.RepositoryType {
	case "raw", "maven2":
		// Valid types
	default:
		return utils.LogAndReturnError("invalid nexus.repository_type: %s. Valid options are 'raw' or 'maven2'", cfg.Nexus.RepositoryType)
	}

	if cfg.Retry.RetryAttempts < 0 {
		return errors.New("retry.retry_attempts cannot be negative")
	}
	if cfg.Retry.Timeout <= 0 {
		return errors.New("retry.timeout must be greater than zero")
	}

	if cfg.Proxy.Enabled {
		if cfg.Proxy.Host == "" {
			return errors.New("proxy.host is required if proxy is enabled")
		}
		if cfg.Proxy.Port <= 0 {
			return errors.New("proxy.port must be greater than zero")
		}
	}

	return validateAuthConfig(&cfg.Auth)
}

func validateAuthConfig(auth *AuthConfig) error {
	switch auth.Type {
	case "basic":
		if auth.UserToken == "" || auth.PassToken == "" {
			return errors.New("auth.type 'basic' requires both user_token and pass_token")
		}
	case "bearer":
		if auth.AccessToken == "" {
			return errors.New("auth.type 'bearer' requires access_token")
		}
	case "header":
		if auth.HeaderName == "" || auth.HeaderValue == "" {
			return errors.New("auth.type 'header' requires both header_name and header_value")
		}
	default:
		return utils.LogAndReturnError("invalid auth.type: %s. Valid options are 'basic', 'bearer', 'header'", auth.Type)
	}

	if count := utils.CountNonEmpty(auth.UserToken, auth.PassToken, auth.AccessToken, auth.HeaderName, auth.HeaderValue); count > 2 {
		return errors.New("only one authentication method should be configured")
	}

	return nil
}
