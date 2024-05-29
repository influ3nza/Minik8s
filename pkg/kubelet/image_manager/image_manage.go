package image_manager

import (
	"context"
	"fmt"
	"minik8s/pkg/api_obj/obj_inner"
	"minik8s/pkg/kubelet/util"
	"os/exec"
	"strings"

	"github.com/containerd/containerd"
)

// GetImageFromLocal 从本地获取镜像
/*
 * 参数
 *  client: *containerd.Client 客户端
 *  name: string 名称
 *  ctx: context.Context 上下文
 *
 * 返回
 *  containerd.Image: 一个containerd.Image类型的对象
 *  error: 错误信息
 */
func GetImageFromLocal(client *containerd.Client, name string, ctx context.Context) (containerd.Image, error) {
	img, err := client.GetImage(ctx, name)
	if err != nil {
		return nil, err
	}
	return img, nil
}

// FetchMasterImage 从master获取镜像
/*
 * 参数
 *  client: *containerd.Client 客户端
 *  imgName: string 镜像名称
 *  namespace: string 命名空间
 *
 * 返回
 *  error: 错误信息
 */
func FetchMasterImage(client *containerd.Client, imgName string, namespace string) error {
	cmd := []string{"-n", namespace, "pull", "--insecure-registry", imgName}
	// util.PrintCmd(namespace, cmd...)
	err := exec.Command("nerdctl", cmd...).Run()
	if err != nil {
		fmt.Printf("fetch From Master Failed %s", err.Error())
		return fmt.Errorf("fetch From Master Failed %s", err.Error())
	}
	return nil
}

// ListImages 列出镜像
/*
 * 参数
 *  client: *containerd.Client 客户端
 *  ctx: context.Context 上下文
 *  filters: ...string 过滤器
 *
 * 返回
 *  []containerd.Image: 一个containerd.Image类型的切片
 */
func ListImages(client *containerd.Client, ctx context.Context, filters ...string) []containerd.Image {
	imgs, err := client.ListImages(ctx, filters...)
	if err != nil {
		return nil
	}
	return imgs
}

// FetchImage 拉取镜像
/*
 * 参数
 *  client: *containerd.Client 客户端
 *  imgName: string 镜像名称
 *  ctx: context.Context 上下文
 *
 * 返回
 *  containerd.Image: 一个containerd.Image类型的对象
 *  error: 错误信息
 */
func FetchImage(client *containerd.Client, imgName string, ctx context.Context) (containerd.Image, error) {
	pull, err := client.Pull(ctx, imgName, containerd.WithPullUnpack)
	if err != nil {
		return nil, err
	}
	return pull, nil
}

// DeleteImage 删除镜像
/*
 * 参数
 *  namespace: string 命名空间
 *  imgName: string 镜像名称
 *
 * 返回
 *  string: 一个字符串
 *  error: 错误信息
 */
func DeleteImage(namespace string, imgName string) (string, error) {
	util.PrintCmd(namespace, "rmi", imgName)
	res, err := util.Exec(namespace, "rmi", imgName)
	if err != nil {
		return "", err
	}
	return res, err
}

// GetImage 获取镜像
/*
 * 参数
 *  client: *containerd.Client 客户端
 *  image: *obj_inner.Image 镜像
 *  ctx: context.Context 上下文
 *
 * 返回
 *  containerd.Image: 一个containerd.Image类型的对象
 *  error: 错误信息
 */
func GetImage(client *containerd.Client, image *obj_inner.Image, ctx context.Context, namespace string) (containerd.Image, error) {
	ParseImage(image)
	if image.ImgPullPolicy == "Always" {
		if strings.Contains(image.Img, "my-registry.io") {
			err := FetchMasterImage(client, image.Img, namespace)
			if err != nil {
				return nil, err
			}
			img, err := GetImageFromLocal(client, image.Img, ctx)
			if err != nil {
				return nil, err
			}
			return img, nil
		}
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
