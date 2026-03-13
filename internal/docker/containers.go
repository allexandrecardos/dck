package docker

import (
	"context"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

func ListContainers(ctx context.Context, cli *client.Client, all bool, size bool) ([]types.Container, error) {
	return cli.ContainerList(ctx, types.ContainerListOptions{All: all, Size: size})
}
