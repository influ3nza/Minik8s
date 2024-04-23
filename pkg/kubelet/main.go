package main

import (
	"SE3356/pkg/kubelet/container_manager"
	"SE3356/pkg/kubelet/image_manage"
	"context"
	"fmt"
	"github.com/containerd/containerd"
	"github.com/containerd/containerd/namespaces"
)
import "SE3356/pkg/kubelet/util"

func main() {
	res := util.PrintCmd("test", "pull", "redis:latest")
	fmt.Println(res)
	// res, _ = util.RunContainer("test", "pause")
	// fmt.Println(res)
	client, err := containerd.New("/run/containerd/containerd.sock")
	defer client.Close()
	if err != nil {
		fmt.Println("Create Client Failed")
	}
	// img := GetImageFromLocal(client, "docker.io/library/redis:alpine")
	// imgs := ListImages(client)
	ctx := namespaces.WithNamespace(context.Background(), "test")
	// fmt.Printf("Successfully pulled %s image\n", image.Name())
	// img, err := client.GetImage(ctx, "docker.io/library/jobserver:alpine")
	img, _ := image_manage.FetchImage(client, "registry.aliyuncs.com/google_containers/pause:latest", ctx)
	if img == nil {
		fmt.Println(err)
	}

	err = container_manager.CreatePauseContainer("test")
	fmt.Println("here")
	if err != nil {
		fmt.Println("starting", err)
		return
	}

	err = container_manager.DeletePauseContainer("test", "pause")
	if err != nil {
		return
	}
	fmt.Println("Delete Succeed")
	//res, err = image_manage.DeleteImage("test", "docker.io/library/busybox:latest")
	//if err != nil {
	//	fmt.Println(res, err)
	//}

}
