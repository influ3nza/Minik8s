package container_manager

import (
	"SE3356/pkg/api_obj"
	"SE3356/pkg/kubelet/image_manage"
	"SE3356/pkg/kubelet/util"
	"context"
	"encoding/json"
	"fmt"
	"github.com/containerd/containerd"
	"github.com/containerd/containerd/cio"
	"github.com/containerd/containerd/oci"
	"github.com/opencontainers/runtime-spec/specs-go"
)

func ListContainers(client *containerd.Client, ctx context.Context, filter ...string) ([]containerd.Container, error) {
	res, err := client.Containers(ctx, filter...)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// CreateK8sContainer 创建一个proj容器
// 参数：
//
//	client : *containerd.Client
//	ctx 	 : context.Context
//	container : api_obj.Container
//
// 返回值：
//
//	container : containerd.Container
func CreateK8sContainer(ctx context.Context, client *containerd.Client, container *api_obj.Container, pod *api_obj.Pod, linuxNamespace string) (containerd.Container, error) {
	// 解析image，配置容器image选项
	image, err := image_manage.GetImage(client, &container.Image, ctx)
	if err != nil {
		return nil, err
	}
	createOpts := []oci.SpecOpts{oci.WithImageConfig(image)}

	// 配置hostname，不允许用户设置，使用default --> pod.MetaData.Name
	createOpts = append(createOpts, oci.WithHostname(pod.MetaData.Name))

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
	entryOci = append(entryOci, oci.WithProcessCwd(container.EntryPoint.WorkingDir))
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
		mounts, e := convertMounts(&pod.Spec, container)
		if e != nil {
			fmt.Println("At CreateK8sContainer line 80 ", e.Error())
			return nil, e
		}
		if mounts != nil && len(mounts) > 0 {
			var ociMounts = make([]specs.Mount, len(mounts), len(mounts))
			i := 0
			for _, mount := range mounts {
				ociMounts[i] = specs.Mount{
					Destination: mount.Container_,
					Source:      mount.Host_ + mount.Subdir_,
					Type:        "bind",
					Options:     []string{"bind"},
				}
			}
			createOpts = append(createOpts, oci.WithMounts(ociMounts))
		}
	}

	// 配置容器namespace pid net ipc uts
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

	// 配置容器labels，包含"podName" ： podname/ "ports" : serialize(container.ports)
	podLabel := map[string]string{
		"podName": pod.MetaData.Name,
	}
	var portLabel = map[string]string{}
	if len(container.Ports) > 0 {
		jsonFy, err := json.Marshal(container.Ports)
		if err != nil {
			return nil, err
		}
		portLabel["ports"] = string(jsonFy)
	}
	// 配置容器snap-shot，将oci opts合并入容器opts
	containerOpts := []containerd.NewContainerOpts{
		containerd.WithNewSnapshot(container.Name+"snapshot", image),
		containerd.WithNewSpec(createOpts...),
		containerd.WithAdditionalContainerLabels(podLabel),
		containerd.WithAdditionalContainerLabels(portLabel),
	}

	containerCreated, err := client.NewContainer(ctx, container.Name, containerOpts...)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	return containerCreated, nil
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

func StopAndRmContainer(pod *api_obj.Pod, container containerd.Container, ifForce bool) error {
	name := container.ID()
	namespace := pod.MetaData.NameSpace
	if ifForce == false {
		_, err := util.StopContainer(namespace, name)
		if err != nil {
			fmt.Println("At Func StopAndRmContainer line 172 ", err.Error())
			return err
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

func CreatePauseContainer(namespace string) error {
	res, err := util.RunContainer(namespace, "pause")
	if err != nil {
		return err
	}
	fmt.Println("res is :", res)
	return nil
}

func DeletePauseContainer(namespace string, name string) error {
	_, _ = util.StopContainer(namespace, name)
	_, _ = util.RemoveContainer(namespace, name)
	return nil
}
