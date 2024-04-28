package pod_manager

import (
	"fmt"
	"github.com/containerd/containerd"
	"github.com/containerd/containerd/namespaces"
	"golang.org/x/net/context"
	"minik8s/pkg/api_obj"
	"minik8s/pkg/kubelet/container_manager"
)

// AddPod todo 需要Master上有DNS服务器
func AddPod(pod *api_obj.Pod) error {
	client, err := containerd.New("/run/containerd/containerd.sock")
	if err != nil {
		fmt.Println("Create Client Failed At line 14")
		return err
	}
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

	//defer func() {
	//	container_manager.DeletePauseContainer(pod.MetaData.NameSpace,
	//		fmt.Sprintf("%s-pause", pod.MetaData.Name))
	//}()

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
	//fmt.Sprintf("%s-pause", pod.MetaData.Name)
	// podIp, err := GetPodIp(pod.MetaData.NameSpace, podId)
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
	err = container_manager.DeleteContainerInPod(ctx, client, podName, namespace, false)
	if err != nil {
		fmt.Println("DeletePod DeleteContainer Failed At line 91 ", err.Error())
		return err
	}
	return nil
}
