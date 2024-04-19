package container_manager

import (
	"context"
	"github.com/containerd/containerd"
	"github.com/containerd/containerd/oci"
)

func ListContainers(client *containerd.Client, ctx context.Context, filter ...string) ([]containerd.Container, error) {
	res, err := client.Containers(ctx, filter...)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func CreateContainer(client *containerd.Client, ctx context.Context, opts ...oci.SpecOpts) {

}
