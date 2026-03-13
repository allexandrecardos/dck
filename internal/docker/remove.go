package docker

import (
	"context"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

func RemoveContainer(ctx context.Context, cli *client.Client, id string, force bool, removeVolumes bool) error {
	return cli.ContainerRemove(ctx, id, types.ContainerRemoveOptions{
		Force:         force,
		RemoveVolumes: removeVolumes,
	})
}

func RemoveImage(ctx context.Context, cli *client.Client, id string) error {
	_, err := cli.ImageRemove(ctx, id, types.ImageRemoveOptions{
		Force:         false,
		PruneChildren: true,
	})
	return err
}

func RemoveVolume(ctx context.Context, cli *client.Client, name string) error {
	return cli.VolumeRemove(ctx, name, true)
}

func RemoveNetwork(ctx context.Context, cli *client.Client, id string) error {
	return cli.NetworkRemove(ctx, id)
}

func StopContainerIfRunning(ctx context.Context, cli *client.Client, id string) error {
	timeout := 10
	return cli.ContainerStop(ctx, id, container.StopOptions{Timeout: &timeout})
}
