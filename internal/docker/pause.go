package docker

import (
	"context"

	"github.com/docker/docker/client"
)

func PauseContainer(ctx context.Context, cli *client.Client, idOrName string) error {
	return cli.ContainerPause(ctx, idOrName)
}
