package rabbitmq

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	"time"

	"github.com/Girein/helpers-go"
	"github.com/streadway/amqp"
)

var response map[string]interface{}

// Connection RabbitMQ
type Connection struct {
	Host, Port, Username, Password, VirtualHost, QueueName string
}

// New set the RabbitMQ Connection
func (connection *Connection) New(serviceName string) {
	connection.Host = os.Getenv("RABBITMQ_HOST")
	connection.Port = os.Getenv("RABBITMQ_PORT")
	connection.Username = os.Getenv("RABBITMQ_USERNAME_" + serviceName)
	connection.Password = os.Getenv("RABBITMQ_PASSWORD_" + serviceName)
	connection.VirtualHost = os.Getenv("RABBITMQ_VHOST_" + serviceName)
	connection.QueueName = os.Getenv("RABBITMQ_QUEUE_" + serviceName)
}

// NewRPCRequest sends message to the RPC worker
func NewRPCRequest(connection *Connection, body map[string]interface{}) (map[string]interface{}, error) {
	messageBody, err := helpers.JSONEncode(body)
	if err != nil {
		return nil, err
	}

	url := connection.Host + ":" + connection.Port + "/" + connection.VirtualHost

	log.Println("AMQP" + " " + url + " | " + connection.QueueName)

	amqpConnection, err := amqp.Dial("amqp://" + connection.Username + ":" + connection.Password + "@" + url)
	if err != nil {
		return nil, err
	}
	defer amqpConnection.Close()

	channel, err := amqpConnection.Channel()
	if err != nil {
		return nil, err
	}
	defer channel.Close()

	queue, err := channel.QueueDeclare(
		connection.QueueName, // name
		true,                 // durable
		false,                // delete when unused
		false,                // exclusive
		false,                // noWait
		nil,                  // arguments
	)
	if err != nil {
		return nil, err
	}

	messages, err := channel.Consume(
		queue.Name, // queue
		"",         // consumer
		true,       // auto-ack
		false,      // exclusive
		false,      // no-local
		false,      // no-wait
		nil,        // args
	)
	if err != nil {
		return nil, err
	}

	correlationID := helpers.RandomString(32)

	err = channel.Publish(
		"",         // exchange
		queue.Name, // routing key
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			ContentType:   "text/plain",
			CorrelationId: correlationID,
			ReplyTo:       queue.Name,
			Body:          []byte(messageBody),
			Expiration:    "60000",
		})
	if err != nil {
		return nil, err
	}

	flag := make(chan string)

	go func() {
		for data := range messages {
			if correlationID == data.CorrelationId {
				if messageBody == string(data.Body) {
					err = errors.New("the consumer is not responding")
				} else {
					json.Unmarshal([]byte(string(data.Body)), &response)
				}

				break
			}
		}

		flag <- "Successfully to get a response from the consumer"
	}()

	select {
	case result := <-flag:
		log.Println(result)
		return response, nil
	case <-time.After(time.Duration(7) * time.Second):
		return nil, errors.New("the response from the consumer took too long")
	}
}

// SendMessage sends message to the consumer
func SendMessage(connection *Connection, body map[string]interface{}) error {
	messageBody, err := helpers.JSONEncode(body)
	if err != nil {
		return err
	}

	url := connection.Host + ":" + connection.Port + "/" + connection.VirtualHost

	log.Println("AMQP" + " " + url + " | " + connection.QueueName)

	amqpConnection, err := amqp.Dial("amqp://" + connection.Username + ":" + connection.Password + "@" + url)
	if err != nil {
		return err
	}
	defer amqpConnection.Close()

	channel, err := amqpConnection.Channel()
	if err != nil {
		return err
	}
	defer channel.Close()

	queue, err := channel.QueueDeclare(
		connection.QueueName, // name
		true,                 // durable
		false,                // delete when unused
		false,                // exclusive
		false,                // noWait
		nil,                  // arguments
	)
	if err != nil {
		return err
	}

	err = channel.Publish(
		"",         // exchange
		queue.Name, // routing key
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(messageBody),
			Expiration:  "60000",
		})
	if err != nil {
		return err
	}

	return nil
}
