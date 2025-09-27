package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLoad(t *testing.T) {
	t.Run("Success with environment variables", func(t *testing.T) {
		os.Setenv("SERVER_PORT", "9090")
		os.Setenv("DB_NAME", "test_db")
		os.Setenv("STOCK_API_KEY", "test_key")
		defer func() {
			os.Unsetenv("SERVER_PORT")
			os.Unsetenv("DB_NAME")
			os.Unsetenv("STOCK_API_KEY")
		}()

		cfg, err := Load()

		assert.NoError(t, err)
		assert.NotNil(t, cfg)
		assert.Equal(t, "9090", cfg.Server.Port)
		assert.Equal(t, "test_db", cfg.Database.Name)
		assert.Equal(t, "test_key", cfg.StockAPI.APIKey)
	})

	t.Run("Validation error - missing API key", func(t *testing.T) {
		os.Unsetenv("STOCK_API_KEY")

		cfg, err := Load()

		assert.Error(t, err)
		assert.Nil(t, cfg)
		assert.Contains(t, err.Error(), "STOCK_API_KEY is required")
	})

	t.Run("Default values", func(t *testing.T) {
		os.Setenv("STOCK_API_KEY", "test_key")
		os.Setenv("DB_NAME", "test_db")
		defer func() {
			os.Unsetenv("STOCK_API_KEY")
			os.Unsetenv("DB_NAME")
		}()

		cfg, err := Load()

		assert.NoError(t, err)
		assert.Equal(t, "8080", cfg.Server.Port)
		assert.Equal(t, "0.0.0.0", cfg.Server.Host)
		assert.Equal(t, "development", cfg.Server.Env)
		assert.Equal(t, 25, cfg.Database.MaxConns)
		assert.Equal(t, 5, cfg.Database.MinConns)
	})
}

func TestDatabaseConfig_GetDSN(t *testing.T) {
	cfg := &DatabaseConfig{
		Host:     "localhost",
		Port:     "26257",
		User:     "root",
		Password: "password",
		Name:     "test_db",
		SSLMode:  "disable",
	}

	dsn := cfg.GetDSN()

	assert.Contains(t, dsn, "host=localhost")
	assert.Contains(t, dsn, "port=26257")
	assert.Contains(t, dsn, "user=root")
	assert.Contains(t, dsn, "password=password")
	assert.Contains(t, dsn, "dbname=test_db")
	assert.Contains(t, dsn, "sslmode=disable")
}

func TestGetEnvAsDuration(t *testing.T) {
	t.Run("Valid duration", func(t *testing.T) {
		os.Setenv("TEST_DURATION", "10s")
		defer os.Unsetenv("TEST_DURATION")

		duration := getEnvAsDuration("TEST_DURATION", 5*time.Second)

		assert.Equal(t, 10*time.Second, duration)
	})

	t.Run("Invalid duration - use default", func(t *testing.T) {
		os.Setenv("TEST_DURATION", "invalid")
		defer os.Unsetenv("TEST_DURATION")

		duration := getEnvAsDuration("TEST_DURATION", 5*time.Second)

		assert.Equal(t, 5*time.Second, duration)
	})

	t.Run("Missing env - use default", func(t *testing.T) {
		duration := getEnvAsDuration("MISSING_DURATION", 5*time.Second)

		assert.Equal(t, 5*time.Second, duration)
	})
}

func TestGetEnvAsInt(t *testing.T) {
	t.Run("Valid integer", func(t *testing.T) {
		os.Setenv("TEST_INT", "42")
		defer os.Unsetenv("TEST_INT")

		value := getEnvAsInt("TEST_INT", 10)

		assert.Equal(t, 42, value)
	})

	t.Run("Invalid integer - use default", func(t *testing.T) {
		os.Setenv("TEST_INT", "not_a_number")
		defer os.Unsetenv("TEST_INT")

		value := getEnvAsInt("TEST_INT", 10)

		assert.Equal(t, 10, value)
	})

	t.Run("Missing env - use default", func(t *testing.T) {
		value := getEnvAsInt("MISSING_INT", 10)

		assert.Equal(t, 10, value)
	})
}
