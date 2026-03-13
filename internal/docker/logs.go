package docker

import (
	"context"
	"io"
	"strconv"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

func LogsContainer(ctx context.Context, cli *client.Client, idOrName string, follow bool, tail int) (io.ReadCloser, error) {
	tailStr := "all"
	if tail >= 0 {
		tailStr = strconv.Itoa(tail)
	}

	opts := types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     follow,
		Tail:       tailStr,
	}

	return cli.ContainerLogs(ctx, idOrName, opts)
}
