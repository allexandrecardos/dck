package docker

import (
	"context"

	"github.com/docker/docker/api/types/container"
)

type Docker interface {
	ListContainers(ctx context.Context, all bool) ([]container.Summary, error)
	GetContainerOptions(ctx context.Context, all bool) ([]ContainerOption, error)
	StopContainer(ctx context.Context, id string) error
}
