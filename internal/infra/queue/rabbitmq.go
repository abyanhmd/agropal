package queue

import (
	"agropal/config"
	"agropal/internal/domain/shared"
	"agropal/pkg"
	"context"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	// Queue names
	QueueNameEmail = "email_task"

	// Connection configuration
	maxConnectionRetries = 5
	retryDelay           = 2 * time.Second
	publishTimeout       = 5 * time.Second
)

// EmailVerificationTask is the payload published to QueueNameEmail for
// one-time verification codes (register/login/reset-password). The email
// worker decodes this to render and send the message, so field changes here
// must stay in lockstep with whatever consumes QueueNameEmail.
type EmailVerificationTask struct {
	Purpose   string    `json:"purpose"`
	To        string    `json:"to"`
	Code      string    `json:"code"`
	ExpiresAt time.Time `json:"expires_at"`
}

// RabbitMQ represents a RabbitMQ connection and channel manager
type RabbitMQ struct {
	logger     *pkg.Logger
	connection *amqp.Connection
	channel    *amqp.Channel
	queue      amqp.Queue
	config     *config.Config
}

// NewRabbitMQ creates a new RabbitMQ connection with retry logic
func NewRabbitMQ(cfg *config.Config, logger *pkg.Logger) (*RabbitMQ, error) {
	ctx := context.Background()

	conn, err := connectWithRetry(ctx, cfg, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to establish RabbitMQ connection: %w", err)
	}

	ch, err := openChannel(conn)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open RabbitMQ channel: %w", err)
	}

	queue, err := declareQueue(ch, QueueNameEmail)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare RabbitMQ queue: %w", err)
	}

	rabbitMQ := &RabbitMQ{
		logger:     logger,
		connection: conn,
		channel:    ch,
		queue:      queue,
		config:     cfg,
	}

	// Set up connection close notification
	go rabbitMQ.monitorConnection(ctx)

	return rabbitMQ, nil
}

// connectWithRetry attempts to connect to RabbitMQ with exponential backoff
func connectWithRetry(ctx context.Context, cfg *config.Config, logger *pkg.Logger) (*amqp.Connection, error) {
	address := buildConnectionString(cfg)

	var conn *amqp.Connection
	var err error

	for attempt := 1; attempt <= maxConnectionRetries; attempt++ {
		conn, err = amqp.Dial(address)
		if err == nil {
			logger.Info(ctx, shared.System, fmt.Sprintf("Successfully connected to RabbitMQ on attempt %d", attempt))
			return conn, nil
		}

		logger.Error(ctx, shared.System, fmt.Sprintf("RabbitMQ connection attempt %d/%d failed: %s", attempt, maxConnectionRetries, err.Error()))

		if attempt < maxConnectionRetries {
			time.Sleep(retryDelay * time.Duration(attempt)) // Exponential backoff
		}
	}

	return nil, fmt.Errorf("failed to connect after %d attempts: %w", maxConnectionRetries, err)
}

// buildConnectionString constructs the AMQP connection string
func buildConnectionString(cfg *config.Config) string {
	return fmt.Sprintf(
		"amqp://%s:%s@%s/",
		cfg.RabbitMQ.Username,
		cfg.RabbitMQ.Password,
		cfg.RabbitMQ.Address,
	)
}

// openChannel opens a new channel on the connection
func openChannel(conn *amqp.Connection) (*amqp.Channel, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	// Set QoS to prevent overwhelming consumers
	if err := ch.Qos(
		10,    // prefetch count
		0,     // prefetch size
		false, // global
	); err != nil {
		ch.Close()
		return nil, fmt.Errorf("failed to set QoS: %w", err)
	}

	return ch, nil
}

// declareQueue declares a queue with the given name
func declareQueue(ch *amqp.Channel, queueName string) (amqp.Queue, error) {
	queue, err := ch.QueueDeclare(
		queueName, // Queue name
		true,      // Durable - survives broker restart
		false,     // Auto-delete - don't delete when unused
		false,     // Exclusive - not exclusive to this connection
		false,     // No-wait - don't wait for server confirmation
		nil,       // Arguments
	)
	if err != nil {
		return amqp.Queue{}, fmt.Errorf("failed to declare queue %s: %w", queueName, err)
	}

	return queue, nil
}

// monitorConnection monitors the connection and logs if it closes unexpectedly
func (r *RabbitMQ) monitorConnection(ctx context.Context) {
	closeErr := <-r.connection.NotifyClose(make(chan *amqp.Error))
	if closeErr != nil {
		r.logger.Error(ctx, shared.System, fmt.Sprintf("RabbitMQ connection closed unexpectedly: %s", closeErr.Error()))
	}
}

// Close gracefully closes the channel and connection
func (r *RabbitMQ) Close() error {
	var channelErr, connErr error

	if r.channel != nil {
		channelErr = r.channel.Close()
	}

	if r.connection != nil {
		connErr = r.connection.Close()
	}

	if channelErr != nil {
		return fmt.Errorf("error closing channel: %w", channelErr)
	}
	if connErr != nil {
		return fmt.Errorf("error closing connection: %w", connErr)
	}

	return nil
}

// PublishMessage publishes a message to the configured default queue with timeout
func (r *RabbitMQ) PublishMessage(ctx context.Context, data []byte) error {
	return r.PublishMessageToQueue(ctx, r.queue.Name, data)
}

// PublishMessageToQueue publishes a message to a specific queue with timeout
func (r *RabbitMQ) PublishMessageToQueue(ctx context.Context, queueName string, data []byte) error {
	if r.channel == nil {
		return fmt.Errorf("channel is not initialized")
	}

	publishCtx, cancel := context.WithTimeout(ctx, publishTimeout)
	defer cancel()

	if err := r.channel.PublishWithContext(
		publishCtx,
		"",        // Exchange - empty for default exchange
		queueName, // Routing key (queue name)
		false,     // Mandatory - don't return undeliverable messages
		false,     // Immediate - deprecated, must be false
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         data,
			DeliveryMode: amqp.Persistent, // Make message persistent
			Timestamp:    time.Now(),
		},
	); err != nil {
		r.logger.Error(ctx, shared.System, fmt.Sprintf("Failed to publish message to queue %s: %s", queueName, err.Error()))
		return fmt.Errorf("failed to publish message to queue %s: %w", queueName, err)
	}

	// Debug, not Info: this runs on every publish and would otherwise flood
	// logs/IO under normal throughput.
	r.logger.Debug(ctx, shared.System, fmt.Sprintf("Message published to queue %s successfully", queueName))
	return nil
}

// IsHealthy checks if the connection and channel are still open
func (r *RabbitMQ) IsHealthy() bool {
	return r.connection != nil && !r.connection.IsClosed() &&
		r.channel != nil
}

// Reconnect attempts to re-establish the connection and channel
func (r *RabbitMQ) Reconnect(ctx context.Context) error {
	r.logger.Info(ctx, shared.System, "Attempting to reconnect to RabbitMQ")

	// Close existing connections
	_ = r.Close()

	// Reconnect
	conn, err := connectWithRetry(ctx, r.config, r.logger)
	if err != nil {
		return fmt.Errorf("failed to reconnect: %w", err)
	}

	ch, err := openChannel(conn)
	if err != nil {
		conn.Close()
		return fmt.Errorf("failed to open channel on reconnect: %w", err)
	}

	queue, err := declareQueue(ch, QueueNameEmail)
	if err != nil {
		ch.Close()
		conn.Close()
		return fmt.Errorf("failed to declare queue on reconnect: %w", err)
	}

	r.connection = conn
	r.channel = ch
	r.queue = queue

	// Restart connection monitoring
	go r.monitorConnection(ctx)

	r.logger.Info(ctx, shared.System, "Successfully reconnected to RabbitMQ")
	return nil
}

// GetQueueInfo returns information about the queue
func (r *RabbitMQ) GetQueueInfo(ctx context.Context) (messageCount, consumerCount int, err error) {
	queue, err := r.channel.QueueInspect(r.queue.Name)
	if err != nil {
		r.logger.Error(ctx, shared.System, fmt.Sprintf("Failed to inspect queue: %s", err.Error()))
		return 0, 0, fmt.Errorf("failed to inspect queue: %w", err)
	}

	return queue.Messages, queue.Consumers, nil
}

// ConsumeMessages starts a blocking consumer on the specified queue.
// It uses a callback handler to process messages and handles manual Acks/Nacks automatically.
func (r *RabbitMQ) ConsumeMessages(ctx context.Context, queueName string, handler func(context.Context, []byte) error) error {
	if r.channel == nil {
		return fmt.Errorf("channel is not initialized")
	}

	// Register the consumer
	deliveries, err := r.channel.Consume(
		queueName, // queue name
		"",        // consumer tag (empty string generates a unique one)
		false,     // auto-ack (false means we MUST manually Ack or Nack)
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		r.logger.Error(ctx, shared.System, fmt.Sprintf("Failed to register consumer on queue %s: %s", queueName, err.Error()))
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	r.logger.Info(ctx, shared.System, fmt.Sprintf("Successfully started consuming from queue: %s", queueName))

	// Loop to listen for messages or context cancellation
	for {
		select {
		case <-ctx.Done():
			r.logger.Info(ctx, shared.System, "Stopping consumer: context canceled")
			return ctx.Err()

		case msg, ok := <-deliveries:
			if !ok {
				r.logger.Error(ctx, shared.System, "RabbitMQ delivery channel closed unexpectedly")
				return fmt.Errorf("delivery channel closed")
			}

			// Pass the message payload to your business logic handler
			err := handler(ctx, msg.Body)
			if err != nil {
				r.logger.Error(ctx, shared.System, fmt.Sprintf("Error processing message: %s", err.Error()))

				// A message that fails every time (e.g. malformed payload) would
				// otherwise requeue-and-fail in a tight loop, burning CPU/network.
				// Pace retries with a short, cancellable backoff.
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(retryDelay):
				}

				// Nack the message.
				// parameter 1 (multiple): false
				// parameter 2 (requeue): true (put it back in the queue to try again)
				// Note: If you have a Dead Letter Queue (DLQ) configured and want to drop it instead, set requeue to false.
				if nackErr := msg.Nack(false, true); nackErr != nil {
					r.logger.Error(ctx, shared.System, fmt.Sprintf("Failed to Nack message: %s", nackErr.Error()))
				}
				continue
			}

			// Acknowledge the message so RabbitMQ removes it from the queue
			if ackErr := msg.Ack(false); ackErr != nil {
				r.logger.Error(ctx, shared.System, fmt.Sprintf("Failed to Ack message: %s", ackErr.Error()))
			}
		}
	}
}
