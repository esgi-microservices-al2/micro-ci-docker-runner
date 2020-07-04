package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"os"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

// BuildImage  Builds an image from a tar context.
// Path needs to point to a tar. dockerfile is the path to the Dockerfile in the archive.
func BuildImage(path string, dockerfile string) (string, string, error) {
	ctx := context.Background()
	var output string = ""
	var errResult error
	var imgID string

	archive, err := os.Open(path)
	if err != nil {
		log.Println(err.Error())
		return "", "", errors.New("Internal error")
	}

	defer archive.Close()

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Println(err.Error())
		return "", "", errors.New("Internal error")
	}

	options := types.ImageBuildOptions{
		SuppressOutput: false,
		Remove:         true,
		ForceRemove:    true,
		PullParent:     true,
		Dockerfile:     dockerfile,
	}

	out, err := cli.ImageBuild(ctx, archive, options)
	if err != nil {
		log.Println(err.Error())
		return "", "", errors.New("Internal error")
	}

	defer out.Body.Close()

	sc := bufio.NewScanner(out.Body)
	for sc.Scan() {
		output += sc.Text()
		if strings.Index(sc.Text(), "{\"errorDetail\"") == 0 {
			var errorMessage DockerError

			json.Unmarshal([]byte(sc.Text()), &errorMessage)
			imgID, errResult = "", errors.New(errorMessage.Error)
		} else if strings.Index(sc.Text(), "{\"aux\"") == 0 {
			var successMessage DockerAux

			json.Unmarshal([]byte(sc.Text()), &successMessage)
			successMessage.Aux.ID = RemoveEncyptionFromID(successMessage.Aux.ID)
			imgID, errResult = successMessage.Aux.ID, nil
		}
	}

	return imgID, output, errResult
}

// PullImage Simple image pull from docker
func PullImage(imageName string) {
	ctx := context.Background()

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	FailOnError(err, "Failed to connect to docker daemon")

	out, err := cli.ImagePull(ctx, imageName, types.ImagePullOptions{})
	FailOnError(err, "Error while pulling the image")

	defer out.Close()

	io.Copy(os.Stdout, out)
}

// CreateContainer Creates a container with the given image and name
func CreateContainer(imageName string, containerName string) (string, error) {
	ctx := context.Background()

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Println(err.Error())
		return "", errors.New("Internal error")
	}

	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: imageName,
		Cmd:   []string{"tail", "-f", "/dev/null"}, // Command to keep the container running in order to send user desired commands
	}, nil, nil, nil, containerName)
	if err != nil {
		log.Println(err.Error())
		return "", errors.New("Internal error")
	}

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		log.Println(err.Error())
		return "", errors.New("Internal error")
	}

	return resp.ID, nil
}

// ExecCommand Executes a command inside of a running container
func ExecCommand(command []string, container string) (int, string, error) {
	ctx := context.Background()
	var stdout string = ""

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Println(err.Error())
		return 1, "", errors.New("Internal error")
	}

	cmdID, err := cli.ContainerExecCreate(ctx, container, types.ExecConfig{
		AttachStdout: true,
		AttachStderr: true,
		Cmd:          command,
	})
	if err != nil {
		log.Println(err.Error())
		return 1, "", errors.New("Internal error")
	}

	connection, err := cli.ContainerExecAttach(ctx, cmdID.ID, types.ExecStartCheck{})
	if err != nil {
		log.Println(err.Error())
		return 1, "", errors.New("Internal error")
	}
	defer connection.Close()

	sc := bufio.NewScanner(connection.Reader)
	for sc.Scan() {
		stdout += string(sc.Bytes())
	}

	inspection, err := cli.ContainerExecInspect(ctx, cmdID.ID)
	if err != nil {
		log.Println(err.Error())
		return 1, "", errors.New("Internal error")
	}

	return inspection.ExitCode, stdout, nil
}

// DeleteContainer Stops and removes a running container from host
func DeleteContainer(container string) error {
	ctx := context.Background()

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Println(err.Error())
		return errors.New("Internal error")
	}

	err = cli.ContainerRemove(ctx, container, types.ContainerRemoveOptions{Force: true})
	return err
}

// DeleteImage Removes an image after the build
func DeleteImage(imageID string) error {
	ctx := context.Background()

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Println(err.Error())
		return errors.New("Internal error")
	}

	_, err = cli.ImageRemove(ctx, imageID, types.ImageRemoveOptions{Force: true})
	if err != nil {
		log.Println(err.Error())
		return errors.New("Internal error")
	}

	return nil
}
