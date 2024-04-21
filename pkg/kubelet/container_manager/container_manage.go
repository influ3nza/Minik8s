package container_manager

import (
	"SE3356/pkg/api_obj"
	"SE3356/pkg/kubelet/image_manage"
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

func CreateContainer(client *containerd.Client, ctx context.Context, container api_obj.Container, namespace string) {
	image, err := image_manage.GetImage(client, &container.Image, ctx)
	if err != nil {
		return
	}
	create_opts := []oci.SpecOpts{oci.WithImageConfig(image)}
	limitOpts, err := ParseResources(container.Resources)
	if err == nil {
		create_opts = append(create_opts, limitOpts...)
	}

}
