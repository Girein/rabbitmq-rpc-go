# RabbitMQ RPC for Go

## Installation
`go get github.com/Girein/rabbitmq-rpc-go`

## Setup
Add these in your `.env` file:
```
RABBITMQ_HOST=localhost
RABBITMQ_PORT=5672
RABBITMQ_USERNAME=guest
RABBITMQ_PASSWORD=guest
RABBITMQ_VHOST_FOOBAR=""
RABBITMQ_QUEUE_FOOBAR=rpc_queue
```

## Usage
```
import "github.com/Girein/rabbitmq-rpc-go"

rabbitMQConnection := new(rabbitmq.Connection)
rabbitMQConnection.New("FOOBAR")
status, message, data := rabbitmq.NewRPCRequest(rabbitMQConnection, rabbitmq.Payload{"RPC Testing", nil, nil})

fmt.Println(status)
fmt.Println(message)
fmt.Println(data)
```