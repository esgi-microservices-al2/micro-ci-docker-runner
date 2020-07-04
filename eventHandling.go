package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/streadway/amqp"
)

// HandleMessage ... Handles a message from the commands microservice
func HandleMessage(message CommandMessage, folderTar string, folderProjects string, eventChannel *amqp.Channel, eventQueue string) {
	ID := fmt.Sprintf("%s_%s", message.ProjectID, message.BuildID)

	var destPath = fmt.Sprintf("%s/%s", folderTar, ID)
	var projectPath = fmt.Sprintf("%s/%s", folderProjects, message.Folder)
	var archivePath = destPath + "/archive.tar"
	var context = "Dockerfile"

	defer deleteWorkspaceHandler(destPath)

	err := os.Mkdir(destPath, 644)
	if err != nil {
		log.Printf(err.Error())
		return
	}

	err = tarCreationHandler(projectPath, archivePath, message.ProjectID, message.BuildID, eventChannel, eventQueue)
	if err != nil {
		log.Println(err.Error())
		return
	}

	imageID, err := buildImageHandler(archivePath, context, message.ProjectID, message.BuildID, eventChannel, eventQueue)
	if err != nil {
		log.Println(err.Error())
		return
	}

	containerID, err := createContainerHandler(imageID, ID, message.ProjectID, message.BuildID, eventChannel, eventQueue)
	defer deleteContainerHandler(containerID)

	if err != nil {
		return
	}

	for _, cmd := range message.Commands {
		err = execCommandHandler(cmd, containerID, message.ProjectID, message.BuildID, eventChannel, eventQueue)
		if err != nil {
			return
		}
	}

	successfullBuild(message.ProjectID, message.BuildID, eventChannel, eventQueue)
}

func execCommandHandler(command []string, containerID string, projectID string, buildID string, ch *amqp.Channel, q string) error {
	var message EventMessage = EventMessage{
		Subject:   "Command",
		BuildID:   buildID,
		ProjectID: projectID,
	}

	exitCode, stdout, err := ExecCommand(command, containerID)
	if err != nil {
		message.Date = (time.Now()).Unix()
		message.Content = err.Error()
		message.Type = "error"

		SendEventMessage(message, ch, q)
		return err
	}

	message.Date = (time.Now()).Unix()
	message.Content = CommandResult{
		Command:  strings.Join(command, " "),
		ExitCode: exitCode,
		Stdout:   stdout,
	}

	if exitCode != 0 {
		message.Type = "error"
		SendEventMessage(message, ch, q)
		return err
	}

	message.Type = "success"
	SendEventMessage(message, ch, q)

	return nil
}

func createContainerHandler(imageID string, name string, projectID string, buildID string, ch *amqp.Channel, q string) (string, error) {
	var message EventMessage = EventMessage{
		Subject:   "Build",
		BuildID:   buildID,
		ProjectID: projectID,
	}

	containerID, err := CreateContainer(imageID, name)
	if err != nil {
		message.Date = (time.Now()).Unix()
		message.Content = err.Error()
		message.Type = "error"

		SendEventMessage(message, ch, q)
		return "", err
	}

	message.Date = (time.Now()).Unix()
	message.Content = "Container created successfully."
	message.Type = "info"

	SendEventMessage(message, ch, q)
	return containerID, nil
}

func buildImageHandler(archivePath string, context string, projectID string, buildID string, ch *amqp.Channel, q string) (string, error) {
	var message EventMessage = EventMessage{
		Subject:   "Build",
		BuildID:   buildID,
		ProjectID: projectID,
	}

	imageID, stdout, err := BuildImage(archivePath, context)
	if err != nil {
		message.Date = (time.Now()).Unix()
		message.Content = err.Error()
		message.Type = "error"

		SendEventMessage(message, ch, q)
		return "", err
	}

	message.Date = (time.Now()).Unix()
	message.Content = stdout
	message.Type = "info"

	SendEventMessage(message, ch, q)
	return imageID, nil
}

func tarCreationHandler(projectPath string, archivePath string, projectID string, buildID string, ch *amqp.Channel, q string) error {
	var message EventMessage = EventMessage{
		Subject:   "Build",
		BuildID:   buildID,
		ProjectID: projectID,
	}

	err := CreateTar(projectPath, archivePath)
	if err != nil {
		message.Content = err.Error()
		message.Date = (time.Now()).Unix()
		message.Type = "error"

		SendEventMessage(message, ch, q)
		return err
	}

	return nil
}

func successfullBuild(projectID string, buildID string, ch *amqp.Channel, q string) {
	var message EventMessage = EventMessage{
		Subject:   "Build",
		BuildID:   buildID,
		ProjectID: projectID,
		Date:      (time.Now()).Unix(),
		Content:   "Build successful !",
		Type:      "success",
	}

	SendEventMessage(message, ch, q)
}

func deleteWorkspaceHandler(path string) {
	err := os.RemoveAll(path)
	if err != nil {
		log.Println(err.Error())
	}
}

func deleteContainerHandler(containerID string) {
	err := DeleteContainer(containerID)
	if err != nil {
		log.Println(err.Error())
	}
}
