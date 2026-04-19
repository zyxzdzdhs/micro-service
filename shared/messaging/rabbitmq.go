package messaging

import (
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQ struct {
	conn *amqp.Connection
}

func NewRabbitMQ(uri string) (*RabbitMQ, error) {
	conn, err := amqp.Dial(uri)
	if err != nil {
		return nil, fmt.Errorf("Failed to connect to rabbitmq: %v", err)
	}

	rmq := &RabbitMQ{
		conn: conn,
	}

	return rmq, nil
}

func (rmq *RabbitMQ) Close() {
	if rmq.conn != nil {
		rmq.conn.Close()
	}
}
