package main

import (
	"fmt"
	"log"

	"github.com/streadway/amqp"
)

func main() {
	host := Getenv("RABBIT_HOST", "localhost")
	user := Getenv("RABBIT_USER", "docker")
	password := Getenv("RABBIT_PASSWORD", "docker")
	port := Getenv("RABBIT_PORT", "5672")
	queueName := Getenv("RABBIT_RUNNER_QUEUE", "commands")

	connectionString := fmt.Sprintf("amqp://%s:%s@%s:%s/", user, password, host, port)

	conn, err := amqp.Dial(connectionString)
	FailOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	channel, err := conn.Channel()
	FailOnError(err, "Failed to open a channel")
	defer channel.Close()

	q, err := channel.QueueDeclare(
		queueName, // name
		false,     // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	FailOnError(err, "Failed to declare a queue")

	msgs, err := channel.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	FailOnError(err, "Failed to register a consumer")
	forever := make(chan bool)

	go func() {
		for d := range msgs {
			log.Printf("Received a message : %s", d.Body)
			// Call docker commands here
		}
	}()

	log.Printf("Waiting for messages. To exit press CTRL+C")
	<-forever
}
