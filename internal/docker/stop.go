package docker

import (
	"context"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

func StopContainer(ctx context.Context, cli *client.Client, idOrName string, timeout time.Duration) error {
	opts := container.StopOptions{}
	if timeout > 0 {
		seconds := int(timeout.Seconds())
		opts.Timeout = &seconds
	}
	return cli.ContainerStop(ctx, idOrName, opts)
}
