package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
)

// HandleMessage ... Handles a message from the commands microservice
func HandleMessage(message CommandMessage, folderTar string, folderProjects string) {
	randomID, err := uuid.NewRandom()
	if err != nil {
		log.Printf(err.Error())
		return
	}

	var destPath = fmt.Sprintf("%s/%s", folderTar, randomID)
	var projectPath = fmt.Sprintf("%s/%s", folderProjects, message.Folder)
	var archivePath = destPath + "/archive.tar"
	var context = "Dockerfile"

	err = os.Mkdir(destPath, 644)
	if err != nil {
		log.Printf(err.Error())
		return
	}

	err = CreateTar(projectPath, archivePath)
	if err != nil {
		log.Println(err.Error())
		return
	}

	imageID, err := buildImageHandler(archivePath, context, randomID.String(), randomID.String())
	if err != nil {
		log.Println(err.Error())
		return
	}

	containerID, err := createContainerHandler(imageID, "test", randomID.String(), randomID.String())
	defer DeleteContainer(containerID)

	if err != nil {
		return
	}

	for _, cmd := range message.Commands {
		err = execCommandHandler(cmd, containerID, randomID.String(), randomID.String())
		if err != nil {
			return
		}
	}
}

func execCommandHandler(command []string, containerID string, buildID string, projectID string) error {
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

		SendEventMessage(message)
		return err
	}

	message.Date = (time.Now()).Unix()
	message.Content = CommandResult{
		ExitCode: exitCode,
		Stdout:   stdout,
	}

	if exitCode != 0 {
		message.Type = "error"
		SendEventMessage(message)
		return err
	}

	message.Type = "success"
	SendEventMessage(message)

	return nil
}

func createContainerHandler(imageID string, name string, buildID string, projectID string) (string, error) {
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

		SendEventMessage(message)
		return "", err
	}

	message.Date = (time.Now()).Unix()
	message.Content = "Container created successfully."
	message.Type = "info"

	SendEventMessage(message)
	return containerID, nil
}

func buildImageHandler(archivePath string, context string, buildID string, projectID string) (string, error) {
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

		SendEventMessage(message)
		return "", err
	}

	message.Date = (time.Now()).Unix()
	message.Content = stdout
	message.Type = "info"

	SendEventMessage(message)
	return imageID, nil
}
