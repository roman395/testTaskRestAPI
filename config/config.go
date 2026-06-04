package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

// Config holds all configuration for the application
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Logging  LoggingConfig  `yaml:"logging"`
	CORS     CORSConfig     `yaml:"cors"`
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Host         string
	Port         int
	ReadTimeout  int `yaml:"read_timeout"`
	WriteTimeout int `yaml:"write_timeout"`
	IdleTimeout  int `yaml:"idle_timeout"`
}

// DatabaseConfig holds database connection configuration
type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
	SSLMode  string `yaml:"sslmode"`

	MaxOpenConns           int `yaml:"max_open_conns"`
	MaxIdleConns           int `yaml:"max_idle_conns"`
	ConnMaxLifetimeMinutes int `yaml:"conn_max_lifetime_minutes"`
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Format             string `yaml:"format"`
	IncludeRequestBody bool   `yaml:"include_request_body"`
}

// CORSConfig holds CORS configuration
type CORSConfig struct {
	AllowedOrigins []string `yaml:"allowed_origins"`
	AllowedMethods []string `yaml:"allowed_methods"`
}

// Load loads configuration from .env and config.yaml files
func Load() (*Config, error) {
	log.Println("[CONFIG] Loading configuration...")

	// Load .env file (optional, don't fail if not found)
	if err := godotenv.Load(); err != nil {
		log.Println("[CONFIG] No .env file found, using environment variables")
	} else {
		log.Println("[CONFIG] .env file loaded successfully")
	}

	// Load YAML configuration
	config := &Config{}
	if err := loadYAMLConfig(config); err != nil {
		log.Printf("[CONFIG] Warning: Failed to load config.yaml: %v", err)
		log.Println("[CONFIG] Using default YAML values")
	} else {
		log.Println("[CONFIG] config.yaml loaded successfully")
	}

	// Load environment variables (override YAML)
	loadEnvConfig(config)

	// Validate configuration
	if err := validateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	log.Println("[CONFIG] Configuration loaded successfully")
	return config, nil
}

// loadYAMLConfig loads configuration from config.yaml file
func loadYAMLConfig(config *Config) error {
	yamlFile, err := os.ReadFile("config.yaml")
	if err != nil {
		return err
	}

	// Set defaults first
	setDefaults(config)

	// Unmarshal YAML
	if err := yaml.Unmarshal(yamlFile, config); err != nil {
		return err
	}

	return nil
}

// loadEnvConfig loads configuration from environment variables
func loadEnvConfig(config *Config) {
	// Server configuration from env
	if host := os.Getenv("SERVER_HOST"); host != "" {
		config.Server.Host = host
	}
	if port := os.Getenv("SERVER_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			config.Server.Port = p
		}
	}

	// Database configuration from env
	if host := os.Getenv("DB_HOST"); host != "" {
		config.Database.Host = host
	}
	if port := os.Getenv("DB_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			config.Database.Port = p
		}
	}
	if user := os.Getenv("DB_USER"); user != "" {
		config.Database.User = user
	}
	if password := os.Getenv("DB_PASSWORD"); password != "" {
		config.Database.Password = password
	}
	if name := os.Getenv("DB_NAME"); name != "" {
		config.Database.Name = name
	}
	if sslmode := os.Getenv("DB_SSLMODE"); sslmode != "" {
		config.Database.SSLMode = sslmode
	}

	// Logging configuration from env
	if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
		// Can be used for log level if needed
	}
}

// setDefaults sets default configuration values
func setDefaults(config *Config) {
	// Server defaults
	config.Server.Host = "0.0.0.0"
	config.Server.Port = 8080
	config.Server.ReadTimeout = 15
	config.Server.WriteTimeout = 15
	config.Server.IdleTimeout = 60

	// Database defaults
	config.Database.Host = "localhost"
	config.Database.Port = 5432
	config.Database.User = "subscriptionuser"
	config.Database.Password = "subscriptionpass"
	config.Database.Name = "subscriptiondb"
	config.Database.SSLMode = "disable"
	config.Database.MaxOpenConns = 25
	config.Database.MaxIdleConns = 5
	config.Database.ConnMaxLifetimeMinutes = 5

	// Logging defaults
	config.Logging.Format = "text"
	config.Logging.IncludeRequestBody = false

	// CORS defaults
	config.CORS.AllowedOrigins = []string{"*"}
	config.CORS.AllowedMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
}

// validateConfig validates the configuration
func validateConfig(config *Config) error {
	// Validate server config
	if config.Server.Port < 1 || config.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", config.Server.Port)
	}

	// Validate database config
	if config.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}
	if config.Database.Port < 1 || config.Database.Port > 65535 {
		return fmt.Errorf("invalid database port: %d", config.Database.Port)
	}
	if config.Database.User == "" {
		return fmt.Errorf("database user is required")
	}
	if config.Database.Name == "" {
		return fmt.Errorf("database name is required")
	}

	return nil
}

// DatabaseURL returns the PostgreSQL connection string
func (c *Config) DatabaseURL() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.Database.User,
		c.Database.Password,
		c.Database.Host,
		c.Database.Port,
		c.Database.Name,
		c.Database.SSLMode,
	)
}

// ServerAddress returns the server address string
func (c *Config) ServerAddress() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}

// ReadTimeoutDuration returns read timeout as time.Duration
func (c *Config) ReadTimeoutDuration() time.Duration {
	return time.Duration(c.Server.ReadTimeout) * time.Second
}

// WriteTimeoutDuration returns write timeout as time.Duration
func (c *Config) WriteTimeoutDuration() time.Duration {
	return time.Duration(c.Server.WriteTimeout) * time.Second
}

// IdleTimeoutDuration returns idle timeout as time.Duration
func (c *Config) IdleTimeoutDuration() time.Duration {
	return time.Duration(c.Server.IdleTimeout) * time.Second
}

// ConnMaxLifetime returns connection max lifetime as time.Duration
func (c *Config) ConnMaxLifetime() time.Duration {
	return time.Duration(c.Database.ConnMaxLifetimeMinutes) * time.Minute
}
