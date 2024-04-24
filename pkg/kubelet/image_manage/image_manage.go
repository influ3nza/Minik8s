package image_manage

import (
	"context"
	"github.com/containerd/containerd"
	"minik8s/pkg/api_obj/obj_inner"
	"minik8s/pkg/kubelet/util"
)

func GetImageFromLocal(client *containerd.Client, name string, ctx context.Context) (containerd.Image, error) {
	img, err := client.GetImage(ctx, name)
	if err != nil {
		return nil, err
	}
	return img, nil
}

func ListImages(client *containerd.Client, ctx context.Context, filters ...string) []containerd.Image {
	imgs, err := client.ListImages(ctx, filters...)
	if err != nil {
		return nil
	}
	return imgs
}

func FetchImage(client *containerd.Client, imgName string, ctx context.Context) (containerd.Image, error) {
	pull, err := client.Pull(ctx, imgName, containerd.WithPullUnpack)
	if err != nil {
		return nil, err
	}
	return pull, nil
}

func DeleteImage(namespace string, imgName string) (string, error) {
	util.PrintCmd(namespace, "rmi", imgName)
	res, err := util.Exec(namespace, "rmi", imgName)
	if err != nil {
		return "", err
	}
	return res, err
}

func GetImage(client *containerd.Client, image *obj_inner.Image, ctx context.Context) (containerd.Image, error) {
	ParseImage(image)
	if image.ImgPullPolicy == "Always" {
		res, err := FetchImage(client, image.Img, ctx)
		if err != nil {
			return nil, err
		}
		return res, nil
	} else if image.ImgPullPolicy == "IfNotPresent" {
		res, _ := GetImageFromLocal(client, image.Img, ctx)
		if res == nil {
			fetch_res, err := FetchImage(client, image.Img, ctx)
			if err != nil {
				return nil, err
			} else {
				return fetch_res, nil
			}
		} else {
			return res, nil
		}
	} else {
		return nil, nil
	}
}
