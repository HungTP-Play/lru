package shared

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQ struct {
	connectionString string
	connection       *amqp.Connection
	ctx              context.Context
}

func getRabbitConnectionString() string {
	rabbitHost := os.Getenv("RABBITMQ_HOST")
	rabbitPort := os.Getenv("RABBITMQ_PORT")

	return fmt.Sprintf("amqp://guest:guest@%v:%v/", rabbitHost, rabbitPort)
}

func NewRabbitMQ(connectionString string) *RabbitMQ {
	if connectionString == "" {
		connectionString = getRabbitConnectionString()
	}
	return &RabbitMQ{
		connectionString: connectionString,
		ctx:              context.Background(),
	}
}

func (r *RabbitMQ) Connect(delay time.Duration) error {
	if delay > 0 {
		time.Sleep(delay)
	}

	connection, err := amqp.Dial(r.connectionString)
	if err != nil {
		return err
	}

	r.connection = connection
	return nil
}

func (r *RabbitMQ) Close() error {
	return r.connection.Close()
}

func (r *RabbitMQ) Publish(queue string, message interface{}) error {
	if r.connection.IsClosed() {
		r.Connect(0)
	}

	channel, err := r.connection.Channel()
	if err != nil {
		return err
	}

	defer channel.Close()

	_, err = channel.QueueDeclare(queue, true, false, false, false, nil)
	if err != nil {
		return err
	}

	body, err := json.Marshal(message)
	if err != nil {
		return err
	}

	err = channel.PublishWithContext(r.ctx, "", queue, true, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        body,
	})

	return err
}

func (r *RabbitMQ) Consume(queue string, callback func([]byte) error, numberOfWorker int) error {
	if r.connection.IsClosed() {
		r.Connect(0)
	}

	channel, err := r.connection.Channel()
	if err != nil {
		return err
	}
	defer channel.Close()

	_, err = channel.QueueDeclare(queue, true, false, false, false, nil)
	if err != nil {
		return err
	}

	forever := make(chan bool)

	msgs, err := channel.Consume(queue, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	for i := 0; i < numberOfWorker; i++ {
		go func() {
			for d := range msgs {
				err := callback(d.Body)
				if err != nil {
					d.Nack(false, true)
				}
				d.Ack(false)
			}
		}()
	}

	<-forever

	return nil
}
