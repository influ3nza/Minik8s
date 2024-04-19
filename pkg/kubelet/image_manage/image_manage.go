package image_manage

import (
	"SE3356/pkg/kubelet/util"
	"context"
	"github.com/containerd/containerd"
)

func GetImageFromLocal(client *containerd.Client, name string, ctx context.Context) containerd.Image {
	img, err := client.GetImage(ctx, name)
	if err != nil {
		return nil
	}
	return img
}

func ListImages(client *containerd.Client, ctx context.Context, filters ...string) []containerd.Image {
	imgs, err := client.ListImages(ctx, filters...)
	if err != nil {
		return nil
	}
	return imgs
}

func FetchImage(client *containerd.Client, imgName string, ctx context.Context) containerd.Image {
	pull, err := client.Pull(ctx, imgName, containerd.WithPullUnpack)
	if err != nil {
		println(err)
		return nil
	}
	return pull
}

func DeleteImage(namespace string, imgName string) (string, error) {
	util.PrintCmd(namespace, "rmi", imgName)
	res, err := util.Exec(namespace, "rmi", imgName)
	if err != nil {
		return "", err
	}
	return res, err
}
