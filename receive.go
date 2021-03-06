package main

import (
	"encoding/json"
	"fmt"
	"log"

	"net/url"

	"github.com/streadway/amqp"
)

func main() {
	host := Getenv("RABBIT_HOST", "localhost")
	user := Getenv("RABBIT_USER", "docker")
	password := Getenv("RABBIT_PASSWORD", "docker")
	port := Getenv("RABBIT_PORT", "5672")
	queueName := Getenv("RABBIT_RUNNER_QUEUE", "commands")
	eventQueueName := Getenv("RABBIT_EVENT_QUEUE", "events")
	folderTar := Getenv("FOLDER_TAR", ".")
	folderProjects := Getenv("FOLDER_PROJECTS", ".")
	consulUri := Getenv("CONSUL_URI", "localhost:8300")
	consulToken := Getenv("CONSUL_TOKEN", "token")

	consulClient, err := NewConsulClient(consulUri, consulToken)
	FailOnError(err, "Failed to connect to consul")
	err = consulClient.Register()
	FailOnError(err, "Failed to register service into consul")
	log.Printf("Registered on Consul successfully")

	connectionString := fmt.Sprintf("amqp://%s:%s@%s:%s/", user, url.QueryEscape(password), host, port)

	conn, err := amqp.Dial(connectionString)
	FailOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	channel, err := conn.Channel()
	FailOnError(err, "Failed to open a channel")
	defer channel.Close()

	eventChannel, err := conn.Channel()
	FailOnError(err, "Failed to open a second channel")
	defer eventChannel.Close()

	q, err := channel.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	FailOnError(err, "Failed to declare a queue")

	msgs, err := channel.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	FailOnError(err, "Failed to register a consumer")
	forever := make(chan bool)

	go func() {
		for d := range msgs {
			var msg CommandMessage
			json.Unmarshal([]byte(d.Body), &msg)

			d.Ack(false)
			HandleMessage(msg, folderTar, folderProjects, eventChannel, eventQueueName)
		}
	}()

	log.Printf("Waiting for messages. To exit press CTRL+C")
	<-forever
}
