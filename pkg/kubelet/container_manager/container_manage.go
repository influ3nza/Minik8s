package container_manager

import (
	"context"
	"encoding/json"
	"fmt"
	"minik8s/pkg/api_obj"
	"minik8s/pkg/api_obj/obj_inner"
	"minik8s/pkg/kubelet/image_manager"
	"minik8s/pkg/kubelet/util"
	"os"
	"time"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/cio"
	"github.com/containerd/containerd/oci"
	"github.com/opencontainers/runtime-spec/specs-go"
)

// CreateK8sContainer 创建一个proj容器
// 参数：
//
//	ctx 	 : context.Context
//	client : *containerd.Client
//	container : *api_obj.Container
//	metaName : metaData.Name in Pod
//	podVolumes : volumes on host in Pod
//	linuxNamespace : Create Pause And Then Generated
//
// 返回值：
//
//	container : containerd.Container
func CreateK8sContainer(ctx context.Context, client *containerd.Client, container *api_obj.Container, metaName string, podVolumes []obj_inner.Volume, linuxNamespace string) (containerd.Container, string, error) {
	// 解析image，配置容器image选项
	image, err := image_manager.GetImage(client, &container.Image, ctx)
	if err != nil {
		return nil, "", err
	}
	createOpts := []oci.SpecOpts{oci.WithImageConfig(image)}

	// 配置hostname，不允许用户设置，使用default --> pod.MetaData.Name
	createOpts = append(createOpts, oci.WithHostname(metaName))

	//配置容器资源
	limitOpts, err := ParseResources(container.Resources)
	if err == nil {
		createOpts = append(createOpts, limitOpts...)
	}

	// 配置容器环境
	if container.Env != nil {
		envs := convertEnv(container)
		envOci := oci.WithEnv(envs)
		createOpts = append(createOpts, envOci)
	}

	// 配置容器EntryPoint：包括WorkDir，Cmd，Args
	var entryOci []oci.SpecOpts
	if container.EntryPoint.WorkingDir != "" {
		entryOci = append(entryOci, oci.WithProcessCwd(container.EntryPoint.WorkingDir))
	}
	if len(container.EntryPoint.Command) != 0 {
		var processArgs []string
		for _, cmd := range container.EntryPoint.Command {
			processArgs = append(processArgs, cmd)
		}
		for _, arg := range container.EntryPoint.Args {
			processArgs = append(processArgs, arg)
		}
		entryOci = append(entryOci, oci.WithProcessArgs(processArgs...))
	}
	createOpts = append(createOpts, entryOci...)

	// 配置容器Mount todo source&dest
	if container.VolumeMounts != nil {
		mounts, e := convertMounts(podVolumes, container)
		if e != nil {
			fmt.Println("At CreateK8sContainer line 80 ", e.Error())
			return nil, "", e
		}
		if mounts != nil && len(mounts) > 0 {
			var ociMounts = make([]specs.Mount, len(mounts), len(mounts))
			i := 0
			for _, mount := range mounts {
				ociMounts[i] = specs.Mount{
					Destination: mount.Container_,
					Source:      mount.Host_ + "/" + mount.Subdir_,
					Type:        "bind",
					Options:     []string{"bind"},
				}
			}
			for _, Mounts := range ociMounts {
				fmt.Println(Mounts.Destination, Mounts.Source)
				_, err := os.Stat(Mounts.Source)
				if os.IsNotExist(err) {
					_, err := util.Mkdir(Mounts.Source)
					if err != nil {
						return nil, "", err
					}
				}
			}
			createOpts = append(createOpts, oci.WithMounts(ociMounts))
		}
	}

	// 配置容器namespace pid net ipc uts /proc/%pid/ns/
	if linuxNamespace != "" {
		var linuxNamespaces = map[string]string{
			"pid":     linuxNamespace + "pid",
			"network": linuxNamespace + "net",
			"ipc":     linuxNamespace + "ipc",
			"uts":     linuxNamespace + "uts",
		}
		var ociLinuxNsOpt []oci.SpecOpts
		for nameSpaceType, nameSpace := range linuxNamespaces {
			linuxNsSpace := specs.LinuxNamespace{
				Type: specs.LinuxNamespaceType(nameSpaceType),
				Path: nameSpace,
			}
			ociLinuxNsOpt = append(ociLinuxNsOpt, oci.WithLinuxNamespace(linuxNsSpace))
		}
		createOpts = append(createOpts, ociLinuxNsOpt...)
	}

	// 配置容器labels，包含“Name” : container.Name, "podName" : podname/ "ports" : serialize(container.ports)
	nameLabel := map[string]string{
		"Name": container.Name,
	}
	podLabel := map[string]string{
		"podName": metaName,
	}
	var portLabel = map[string]string{}
	if len(container.Ports) > 0 {
		jsonFy, err := json.Marshal(container.Ports)
		if err != nil {
			return nil, "", err
		}
		portLabel["ports"] = string(jsonFy)
	}
	// 配置容器snap-shot，将oci opts合并入容器opts
	containerOpts := []containerd.NewContainerOpts{
		containerd.WithNewSnapshot(container.Name+"snapshot", image),
		containerd.WithNewSpec(createOpts...),
		containerd.WithContainerLabels(nameLabel),
		containerd.WithAdditionalContainerLabels(podLabel),
		containerd.WithAdditionalContainerLabels(portLabel),
	}

	containerId, err := GenerateUUIDForContainer()
	if err != nil {
		fmt.Println("Failed At CreateK8sContainer line 154 ", err.Error())
		return nil, "", err
	}

	containerCreated, err := client.NewContainer(ctx, containerId, containerOpts...)
	if err != nil {
		fmt.Println(err.Error())
		return nil, "", err
	}

	return containerCreated, containerId, nil
}

// StartContainer 启动创建好的container
// 参数
//
//	ctx context.Context
//	container containerd.Container 通过createContainer创建
//
// 返回
//
//	uint32 pid
//	err error
func StartContainer(ctx context.Context, container containerd.Container) (uint32, error) {
	newTask, err := container.NewTask(ctx, cio.NewCreator(cio.WithStdio))
	if err != nil {
		fmt.Println("New Task failed: ", err.Error())
		return 0, err
	}

	err = newTask.Start(ctx)
	if err != nil {
		fmt.Println("Start Task failed: ", err.Error())
		return 0, err
	}
	return newTask.Pid(), nil
}

func StopAndRmContainer(namespace string, name string, ifForce bool) error {
	// name := container.ID()
	// namespace := pod.MetaData.NameSpace
	if ifForce == false {
		_, err := util.StopContainer(namespace, name)
		if err != nil {
			fmt.Println("At Func StopAndRmContainer line 172 ", err.Error())
			//return err
		}

		_, err = util.RemoveContainer(namespace, name)
		if err != nil {
			fmt.Println("At Func StopAndRmContainer line 178 ", err.Error())
			return err
		}
	} else {
		_, err := util.RmForce(namespace, name)
		if err != nil {
			fmt.Println("At Func StopAndRmContainer line 184 ", err.Error())
			return err
		}
	}
	return nil
}

func DeleteContainerInPod(ctx context.Context, client *containerd.Client, podName string, podNamespace string, ifForce bool) error {
	walker := &ContainerWalker{
		Client: client,
		OnFound: func(ctx context.Context, found Found) error {
			fmt.Println(found.Container.ID())
			err := StopAndRmContainer(podNamespace, found.Container.ID(), ifForce)
			if err != nil {
				fmt.Println("Stop Rm Container in Walker Failed at line 221 ", err.Error())
			}
			return nil
		},
	}

	filter := map[string]string{
		"podName": podName,
	}
	_, err := walker.Walk(ctx, filter)
	if err != nil {
		fmt.Println("Failed At DeleteContainerInPod line 228 ", err.Error())
		return err
	}
	return nil
}

func CreatePauseContainer(ctx context.Context, client *containerd.Client, namespace string, name string) (string, error) {
	// client, err := containerd.New("/run/containerd/containerd.sock")
	// ctx := namespaces.WithNamespace(context.Background(), namespace)
	img := obj_inner.Image{
		Img:           util.FirstSandbox,
		ImgPullPolicy: "Always",
	}
	_, err := image_manager.GetImage(client, &img, ctx)
	if err != nil {
		fmt.Println("Pull Pause Image Failed At CreatePauseContainer line 224, ", err.Error())
		return "", err
	}

	res, err := util.RunContainer(namespace, name)
	if err != nil {
		return "", err
	}
	fmt.Println("res is :", res)
	return res, nil
}

func DeletePauseContainer(namespace string, id string) error {
	_, _ = util.StopContainer(namespace, id)
	_, _ = util.RemoveContainer(namespace, id)
	return nil
}

// MonitorPodContainers 监控指定 Pod 中的所有容器
func MonitorPodContainers(ctx context.Context, client *containerd.Client, podName string) {
	// 获取所有容器
	containers, err := client.Containers(ctx)
	if err != nil {
		fmt.Println("Failed to get containers:", err)
		return
	}

	// 遍历所有容器
	for _, container := range containers {
		// 检查容器的标签，查找所属的 Pod
		labels, err := container.Labels(ctx)
		if err != nil {
			fmt.Println("Failed to get container labels:", err)
			continue
		}

		// 检查容器是否属于指定的 Pod
		if podLabel, ok := labels["podName"]; ok && podLabel == podName {
			// 如果属于指定的 Pod，则监控该容器的状态
			go MonitorContainerState(ctx, client, container.ID())
		}
	}
}

// MonitorContainerState 监控容器的状态
func MonitorContainerState(ctx context.Context, client *containerd.Client, containerID string) {
	for {
		// 加载容器
		container, err := client.LoadContainer(ctx, containerID)
		if err != nil {
			fmt.Println("Failed to load container:", err)
			return
		}

		// 获取容器的任务
		task, err := container.Task(ctx, nil)
		if err != nil {
			fmt.Println("Failed to get task:", err)
			return
		}

		// 获取任务的状态
		status, err := task.Status(ctx)
		if err != nil {
			fmt.Println("Failed to get task status:", err)
			return
		}

		// 打印容器的状态信息
		fmt.Printf("Container %s Status: %v\n", containerID, status)

		// 等待一段时间后继续监控
		time.Sleep(10 * time.Second)
	}
}
