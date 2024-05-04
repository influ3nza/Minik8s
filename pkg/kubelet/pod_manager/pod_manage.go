package pod_manager

import (
	"fmt"
	"github.com/containerd/containerd"
	"github.com/containerd/containerd/namespaces"
	"golang.org/x/net/context"
	"minik8s/pkg/api_obj"
	"minik8s/pkg/api_obj/obj_inner"
	"minik8s/pkg/kubelet/container_manager"
	"time"
)

// AddPod todo 需要Master上有DNS服务器
func AddPod(pod *api_obj.Pod) error {
	client, err := containerd.New("/run/containerd/containerd.sock")
	if err != nil {
		fmt.Println("Create Client Failed At line 14")
		return err
	}
	defer client.Close()
	ctx := namespaces.WithNamespace(context.Background(), pod.MetaData.NameSpace)
	res, err := container_manager.CreatePauseContainer(ctx, client, pod.MetaData.NameSpace,
		/*fmt.Sprintf("%s-pause", pod.MetaData.Name)*/ pod.MetaData.Name)
	containerPauseId := ""
	fmt.Println("Create Pause At AddPod line 23 ", res)
	if err != nil {
		fmt.Println("Create Pause Failed At line 20")
		return err
	}
	containerPauseId = res[:12]

	files, err := GetPodNetConfFile(pod.MetaData.NameSpace, containerPauseId)
	if err != nil {
		fmt.Println("Add Pod Failed At line 37", err.Error())
		container_manager.DeletePauseContainer(pod.MetaData.NameSpace, containerPauseId)
		return err
	}
	defer RmLocalFile(files)
	fmt.Println(files)

	pid, err := GetPodPid(pod.MetaData.NameSpace, containerPauseId)
	if err != nil {
		fmt.Println("Add Pod Failed At line 44 ", err.Error())
		container_manager.DeletePauseContainer(pod.MetaData.NameSpace, containerPauseId)
		return err
	}

	ns := fmt.Sprintf("/proc/%d/ns/", pid)
	for _, container := range pod.Spec.Containers {
		startRes, podId, err_ := container_manager.CreateK8sContainer(ctx, client, &container, pod.MetaData.Name, pod.Spec.Volumes, ns)
		if err_ != nil {
			fmt.Println("Add Pod Failed At Line 53 ", err_.Error())
			container_manager.DeletePauseContainer(pod.MetaData.NameSpace, containerPauseId)
			return err_
		}

		pid_, err_ := container_manager.StartContainer(ctx, startRes)
		if err_ != nil || pid_ == 0 {
			fmt.Println("Add Pod Failed At Line 58 ")
			container_manager.DeletePauseContainer(pod.MetaData.NameSpace, containerPauseId)
			return err_
		}

		err_ = GenPodNetConfFile(pod.MetaData.NameSpace, podId)
		if err_ != nil {
			fmt.Println("Add Pod Failed At line 64", err_.Error())
			container_manager.DeletePauseContainer(pod.MetaData.NameSpace, containerPauseId)
			return err_
		}
	}
	podIpInPause, err := GetPodIp(pod.MetaData.NameSpace, fmt.Sprintf("%s-pause", pod.MetaData.Name))
	fmt.Println("create pod success!")
	if err != nil {
		fmt.Println("Add Pod Failed At line 72 ", err.Error())
		container_manager.DeletePauseContainer(pod.MetaData.NameSpace, containerPauseId)
		return err
	}
	pod.PodStatus.PodIP = podIpInPause

	return nil
}

func DeletePod(podName string, namespace string) error {
	client, err := containerd.New("/run/containerd/containerd.sock")
	ctx := namespaces.WithNamespace(context.Background(), namespace)
	if err != nil {
		fmt.Println("DeletePod Create Client Failed At line 85 ", err.Error())
		return err
	}
	defer client.Close()
	err = container_manager.DeleteContainerInPod(ctx, client, podName, namespace, false)
	if err != nil {
		fmt.Println("DeletePod DeleteContainer Failed At line 91 ", err.Error())
		return err
	}
	return nil
}

// MonitorPodContainers 监控指定 Pod 中的所有容器
func MonitorPodContainers(podName string, namespace string) string {
	client, err := containerd.New("/run/containerd/containerd.sock")
	ctx := namespaces.WithNamespace(context.Background(), namespace)
	if err != nil {
		fmt.Println("MonitorPodContainers Failed At line 108 Create Client Failed ", err.Error())
		return ""
	}
	defer client.Close()

	walker := container_manager.ContainerWalker{
		Client: client,
		OnFound: func(ctx context.Context, found container_manager.Found) error {
			state, u, err_ := container_manager.MonitorContainerState(ctx, found.Container)
			if err_ != nil {
				return err_
			}
			if state == "running" {
				return nil
			}
			if state == "created" {
				return fmt.Errorf(obj_inner.Pending)
			}
			if state == "stopped" {
				if u == 0 {
					return fmt.Errorf(obj_inner.Succeeded)
				} else {
					return fmt.Errorf(obj_inner.Failed)
				}
			}
			return nil
		},
	}

	filter := map[string]string{
		"podName": podName,
	}
	res, err := walker.WalkStatus(ctx, filter)
	if err != nil {
		fmt.Println("Failed At MonitorPodStatus line 129 ", err.Error())
		return ""
	}
	return res
}

func GetPodMetrics(podName string, namespace string) *api_obj.PodMetrics {
	client, err := containerd.New("/run/containerd/containerd.sock")
	ctx := namespaces.WithNamespace(context.Background(), namespace)
	if err != nil {
		fmt.Println("GetPodMetrics Failed At line 134 ", err.Error())
		return nil
	}
	defer client.Close()

	podMetric := api_obj.PodMetrics{
		Timestamp:        time.Time{},
		Window:           3 * time.Second,
		ContainerMetrics: []api_obj.ContainerMetrics{},
	}
	walker := container_manager.ContainerWalker{
		Client: client,
		OnFound: func(ctx context.Context, found container_manager.Found) error {
			res, erro := container_manager.GetContainersMetrics(ctx, found.Container, podMetric.Window)
			if erro != nil {
				fmt.Println("Get Metrics Failed At line 145 ", err.Error())
				return erro
			}
			podMetric.ContainerMetrics = append(podMetric.ContainerMetrics, *res)
			return nil
		},
	}

	filter := map[string]string{
		"podName": podName,
	}

	count, err := walker.Walk(ctx, filter)
	if err != nil {
		fmt.Println("Get Metrics Failed At line 159 ", err.Error())
	}
	for _, containerM := range podMetric.ContainerMetrics {
		fmt.Println("container metrics is ", containerM)
	}
	fmt.Println("container number is ", count)
	return &podMetric
}
