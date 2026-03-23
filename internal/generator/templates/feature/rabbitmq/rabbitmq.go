package messaging

import (
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

// RabbitMQConfig holds connection details for RabbitMQ.
type RabbitMQConfig struct {
	Host     string `env:"RABBITMQ_HOST"     envDefault:"localhost"`
	Port     int    `env:"RABBITMQ_PORT"     envDefault:"5672"`
	User     string `env:"RABBITMQ_USER"     envDefault:"guest"`
	Password string `env:"RABBITMQ_PASSWORD" envDefault:"guest"`
	VHost    string `env:"RABBITMQ_VHOST"    envDefault:"/"`
}

// RabbitMQConn wraps an AMQP connection and channel.
type RabbitMQConn struct {
	Conn    *amqp.Connection
	Channel *amqp.Channel
}

// NewRabbitMQConn creates an AMQP connection and opens a channel.
// Inject this into your publisher/consumer services.
func NewRabbitMQConn(cfg RabbitMQConfig) (*RabbitMQConn, error) {
	url := fmt.Sprintf("amqp://%s:%s@%s:%d%s",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.VHost,
	)

	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close() //nolint:errcheck
		return nil, fmt.Errorf("failed to open AMQP channel: %w", err)
	}

	log.Println("✅ RabbitMQ connection established")
	return &RabbitMQConn{Conn: conn, Channel: ch}, nil
}

// Close gracefully closes the channel and connection.
func (r *RabbitMQConn) Close() {
	if r.Channel != nil {
		r.Channel.Close() //nolint:errcheck
	}
	if r.Conn != nil && !r.Conn.IsClosed() {
		r.Conn.Close() //nolint:errcheck
	}
}
