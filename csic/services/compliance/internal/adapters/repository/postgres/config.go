package postgres

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/viper"
)

// Config holds database configuration
type Config struct {
	Host            string
	Port            int
	Username        string
	Password        string
	Name            string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime int
}

// LoadConfig loads database configuration from environment or config file
func LoadConfig() (*Config, error) {
	// Try to get from environment first
	host := os.Getenv("DB_HOST")
	if host == "" {
		host = viper.GetString("database.host")
	}
	if host == "" {
		host = "localhost"
	}

	port := 5432
	if p := os.Getenv("DB_PORT"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil {
			port = parsed
		}
	} else if viper.IsSet("database.port") {
		port = viper.GetInt("database.port")
	}

	username := os.Getenv("DB_USER")
	if username == "" {
		username = viper.GetString("database.username")
	}
	if username == "" {
		username = "postgres"
	}

	password := os.Getenv("DB_PASSWORD")
	if password == "" {
		password = viper.GetString("database.password")
	}
	if password == "" {
		password = "postgres"
	}

	name := os.Getenv("DB_NAME")
	if name == "" {
		name = viper.GetString("database.name")
	}
	if name == "" {
		name = "csic_platform"
	}

	sslmode := "disable"
	if s := os.Getenv("DB_SSLMODE"); s != "" {
		sslmode = s
	} else if viper.IsSet("database.sslmode") {
		sslmode = viper.GetString("database.sslmode")
	}

	maxOpenConns := 25
	if viper.IsSet("database.max_open_conns") {
		maxOpenConns = viper.GetInt("database.max_open_conns")
	}

	maxIdleConns := 5
	if viper.IsSet("database.max_idle_conns") {
		maxIdleConns = viper.GetInt("database.max_idle_conns")
	}

	connMaxLifetime := 300
	if viper.IsSet("database.conn_max_lifetime") {
		connMaxLifetime = viper.GetInt("database.conn_max_lifetime")
	}

	return &Config{
		Host:            host,
		Port:            port,
		Username:        username,
		Password:        password,
		Name:            name,
		SSLMode:         sslmode,
		MaxOpenConns:    maxOpenConns,
		MaxIdleConns:    maxIdleConns,
		ConnMaxLifetime: connMaxLifetime,
	}, nil
}

// ConnectionString returns the PostgreSQL connection string
func (c *Config) ConnectionString() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.Username,
		c.Password,
		c.Host,
		c.Port,
		c.Name,
		c.SSLMode,
	)
}
