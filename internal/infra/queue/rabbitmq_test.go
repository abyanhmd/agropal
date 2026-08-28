package queue

import (
	"agropal/config"
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"
)

func TestBuildConnectionString(t *testing.T) {
	t.Run("Formats Amqp URL From Config", func(t *testing.T) {
		cfg := &config.Config{}
		cfg.RabbitMQ.Username = "user"
		cfg.RabbitMQ.Password = "pass"
		cfg.RabbitMQ.Address = "broker.internal:5672"

		got := buildConnectionString(cfg)
		want := "amqp://user:pass@broker.internal:5672/"
		if got != want {
			t.Errorf("buildConnectionString() = %q, want %q", got, want)
		}
	})

	t.Run("Handles Empty Credentials", func(t *testing.T) {
		cfg := &config.Config{}
		cfg.RabbitMQ.Address = "localhost:5672"

		got := buildConnectionString(cfg)
		want := "amqp://:@localhost:5672/"
		if got != want {
			t.Errorf("buildConnectionString() = %q, want %q", got, want)
		}
	})
}

// These IsHealthy/Close/Publish/Consume tests exercise the nil-guard
// branches directly against a zero/partially-populated RabbitMQ struct,
// without needing a live broker - white-box, same-package access to the
// unexported connection/channel fields is what makes that possible.
func TestRabbitMQ_IsHealthy(t *testing.T) {
	t.Run("False When Never Connected", func(t *testing.T) {
		r := &RabbitMQ{}
		if r.IsHealthy() {
			t.Error("IsHealthy() = true, want false for a zero-value RabbitMQ")
		}
	})

	t.Run("False When Channel Is Nil", func(t *testing.T) {
		r := &RabbitMQ{connection: nil, channel: nil}
		if r.IsHealthy() {
			t.Error("IsHealthy() = true, want false when channel is nil")
		}
	})
}

func TestRabbitMQ_Close(t *testing.T) {
	t.Run("No Error When Nothing Was Ever Connected", func(t *testing.T) {
		r := &RabbitMQ{}
		if err := r.Close(); err != nil {
			t.Errorf("Close() error = %v, want nil", err)
		}
	})
}

func TestRabbitMQ_PublishMessageToQueue(t *testing.T) {
	t.Run("Errors When Channel Is Not Initialized", func(t *testing.T) {
		r := &RabbitMQ{logger: testLogger}
		err := r.PublishMessageToQueue(context.Background(), QueueNameEmail, []byte(`{}`))
		if err == nil {
			t.Fatal("PublishMessageToQueue() error = nil, want an error for a nil channel")
		}
	})
}

func TestRabbitMQ_ConsumeMessages(t *testing.T) {
	t.Run("Errors When Channel Is Not Initialized", func(t *testing.T) {
		r := &RabbitMQ{logger: testLogger}
		err := r.ConsumeMessages(context.Background(), QueueNameEmail, func(context.Context, []byte) error { return nil })
		if err == nil {
			t.Fatal("ConsumeMessages() error = nil, want an error for a nil channel")
		}
	})
}

// newTestRabbitMQ connects to the shared broker container and purges the
// default queue first, so leftover messages from a previous subtest can't
// leak into this one. It registers its own cleanup.
func newTestRabbitMQ(t *testing.T) *RabbitMQ {
	t.Helper()
	if testConfig == nil {
		t.Skip("no test broker available (see main_test.go)")
	}

	r, err := NewRabbitMQ(testConfig, testLogger)
	if err != nil {
		t.Fatalf("NewRabbitMQ() error = %v, want nil", err)
	}
	if _, err := r.channel.QueuePurge(r.queue.Name, false); err != nil {
		t.Fatalf("failed to purge queue before test: %v", err)
	}
	t.Cleanup(func() { _ = r.Close() })

	return r
}

func TestNewRabbitMQ(t *testing.T) {
	t.Run("Connects Declares Queue And Reports Healthy", func(t *testing.T) {
		r := newTestRabbitMQ(t)

		if !r.IsHealthy() {
			t.Error("IsHealthy() = false, want true right after connecting")
		}
	})
}

func TestRabbitMQ_PublishAndConsume(t *testing.T) {
	t.Run("Round Trips A Message Through The Default Queue", func(t *testing.T) {
		r := newTestRabbitMQ(t)

		task := EmailVerificationTask{
			Purpose:   "register",
			To:        "user@example.com",
			Code:      "123456",
			ExpiresAt: time.Now().Add(10 * time.Minute).UTC(),
		}
		payload, err := json.Marshal(task)
		if err != nil {
			t.Fatalf("failed to marshal task: %v", err)
		}

		if err := r.PublishMessage(context.Background(), payload); err != nil {
			t.Fatalf("PublishMessage() error = %v, want nil", err)
		}

		received := make(chan []byte, 1)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		go func() {
			_ = r.ConsumeMessages(ctx, QueueNameEmail, func(_ context.Context, body []byte) error {
				received <- body
				cancel() // stop the consumer loop once we've got our message
				return nil
			})
		}()

		select {
		case body := <-received:
			var got EmailVerificationTask
			if err := json.Unmarshal(body, &got); err != nil {
				t.Fatalf("failed to unmarshal consumed message: %v", err)
			}
			if got.To != task.To || got.Code != task.Code || got.Purpose != task.Purpose {
				t.Errorf("consumed task = %+v, want %+v", got, task)
			}
		case <-time.After(5 * time.Second):
			t.Fatal("timed out waiting to consume the published message")
		}
	})

	t.Run("Nacked Message Is Requeued For Retry", func(t *testing.T) {
		r := newTestRabbitMQ(t)

		if err := r.PublishMessage(context.Background(), []byte(`{"retry":true}`)); err != nil {
			t.Fatalf("PublishMessage() error = %v, want nil", err)
		}

		var attempts int
		attemptErr := errors.New("simulated processing failure")

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		done := make(chan struct{})
		go func() {
			_ = r.ConsumeMessages(ctx, QueueNameEmail, func(_ context.Context, body []byte) error {
				attempts++
				if attempts < 2 {
					return attemptErr
				}
				close(done)
				cancel()
				return nil
			})
		}()

		select {
		case <-done:
			if attempts < 2 {
				t.Errorf("attempts = %d, want at least 2 (nack should requeue for a retry)", attempts)
			}
		case <-time.After(10 * time.Second):
			t.Fatal("timed out waiting for the message to be retried and acked")
		}
	})
}

func TestRabbitMQ_GetQueueInfo(t *testing.T) {
	t.Run("Reports Pending Message Count", func(t *testing.T) {
		r := newTestRabbitMQ(t)

		if err := r.PublishMessage(context.Background(), []byte(`{}`)); err != nil {
			t.Fatalf("PublishMessage() error = %v, want nil", err)
		}

		// Publishing is async from the broker's perspective; give it a brief
		// moment to land before inspecting the queue.
		time.Sleep(200 * time.Millisecond)

		count, _, err := r.GetQueueInfo(context.Background())
		if err != nil {
			t.Fatalf("GetQueueInfo() error = %v, want nil", err)
		}
		if count < 1 {
			t.Errorf("GetQueueInfo() messageCount = %d, want at least 1", count)
		}
	})
}

func TestRabbitMQ_CloseIntegration(t *testing.T) {
	t.Run("Closes Connection And Publishing Afterward Fails", func(t *testing.T) {
		r := newTestRabbitMQ(t)

		if err := r.Close(); err != nil {
			t.Fatalf("Close() error = %v, want nil", err)
		}
		if r.IsHealthy() {
			t.Error("IsHealthy() = true, want false after Close()")
		}
		if err := r.PublishMessage(context.Background(), []byte(`{}`)); err == nil {
			t.Error("PublishMessage() error = nil, want an error after the connection was closed")
		}
	})
}

func TestRabbitMQ_Reconnect(t *testing.T) {
	t.Run("Restores A Healthy Connection After Close", func(t *testing.T) {
		r := newTestRabbitMQ(t)

		if err := r.Close(); err != nil {
			t.Fatalf("Close() error = %v, want nil", err)
		}

		if err := r.Reconnect(context.Background()); err != nil {
			t.Fatalf("Reconnect() error = %v, want nil", err)
		}
		if !r.IsHealthy() {
			t.Error("IsHealthy() = false, want true after Reconnect()")
		}
		if err := r.PublishMessage(context.Background(), []byte(`{}`)); err != nil {
			t.Errorf("PublishMessage() error = %v, want nil after Reconnect()", err)
		}
	})
}
