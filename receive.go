package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"net/url"

	"github.com/google/uuid"
	"github.com/streadway/amqp"
)

var folderProjects string = Getenv("FOLDER_PROJECTS", ".")
var folderTar string = Getenv("FOLDER_TAR", ".")

type commandMessage struct {
	Folder   string   `json:"folder"`
	Commands []string `json:"commands"`
}

func handleMessage(message commandMessage) {
	randomID, err := uuid.NewRandom()
	if err != nil {
		log.Printf(err.Error())
		return
	}

	var destPath = fmt.Sprintf("%s/%s", folderTar, randomID)
	var projectPath = fmt.Sprintf("%s/%s", folderProjects, message.Folder)
	var archivePath = destPath + "/archive.tar"
	var context = fmt.Sprintf("%s/Dockerfile", message.Folder)

	err = os.Mkdir(destPath, 644)
	if err != nil {
		log.Printf(err.Error())
		return
	}

	file, err := os.Create(destPath + "/logs")
	if err != nil {
		log.Printf(err.Error())

		return
	}

	err = CreateTar(projectPath, archivePath)
	if err != nil {
		log.Printf(err.Error())
		return
	}

	BuildImage(archivePath, context, file)
}

func main() {
	host := Getenv("RABBIT_HOST", "localhost")
	user := Getenv("RABBIT_USER", "docker")
	password := Getenv("RABBIT_PASSWORD", "docker")
	port := Getenv("RABBIT_PORT", "5672")
	queueName := Getenv("RABBIT_RUNNER_QUEUE", "commands")

	connectionString := fmt.Sprintf("amqp://%s:%s@%s:%s/", user, url.QueryEscape(password), host, port)

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
			var msg commandMessage
			json.Unmarshal([]byte(d.Body), &msg)

			handleMessage(msg)

			d.Ack(false)
		}
	}()

	log.Printf("Waiting for messages. To exit press CTRL+C")
	<-forever
}
