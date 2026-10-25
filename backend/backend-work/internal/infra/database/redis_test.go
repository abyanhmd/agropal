package database

import (
	"agropal/config"
	"testing"

	"github.com/alicebob/miniredis/v2"
)

func TestNewRedis(t *testing.T) {
	t.Run("Connects Successfully And Pings The Server", func(t *testing.T) {
		mr := miniredis.RunT(t)

		cfg := &config.Config{}
		cfg.Redis.Address = mr.Addr()

		client := NewRedis(cfg)
		defer client.Close()

		if err := client.Ping(t.Context()).Err(); err != nil {
			t.Errorf("Ping() error = %v, want nil", err)
		}
	})

	t.Run("Uses The Configured DB Index", func(t *testing.T) {
		mr := miniredis.RunT(t)

		cfg := &config.Config{}
		cfg.Redis.Address = mr.Addr()
		cfg.Redis.Db = 3

		client := NewRedis(cfg)
		defer client.Close()

		if err := client.Set(t.Context(), "key", "value", 0).Err(); err != nil {
			t.Fatalf("Set() error = %v, want nil", err)
		}

		// miniredis' out-of-band Exists/Get helpers check against whichever
		// DB was last selected on the shared instance (default 0), so
		// select DB 3 to confirm the client actually wrote there.
		mr.Select(3)
		if !mr.Exists("key") {
			t.Error("expected key to have been written to DB index 3, matching cfg.Redis.Db")
		}
		mr.Select(0)
		if mr.Exists("key") {
			t.Error("expected key NOT to exist in DB index 0")
		}
	})

	t.Run("Panics When The Server Is Unreachable", func(t *testing.T) {
		cfg := &config.Config{}
		// Port 0 isn't a listening service on any host, so the connect
		// attempt fails fast (connection refused).
		cfg.Redis.Address = "127.0.0.1:0"

		defer func() {
			if r := recover(); r == nil {
				t.Error("expected NewRedis to panic when the server is unreachable, it did not")
			}
		}()

		NewRedis(cfg)
	})
}
