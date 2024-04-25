package main

import (
	"context"
	"fmt"
	"github.com/containerd/containerd"
	"github.com/containerd/containerd/namespaces"
	"minik8s/pkg/api_obj"
	"minik8s/pkg/api_obj/obj_inner"
	"minik8s/pkg/kubelet/container_manager"
	"minik8s/pkg/kubelet/pod_manager"
)

//func main() {
//	res := util.PrintCmd("test", "pull", "redis:latest")
//	fmt.Println(res)
//	// res, _ = util.RunContainer("test", "pause")
//	// fmt.Println(res)
//	client, err := containerd.New("/run/containerd/containerd.sock")
//	defer client.Close()
//	if err != nil {
//		fmt.Println("Create Client Failed")
//	}
//	// img := GetImageFromLocal(client, "docker.io/library/redis:alpine")
//	// imgs := ListImages(client)
//	ctx := namespaces.WithNamespace(context.Background(), "test")
//	// fmt.Printf("Successfully pulled %s image\n", image.Name())
//	// img, err := client.GetImage(ctx, "docker.io/library/jobserver:alpine")
//	img, _ := image_manage.FetchImage(client, "registry.aliyuncs.com/google_containers/pause:latest", ctx)
//	if img == nil {
//		fmt.Println(err)
//	}
//
//	err = container_manager.CreatePauseContainer("test")
//	fmt.Println("here")
//	if err != nil {
//		fmt.Println("starting", err)
//		return
//	}
//
//	err = container_manager.DeletePauseContainer("test", "pause")
//	if err != nil {
//		return
//	}
//	fmt.Println("Delete Succeed")
//	//res, err = image_manage.DeleteImage("test", "docker.io/library/busybox:latest")
//	//if err != nil {
//	//	fmt.Println(res, err)
//	//}
//
//}

func main() {
	client, err := containerd.New("/run/containerd/containerd.sock")
	defer client.Close()
	pod := api_obj.Pod{
		ApiVersion: "1",
		Kind:       "pod",
		MetaData: obj_inner.ObjectMeta{
			Name:      "testpod",
			NameSpace: "test1",
			Labels: map[string]string{
				"testlabel": "podlabel",
			},
			Annotations: nil,
			UUID:        "",
		},
		Spec: api_obj.PodSpec{
			Containers: []api_obj.Container{
				{
					Name: "testName",
					Image: obj_inner.Image{
						Img:           "docker.io/library/ubuntu:latest",
						ImgPullPolicy: "Always",
					},
					EntryPoint: obj_inner.EntryPoint{
						//Command:    []string{"ls"},
						WorkingDir: "/",
					},
					Ports: []obj_inner.ContainerPort{
						{
							ContainerPort: 0,
							HostIP:        "0.0.0.0",
							HostPort:      0,
							Name:          "no name",
							Protocol:      "TCP",
						},
					},
					Env: []obj_inner.EnvVar{
						{
							Name:  "env1",
							Value: "env1Value",
						},
					},
					VolumeMounts: []obj_inner.VolumeMount{
						{
							MountPath: "/home",
							SubPath:   "config",
							Name:      "testMount",
							ReadOnly:  false,
						},
					},
					Resources: obj_inner.ResourceRequirements{
						Limits: map[string]obj_inner.Quantity{
							"CPU":    obj_inner.Quantity("0.5"),
							"Memory": obj_inner.Quantity("200MiB"),
						},
						Requests: map[string]obj_inner.Quantity{
							"CPU":    obj_inner.Quantity("0.25"),
							"Memory": obj_inner.Quantity("100MiB"),
						},
					},
				},
			},
			Volumes: []obj_inner.Volume{
				{
					Name: "testMount",
					Type: "",
					Path: "/testOOO",
				},
			},
			NodeName:      "",
			NodeSelector:  nil,
			RestartPolicy: "",
		},
	}
	fmt.Println(pod.Spec.Containers[0].Name)

	ctx := namespaces.WithNamespace(context.Background(), pod.MetaData.NameSpace)
	if client == nil {
		fmt.Println("Create Client Failed : ", err.Error())
	}

	containers, err := container_manager.ListContainers(client, ctx)
	if err != nil {
		fmt.Println("List Containers Failed : ", err)
	}

	if len(containers) > 0 {
		fmt.Println("There should be no containers in \"test\"")
	}

	//createdContainer, err := container_manager.CreateK8sContainer(ctx, client, &container, "test", volumes, "")
	//if createdContainer == nil {
	//	fmt.Println("Create Container Failed ", err.Error())
	//}
	//defer createdContainer.Delete(ctx, containerd.WithSnapshotCleanup)
	//pid, err := container_manager.StartContainer(ctx, createdContainer)
	//if pid == 0 {
	//	fmt.Println("Run Container Failed ", err.Error())
	//}
	err = pod_manager.AddPod(&pod)
	if err != nil {
		fmt.Println("Main Failed At line 154 ", err.Error())
	}
	fmt.Println("Pod Ip is ", pod.PodStatus.PodIP)
}
