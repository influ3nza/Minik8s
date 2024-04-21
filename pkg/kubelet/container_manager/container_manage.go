package container_manager

import (
	"SE3356/pkg/api_obj"
	"SE3356/pkg/api_obj/obj_inner"
	"SE3356/pkg/kubelet/image_manage"
	"context"
	"fmt"

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

func startContainer(client *containerd.Client, image *obj_inner.Image, ctx context.Context) {
	requirements := obj_inner.ResourceRequirements{
		Requests: map[string]obj_inner.Quantity{
			obj_inner.CPU_REQUEST:    "0.5",
			obj_inner.MEMORY_REQUEST: "256Mi",
		},
		Limits: map[string]obj_inner.Quantity{
			obj_inner.CPU_LIMIT:    "1",
			obj_inner.MEMORY_LIMIT: "512Mi",
		},
	}
	specOpts, err := ParseResources(requirements)
	if err != nil {
		fmt.Println("Failed to parse resources:", err)
		return
	}

	myImage, err := image_manage.GetImage(client, image, ctx)
	if err != nil {
		fmt.Println("Failed to get images:", err)
		return
	}

	// 创建容器配置
	containerOpts := []oci.SpecOpts{
		oci.WithImageConfig(myImage),
	}
	containerOpts = append(containerOpts, specOpts...)

	// 创建一个容器
	container, err := client.NewContainer(ctx,
		"my-container",
		containerd.WithNewSpec(containerOpts...),
	)
	if err != nil {
		fmt.Println("Failed to create container:", err)
		return
	}
	defer container.Delete(ctx, containerd.WithSnapshotCleanup)

	// 启动容器
	task, err := container.NewTask(ctx, nil)
	if err != nil {
		fmt.Println("Failed to create task:", err)
		return
	}
	defer task.Delete(ctx)

	// 等待容器退出
	exitStatus, err := task.Wait(ctx)
	if err != nil {
		fmt.Println("Failed to wait for task:", err)
		return
	}

	fmt.Println("Container exited with status:", exitStatus)
}
