package rabbitmq

import (
	"encoding/json"
	"log"
	"os"

	"github.com/Girein/helpers-go"
	"github.com/streadway/amqp"
)

var response map[string]interface{}

// Payload RabbitMQ
type Payload struct {
	Route string      `json:"route"`
	Param interface{} `json:"param"`
	Data  interface{} `json:"data"`
}

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
func NewRPCRequest(connection *Connection, body map[string]interface{}) map[string]interface{} {
	url := connection.Host + ":" + connection.Port + "/" + connection.VirtualHost

	log.Println("AMQP" + " " + url + " | " + connection.QueueName)

	amqpConnection, err := amqp.Dial("amqp://" + connection.Username + ":" + connection.Password + "@" + url)
	helpers.LogIfError(err, "Failed to connect to RabbitMQ")
	defer amqpConnection.Close()

	channel, err := amqpConnection.Channel()
	helpers.LogIfError(err, "Failed to open a channel in RabbitMQ")
	defer channel.Close()

	queue, err := channel.QueueDeclare(
		connection.QueueName, // name
		true,                 // durable
		false,                // delete when unused
		false,                // exclusive
		false,                // noWait
		nil,                  // arguments
	)
	helpers.LogIfError(err, "Failed to declare a queue in RabbitMQ")

	messages, err := channel.Consume(
		queue.Name, // queue
		"",         // consumer
		true,       // auto-ack
		false,      // exclusive
		false,      // no-local
		false,      // no-wait
		nil,        // args
	)
	helpers.LogIfError(err, "Failed to register a consumer in RabbitMQ")

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
			Body:          []byte(helpers.JSONEncode(body)),
			Expiration:    "60000",
		})
	helpers.LogIfError(err, "Failed to publish a message in RabbitMQ")

	for data := range messages {
		if correlationID == data.CorrelationId {
			json.Unmarshal([]byte(string(data.Body)), &response)
			break
		}
	}

	return response
}

// SendMessage sends message to the consumer
func SendMessage(connection *Connection, body map[string]interface{}) {
	url := connection.Host + ":" + connection.Port + "/" + connection.VirtualHost

	log.Println("AMQP" + " " + url + " | " + connection.QueueName)

	amqpConnection, err := amqp.Dial("amqp://" + connection.Username + ":" + connection.Password + "@" + url)
	helpers.LogIfError(err, "Failed to connect to RabbitMQ")
	defer amqpConnection.Close()

	channel, err := amqpConnection.Channel()
	helpers.LogIfError(err, "Failed to open a channel in RabbitMQ")
	defer channel.Close()

	queue, err := channel.QueueDeclare(
		connection.QueueName, // name
		true,                 // durable
		false,                // delete when unused
		false,                // exclusive
		false,                // noWait
		nil,                  // arguments
	)
	helpers.LogIfError(err, "Failed to declare a queue in RabbitMQ")

	err = channel.Publish(
		"",         // exchange
		queue.Name, // routing key
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(helpers.JSONEncode(body)),
			Expiration:  "60000",
		})
	helpers.LogIfError(err, "Failed to publish a message in RabbitMQ")
}
