package image_manager

import (
	"fmt"
	"minik8s/pkg/api_obj/obj_inner"
	"strings"
)

// 解析image的pull policy 当有latest tag时，总是pull， 否则如果不设置字段，则先从本地查找
func ParseImage(img *obj_inner.Image) {
	if !strings.Contains(img.Img, ":") {
		img.Img += ":latest"
	} else if strings.Contains(img.Img[strings.Index(img.Img, ":"):], "/") {
		if !strings.Contains(img.Img[strings.Index(img.Img, ":")+1:], ":") {
			img.Img += ":latest"
		}
	}

	if !strings.Contains(img.Img, "/") {
		img.Img = "docker.io/library/" + img.Img
	}

	if strings.Contains(img.Img, ":latest") {
		img.ImgPullPolicy = "Always"
	}
	fmt.Println("Parsed Image is ", img.Img)
}
