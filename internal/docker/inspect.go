package docker

import (
	"context"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

func InspectContainer(ctx context.Context, cli *client.Client, idOrName string) (types.ContainerJSON, error) {
	return cli.ContainerInspect(ctx, idOrName)
}
