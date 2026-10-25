package database

import (
	"agropal/config"
	"testing"
)

func TestNewPostgres(t *testing.T) {
	t.Run("Returns A Working Connection For A Valid Config", func(t *testing.T) {
		if testConfig == nil {
			t.Skip("no test database available (see main_test.go)")
		}

		db := NewPostgres(testConfig)
		defer db.Close()

		if err := db.Ping(); err != nil {
			t.Errorf("Ping() error = %v, want nil", err)
		}
	})

	t.Run("Panics On Non Numeric Port", func(t *testing.T) {
		cfg := &config.Config{}
		cfg.Database.Port = "not-a-number"

		defer func() {
			if r := recover(); r == nil {
				t.Error("expected NewPostgres to panic on a non-numeric port, it did not")
			}
		}()

		NewPostgres(cfg)
	})

	t.Run("Panics When Connection Fails", func(t *testing.T) {
		cfg := &config.Config{}
		// Port 0 isn't a listening service on any host, so the connect
		// attempt fails fast (connection refused) without needing Docker.
		cfg.Database.Host = "127.0.0.1"
		cfg.Database.Port = "0"
		cfg.Database.User = "test"
		cfg.Database.Password = "test"
		cfg.Database.Dbname = "test"

		defer func() {
			if r := recover(); r == nil {
				t.Error("expected NewPostgres to panic when the connection fails, it did not")
			}
		}()

		NewPostgres(cfg)
	})
}
