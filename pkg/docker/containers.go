package docker

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types/container"
)

func (d *dockerClient) ListContainers(ctx context.Context, all bool) ([]container.Summary, error) {
	return d.cli.ContainerList(ctx, container.ListOptions{All: all, Size: true})
}

func (d *dockerClient) StopContainer(ctx context.Context, id string) error {
	return d.cli.ContainerStop(ctx, id, container.StopOptions{})
}

func (d *dockerClient) StartContainer(ctx context.Context, id string) error {
	return d.cli.ContainerStart(ctx, id, container.StartOptions{})
}

func (d *dockerClient) GetContainerOptions(ctx context.Context, all bool) ([]ContainerOption, error) {
	containers, err := d.ListContainers(ctx, all)
	if err != nil {
		return nil, err
	}

	options := make([]ContainerOption, 0, len(containers))

	for _, c := range containers {
		idShort := c.ID
		if len(idShort) > 12 {
			idShort = idShort[:12]
		}

		label := fmt.Sprintf("%s - %s - %s", c.Image, idShort, c.Status)

		options = append(options, ContainerOption{
			ID:    c.ID,
			Label: label,
		})
	}

	return options, nil
}
