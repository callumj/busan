package remote

import (
	"github.com/fsouza/go-dockerclient"
)

var dockerHost string
var DockerClient *docker.Client

func ConfigureDockerEndpoint(endpoint string) error {
	if len(endpoint) == 0 {
		endpoint = "unix:///var/run/docker.sock"
	}

	consClient, err := docker.NewClient(endpoint)
	if err != nil {
		return err
	}

	DockerClient = consClient
	dockerHost = endpoint

	return nil
}
