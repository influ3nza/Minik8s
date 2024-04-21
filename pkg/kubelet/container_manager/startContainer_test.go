package container_manager

import (
	"SE3356/pkg/api_obj/obj_inner"
	"context"
	"fmt"
	"testing"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/namespaces"
	"github.com/stretchr/testify/assert"
)

func TestStartContainer(t *testing.T) {
	// 连接到 containerd
	client, err := containerd.New("\\\\.\\pipe\\containerd-containerd")
	if err != nil {
		fmt.Println("Failed to create client:", err)
		return
	}
	defer client.Close()

	// 创建一个虚拟的 Image 对象
	image := &obj_inner.Image{
		Img:           "your-image-name",
		ImgPullPolicy: "Always",
	}

	// 创建一个上下文
	ctx := namespaces.WithNamespace(context.Background(), "your-namespace")

	// 调用被测试的函数
	startContainer(client, image, ctx)

	// 添加断言来验证函数的行为是否符合预期

	// 断言是否成功创建了容器
	containers, err := client.LoadContainer(ctx, "my-container")
	assert.NoError(t, err, "Failed to load container")
	assert.NotNil(t, containers, "Container not created")
	fmt.Println("Container created:", containers.ID())

	// 在这里继续添加其他断言，验证更多的行为
}
