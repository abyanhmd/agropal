package config

import (
	"os"
	"path/filepath"
	"testing"
)

const testConfigYAML = `
stage: "dev"
server:
  host: "0.0.0.0"
  port: 8080
path:
  migration: "internal/migration"
  log: "logs"
`

func writeTestConfigFile(t *testing.T, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write test config file: %v", err)
	}

	return path
}

func setEnvBindings(t *testing.T) {
	t.Helper()

	t.Setenv("ENV", "test")

	t.Setenv("POSTGRES_HOST", "db-host")
	t.Setenv("POSTGRES_PORT", "5432")
	t.Setenv("POSTGRES_USER", "db-user")
	t.Setenv("POSTGRES_PASSWORD", "db-password")
	t.Setenv("POSTGRES_NAME", "db-name")

	t.Setenv("REDIS_ADDRESS", "redis:6379")
	t.Setenv("REDIS_PASSWORD", "redis-password")
	t.Setenv("REDIS_DB", "1")
	t.Setenv("REDIS_PROTOCOL", "3")

	t.Setenv("RABBITMQ_ADDRESS", "rabbitmq:5672")
	t.Setenv("RABBITMQ_URL", "amqp://rabbitmq:5672")
	t.Setenv("RABBITMQ_USERNAME", "mq-user")
	t.Setenv("RABBITMQ_PASSWORD", "mq-password")

	t.Setenv("SMTP_HOST", "smtp.example.com")
	t.Setenv("SMTP_PORT", "587")
	t.Setenv("SMTP_USERNAME", "smtp-user")
	t.Setenv("SMTP_FROM", "noreply@example.com")

	t.Setenv("SECRET_JWT", "jwt-secret")
	t.Setenv("SECRET_GCM_TOKEN", "gcm-secret")
	t.Setenv("SECRET_SMTP", "smtp-secret")
}

func TestLoadConfig(t *testing.T) {
	t.Run("Loads Values From Config File And Environment", func(t *testing.T) {
		setEnvBindings(t)
		path := writeTestConfigFile(t, testConfigYAML)

		cfg := LoadConfig(path)

		if cfg.Stage != "dev" {
			t.Errorf("Stage = %q, want %q", cfg.Stage, "dev")
		}
		if cfg.Server.Host != "0.0.0.0" {
			t.Errorf("Server.Host = %q, want %q", cfg.Server.Host, "0.0.0.0")
		}
		if cfg.Server.Port != 8080 {
			t.Errorf("Server.Port = %d, want %d", cfg.Server.Port, 8080)
		}
		if cfg.Path.Migration != "internal/migration" {
			t.Errorf("Path.Migration = %q, want %q", cfg.Path.Migration, "internal/migration")
		}
		if cfg.Path.Log != "logs" {
			t.Errorf("Path.Log = %q, want %q", cfg.Path.Log, "logs")
		}

		if cfg.Database.Host != "db-host" {
			t.Errorf("Database.Host = %q, want %q", cfg.Database.Host, "db-host")
		}
		if cfg.Database.Port != "5432" {
			t.Errorf("Database.Port = %q, want %q", cfg.Database.Port, "5432")
		}
		if cfg.Database.User != "db-user" {
			t.Errorf("Database.User = %q, want %q", cfg.Database.User, "db-user")
		}
		if cfg.Database.Password != "db-password" {
			t.Errorf("Database.Password = %q, want %q", cfg.Database.Password, "db-password")
		}
		if cfg.Database.Dbname != "db-name" {
			t.Errorf("Database.Dbname = %q, want %q", cfg.Database.Dbname, "db-name")
		}

		if cfg.Redis.Address != "redis:6379" {
			t.Errorf("Redis.Address = %q, want %q", cfg.Redis.Address, "redis:6379")
		}
		if cfg.Redis.Password != "redis-password" {
			t.Errorf("Redis.Password = %q, want %q", cfg.Redis.Password, "redis-password")
		}
		if cfg.Redis.Db != 1 {
			t.Errorf("Redis.Db = %d, want %d", cfg.Redis.Db, 1)
		}
		if cfg.Redis.Protocol != 3 {
			t.Errorf("Redis.Protocol = %d, want %d", cfg.Redis.Protocol, 3)
		}

		if cfg.RabbitMQ.Address != "rabbitmq:5672" {
			t.Errorf("RabbitMQ.Address = %q, want %q", cfg.RabbitMQ.Address, "rabbitmq:5672")
		}
		if cfg.RabbitMQ.URL != "amqp://rabbitmq:5672" {
			t.Errorf("RabbitMQ.URL = %q, want %q", cfg.RabbitMQ.URL, "amqp://rabbitmq:5672")
		}
		if cfg.RabbitMQ.Username != "mq-user" {
			t.Errorf("RabbitMQ.Username = %q, want %q", cfg.RabbitMQ.Username, "mq-user")
		}
		if cfg.RabbitMQ.Password != "mq-password" {
			t.Errorf("RabbitMQ.Password = %q, want %q", cfg.RabbitMQ.Password, "mq-password")
		}

		if cfg.Smtp.Host != "smtp.example.com" {
			t.Errorf("Smtp.Host = %q, want %q", cfg.Smtp.Host, "smtp.example.com")
		}
		if cfg.Smtp.Port != "587" {
			t.Errorf("Smtp.Port = %q, want %q", cfg.Smtp.Port, "587")
		}
		if cfg.Smtp.Username != "smtp-user" {
			t.Errorf("Smtp.Username = %q, want %q", cfg.Smtp.Username, "smtp-user")
		}
		if cfg.Smtp.From != "noreply@example.com" {
			t.Errorf("Smtp.From = %q, want %q", cfg.Smtp.From, "noreply@example.com")
		}

		if cfg.Secret.Jwt != "jwt-secret" {
			t.Errorf("Secret.Jwt = %q, want %q", cfg.Secret.Jwt, "jwt-secret")
		}
		if cfg.Secret.GcmToken != "gcm-secret" {
			t.Errorf("Secret.GcmToken = %q, want %q", cfg.Secret.GcmToken, "gcm-secret")
		}
		if cfg.Secret.Smtp != "smtp-secret" {
			t.Errorf("Secret.Smtp = %q, want %q", cfg.Secret.Smtp, "smtp-secret")
		}
	})

	t.Run("Missing Env Bindings Leave Zero Values", func(t *testing.T) {
		t.Setenv("ENV", "test")
		path := writeTestConfigFile(t, testConfigYAML)

		cfg := LoadConfig(path)

		if cfg.Database.Host != "" {
			t.Errorf("Database.Host = %q, want empty string when POSTGRES_HOST is unset", cfg.Database.Host)
		}
		if cfg.Secret.Jwt != "" {
			t.Errorf("Secret.Jwt = %q, want empty string when SECRET_JWT is unset", cfg.Secret.Jwt)
		}
	})

	t.Run("Panics When Config File Does Not Exist", func(t *testing.T) {
		t.Setenv("ENV", "test")

		defer func() {
			if r := recover(); r == nil {
				t.Error("LoadConfig() did not panic, want panic for missing config file")
			}
		}()

		LoadConfig(filepath.Join(t.TempDir(), "does-not-exist.yaml"))
	})

	t.Run("Panics When Config File Has Invalid Syntax", func(t *testing.T) {
		t.Setenv("ENV", "test")
		path := writeTestConfigFile(t, "stage: [this is not: valid yaml")

		defer func() {
			if r := recover(); r == nil {
				t.Error("LoadConfig() did not panic, want panic for malformed config file")
			}
		}()

		LoadConfig(path)
	})
}
