package container_manager

import (
	"context"
	"fmt"
	"minik8s/pkg/api_obj"
	"minik8s/pkg/api_obj/obj_inner"
	"testing"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/namespaces"
)

func TestCreateK8sContainer(t *testing.T) {
	client, err := containerd.New("/run/containerd/containerd.sock")
	defer client.Close()
	volumes := []obj_inner.Volume{
		{
			Name: "testMount",
			Type: "",
			Path: "/mydata/mnttest",
		},
	}
	var container = api_obj.Container{
		Name: "testName",
		Image: obj_inner.Image{
			Img:           "docker.io/library/busybox:latest",
			ImgPullPolicy: "Always",
		},
		EntryPoint: obj_inner.EntryPoint{
			Command: []string{"ls"},
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
				MountPath: "/var/lib",
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
	}
	fmt.Println(container.Name)

	nameSpace := "test"
	ctx := namespaces.WithNamespace(context.Background(), nameSpace)
	if client == nil {
		t.Fatal("Create Client Failed : ", err.Error())
	}

	containers, err := ListContainers(client, ctx)
	if containers == nil {
		t.Fatal("List Containers Failed : ", err.Error())
	}

	if len(containers) > 0 {
		t.Fatal("There should be no containers in \"test\"")
	}

	createdContainer, err := CreateK8sContainer(ctx, client, &container, "test", volumes, "")
	if createdContainer == nil {
		t.Fatal("Create Container Failed ", err.Error())
	}
	defer createdContainer.Delete(ctx, containerd.WithSnapshotCleanup)
	pid, err := StartContainer(ctx, createdContainer)
	if pid == 0 {
		t.Fatal("Run Container Failed ", err.Error())
	}

}
