package image_manager

import (
	"context"
	"github.com/containerd/containerd"
	"github.com/containerd/containerd/namespaces"
	"minik8s/pkg/api_obj/obj_inner"
	"testing"
)

func TestGetImageFromLocal(t *testing.T) {
	client, err := containerd.New("/run/containerd/containerd.sock")
	defer client.Close()
	if err != nil {
		t.Fatal("Create Client Failed")
	}
	// img := GetImageFromLocal(client, "docker.io/library/redis:alpine")
	// imgs := ListImages(client)
	ctx := namespaces.WithNamespace(context.Background(), "test")
	// fmt.Printf("Successfully pulled %s image\n", image.Name())
	// img, err := client.GetImage(ctx, "docker.io/library/jobserver:alpine")
	img, _ := FetchImage(client, "docker.io/library/busybox:latest", ctx)
	if img == nil {
		t.Fatal(err)
	}

	res, err := DeleteImage("test", "docker.io/library/busybox:latest")
	if err != nil {
		t.Fatal(res, err)
	}
}

func TestFetchImage(t *testing.T) {
	client, err := containerd.New("/run/containerd/containerd.sock")
	defer client.Close()
	if err != nil {
		t.Fatal("Create Client Failed")
	}
	ctx := namespaces.WithNamespace(context.Background(), "test")
	img := &obj_inner.Image{
		Img: "my-registry.io:5000/foo",
	}
	_, err = GetImage(client, img, ctx)
	if err != nil {
		t.Fatal("Get Image Failed")
	}
	_, err = DeleteImage("test", img.Img)
	if err != nil {
		t.Fatal("Delete Img Error")
	}
}

//func TestListImage(t *testing.T) {
//	client, err := containerd.New("/run/containerd/containerd.sock")
//	defer client.Close()
//	if err != nil {
//		t.Fatal("Create Client Failed")
//	}
//	ctx := namespaces.WithNamespace(context.Background(), "test")
//	img, _ := GetImageFromLocal(client, "docker.io/library/busybox:latest", ctx)
//	if img == nil {
//		t.Fatal()
//	}
//
//	imgs := ListImages(client, ctx)
//	if imgs == nil {
//		t.Fatal()
//	}
//}

//func TestParseImage(t *testing.T) {
//	var s1 obj_inner.Image = obj_inner.Image{
//		Img:           "busybox",
//		ImgPullPolicy: "IfNotPresent",
//	}
//
//	ParseImage(&s1)
//
//	t.Log(s1.Img, "  ", s1.ImgPullPolicy)
//}
