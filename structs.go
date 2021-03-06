package main

type CommandMessage struct {
	BuildID   int        `json:"buildId"`
	ProjectID int        `json:"projectId"`
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
	ProjectID int         `json:"projectId"`
	BuildID   int         `json:"buildId"`
	Date      int64       `json:"date"`
	Content   interface{} `json:"content"`
	Type      string      `json:"type"`
}

type CommandResult struct {
	Command  string `json:"command"`
	ExitCode int    `json:"exitCode"`
	Stdout   string `json:"stdout"`
}
