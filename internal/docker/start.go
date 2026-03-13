package docker

import (
	"context"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

func StartContainer(ctx context.Context, cli *client.Client, idOrName string) error {
	return cli.ContainerStart(ctx, idOrName, container.StartOptions{})
}
