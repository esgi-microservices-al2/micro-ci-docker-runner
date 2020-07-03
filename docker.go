package main

import (
	"context"
	"io"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
)

// TODO: Improve error handling. the program shouldn't crash if a docker task fails. We shoud notify the user.

// BuildImage ... Builds an image from a tar context.
// Path needs to point to a tar. dockerfile is the path to the Dockerfile in the archive.
func BuildImage(path string, dockerfile string, logs *os.File) {
	ctx := context.Background()

	archive, err := os.Open(path)
	FailOnError(err, "Failed opening build context")
	defer archive.Close()

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	FailOnError(err, "Failed to connect to docker daemon")

	options := types.ImageBuildOptions{
		SuppressOutput: false,
		Remove:         true,
		ForceRemove:    true,
		PullParent:     true,
		Dockerfile:     dockerfile,
	}

	out, err := cli.ImageBuild(ctx, archive, options)
	defer out.Body.Close()
	FailOnError(err, "Failed building the image")

	io.Copy(logs, out.Body)
}

// PullImage ... Simple image pull from docker
func PullImage(imageName string) {
	ctx := context.Background()

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	FailOnError(err, "Failed to connect to docker daemon")

	out, err := cli.ImagePull(ctx, imageName, types.ImagePullOptions{})
	FailOnError(err, "Error while pulling the image")

	defer out.Close()

	io.Copy(os.Stdout, out)
}

// CreateContainer ... Creates a container with the given image, name and commands
func CreateContainer(imageName string, containerName string, commands []string) {
	ctx := context.Background()

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	FailOnError(err, "Failed to connect to docker daemon")

	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: imageName,
		Cmd:   commands,
	}, nil, nil, nil, containerName)
	FailOnError(err, "Failed creating the container")

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		FailOnError(err, "Failed to start the container")
	}

	statusCh, errCh := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			panic(err)
		}
	case <-statusCh:
	}

	out, err := cli.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{ShowStdout: true})
	FailOnError(err, "Failed to fetch container logs")

	stdcopy.StdCopy(os.Stdout, os.Stderr, out)
}
