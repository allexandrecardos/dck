package docker

import (
	"github.com/docker/docker/client"
)

type dockerClient struct {
	cli *client.Client
}

func Client() (Docker, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}
	return &dockerClient{cli}, nil
}
