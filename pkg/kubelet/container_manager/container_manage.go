package container_manager

import (
	"context"
	"encoding/json"
	"fmt"
	"minik8s/pkg/api_obj"
	"minik8s/pkg/api_obj/obj_inner"
	"minik8s/pkg/kubelet/image_manager"
	"minik8s/pkg/kubelet/pod_manager"
	"minik8s/pkg/kubelet/util"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"time"

	v2 "github.com/containerd/cgroups/v2/stats"
	"github.com/containerd/containerd"
	"github.com/containerd/containerd/cio"
	"github.com/containerd/containerd/oci"
	"github.com/gogo/protobuf/proto"
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
		fmt.Println("append opts")
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
			fmt.Println("At CreateK8sContainer line 82 ", e.Error())
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
	containerId, err := GenerateUUIDForContainer()
	if err != nil {
		fmt.Println("Failed At CreateK8sContainer line 155 ", err.Error())
		return nil, "", err
	}
	// 配置容器snap-shot，将oci opts合并入容器opts
	containerOpts := []containerd.NewContainerOpts{
		containerd.WithNewSnapshot(container.Name+containerId+"snapshot", image),
		containerd.WithNewSpec(createOpts...),
		containerd.WithContainerLabels(nameLabel),
		containerd.WithAdditionalContainerLabels(podLabel),
		containerd.WithAdditionalContainerLabels(portLabel),
	}

	containerCreated, err := client.NewContainer(ctx, containerId, containerOpts...)
	if err != nil {
		fmt.Println(err.Error())
		return nil, "", err
	}

	return containerCreated, containerId, nil
}

func StartFuncContainer(client *containerd.Client, namespace string, name string, podName string) (string, error) {
	err := image_manager.FetchMasterImage(client, name, namespace)
	if err != nil {
		return "", fmt.Errorf("fetch Func Image Failed %s", err.Error())
	}
	cmd := []string{"-n", namespace, "run", "--name", podName, "--net", "flannel", "--label", fmt.Sprintf("podName=%s", podName), name}
	opt, err := exec.Command("nerdctl", cmd...).CombinedOutput()

	if err != nil {
		fmt.Println("Create Func Failed At line 178")
		return "", err
	}
	funcContainerId := strings.TrimSuffix(string(opt), "\n")
	// 截取后面64个字母
	if len(funcContainerId) > 64 {
		funcContainerId = funcContainerId[len(funcContainerId)-64:]
	}

	ip, err := pod_manager.GetPodIp(namespace, funcContainerId)
	if err != nil {
		fmt.Println("Get Func Ip Failed ", err.Error())
		return "", err
	}

	return ip, nil
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

func StopContainer(namespace string, name string) {
	_, err := util.StopContainer(namespace, name)
	if err != nil {
		fmt.Println("At Func StopAndRmContainer line 172 ", err.Error())
	}
}

func StopContainerInPod(ctx context.Context, client *containerd.Client, namespace string, podName string) error {
	walker := &ContainerWalker{
		Client: client,
		OnFound: func(ctx context.Context, found Found) error {
			// fmt.Println(found.Container.ID())
			StopContainer(namespace, found.Container.ID())
			return nil
		},
	}

	filter := map[string]string{
		"podName": podName,
	}

	_, err := walker.Walk(ctx, filter)
	if err != nil {
		fmt.Println("Failed At DeleteContainerInPod line 216 ", err.Error())
		return err
	}
	return nil
}

func StopAndRmContainer(namespace string, name string, ifForce bool) error {
	if ifForce == false {
		_, err := util.StopContainer(namespace, name)
		if err != nil {
			fmt.Println("At Func StopAndRmContainer line 226 ", err.Error())
			//return err
		}

		_, err = util.RemoveContainer(namespace, name)
		if err != nil {
			fmt.Println("At Func StopAndRmContainer line 232 ", err.Error())
			return err
		}
	} else {
		_, err := util.RmForce(namespace, name)
		if err != nil {
			fmt.Println("At Func StopAndRmContainer line 238 ", err.Error())
			return err
		}
	}
	return nil
}

func DeleteContainerInPod(ctx context.Context, client *containerd.Client, podName string, podNamespace string, ifForce bool) error {
	walker := &ContainerWalker{
		Client: client,
		OnFound: func(ctx context.Context, found Found) error {
			// fmt.Println(found.Container.ID())
			err := StopAndRmContainer(podNamespace, found.Container.ID(), ifForce)
			if err != nil {
				fmt.Println("Stop Rm Container in Walker Failed at line 252 ", err.Error())
			}
			return nil
		},
	}

	filter := map[string]string{
		"podName": podName,
	}
	_, err := walker.Walk(ctx, filter)
	if err != nil {
		fmt.Println("Failed At DeleteContainerInPod line 263 ", err.Error())
		return err
	}
	return nil
}

func CreatePauseContainer(ctx context.Context, client *containerd.Client, namespace string, name string) (string, error) {
	img := obj_inner.Image{
		Img:           util.FirstSandbox,
		ImgPullPolicy: "Always",
	}
	_, err := image_manager.GetImage(client, &img, ctx)
	if err != nil {
		fmt.Println("Pull Pause Image Failed At CreatePauseContainer line 276, ", err.Error())
		return "", err
	}

	res, err := util.RunContainer(namespace, name)
	if err != nil {
		return "", err
	}
	fmt.Println("res is :", res)
	return res, nil
}

func ReStartPauseContainer(namespace string, id string) error {
	res, err := util.StartContainer(namespace, id)
	if err != nil {
		return err
	}
	fmt.Println("Restart res is ", res)
	return nil
}

func DeletePauseContainer(namespace string, id string) error {
	_, _ = util.StopContainer(namespace, id)
	_, _ = util.RemoveContainer(namespace, id)
	return nil
}

// MonitorContainerState 监控容器的状态
func MonitorContainerState(ctx context.Context, container containerd.Container) (string, uint32, error) {
	// 获取容器的任务
	task, err := container.Task(ctx, nil)
	if err != nil {
		fmt.Println("Failed to get task:", err)
		return "", 0, err
	}

	// 获取任务的状态
	status, err := task.Status(ctx)
	if err != nil {
		fmt.Println("Failed to get task status:", err)
		return "", 0, err
	}
	fmt.Printf("Pod id is %s, status is %v\n", container.ID(), status)

	return string(status.Status), status.ExitStatus, nil
}

type metricsCollection struct {
	begin      time.Time
	task       containerd.Task
	lastTime   time.Time
	lastCPU    uint64
	CPUPercent uint64
	memory     uint64
}

func GetContainersMetrics(ctx context.Context, c containerd.Container, window time.Duration) (*api_obj.ContainerMetrics, error) {
	task, err := c.Task(ctx, nil)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	collection := &metricsCollection{
		begin:      time.Now(),
		task:       task,
		lastTime:   time.Now(),
		lastCPU:    uint64(0),
		CPUPercent: uint64(0),
		memory:     uint64(0),
	}

	cm := CollectContainerMetrics(ctx, collection, window, c)

	return cm, nil
}

func CollectContainerMetrics(ctx context.Context, collection *metricsCollection, window time.Duration, c containerd.Container) *api_obj.ContainerMetrics {
	var currentMetrics v2.Metrics
	var temporaryInterface interface{}
	var curTime time.Time
	var allCPUUsage uint64
	collection.begin = time.Now()
	task := collection.task
	metrics, err := task.Metrics(ctx)
	if err != nil {
		fmt.Println("CollectContainerMetrics Failed At line 353 ", err.Error())
		return nil
	}
	curTime = time.Now()
	temporaryInterface = reflect.New(reflect.TypeOf(currentMetrics)).Interface()
	err = proto.Unmarshal(metrics.Data.Value, temporaryInterface.(proto.Message))
	// fmt.Println(temporaryInterface)
	if err != nil {
		fmt.Println(err.Error())
	}
	switch value := temporaryInterface.(type) {
	case *v2.Metrics:
		currentMetrics = *value
	default:
		return nil
	}
	collection.lastTime = curTime
	collection.lastCPU = currentMetrics.CPU.UsageUsec
	collection.memory = currentMetrics.Memory.Usage
	time.Sleep(window)

	task = collection.task
	metrics, err = task.Metrics(ctx)
	if err != nil {
		fmt.Println("CollectContainerMetrics Failed At line 377 ", err.Error())
		return nil
	}
	curTime = time.Now()
	temporaryInterface = reflect.New(reflect.TypeOf(currentMetrics)).Interface()
	err = proto.Unmarshal(metrics.Data.Value, temporaryInterface.(proto.Message))
	if err != nil {
		fmt.Println(err.Error())
	}
	switch value := temporaryInterface.(type) {
	case *v2.Metrics:
		currentMetrics = *value
	default:
		return nil
	}
	collection.memory += currentMetrics.Memory.Usage
	collection.memory /= 2

	allCPUUsage = currentMetrics.CPU.UsageUsec
	cpuNow := allCPUUsage - collection.lastCPU

	timeDelta := curTime.Sub(collection.lastTime)
	collection.CPUPercent = uint64(float64(cpuNow) / float64(timeDelta.Nanoseconds()) * 1000)
	task1 := collection.task
	memPercent := 100 * float64(collection.memory) / float64(currentMetrics.Memory.UsageLimit)
	var cm api_obj.ContainerMetrics
	if task1.ID() == c.ID() {
		cm = api_obj.ContainerMetrics{
			Name: task.ID(),
			Usage: api_obj.ResourceList{
				CPUPercent:    collection.CPUPercent,
				MemoryUsage:   collection.memory,
				MemoryPercent: memPercent,
			},
		}
	}
	// fmt.Println("container matrics is ", cm.Usage)
	return &cm
}
