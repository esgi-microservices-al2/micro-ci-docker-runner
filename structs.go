package main

type CommandMessage struct {
	BuildID   string     `json:"buildId"`
	ProjectID string     `json:"projectId"`
	Folder    string     `json:"folder"`
	Commands  [][]string `json:"commands"`
}

type DockerId struct {
	ID string `json:"Id"`
}

type DockerStream struct {
	Stream string `json:"stream"`
}

type DockerAux struct {
	Aux DockerId `json:"aux"`
}

type DockerMessage struct {
	Message string `json:"message"`
}

type DockerError struct {
	ErrorDetail DockerMessage `json:"errorDetail"`
	Error       string        `json:"error"`
}

type EventMessage struct {
	Subject   string      `json:"subject"`
	ProjectID string      `json:"projectId"`
	BuildID   string      `json:"buildId"`
	Date      int64       `json:"date"`
	Content   interface{} `json:"content"`
	Type      string      `json:"type"`
}

type CommandResult struct {
	ExitCode int    `json:"exitCode"`
	Stdout   string `json:"stdout"`
}
