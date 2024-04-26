package pod_manager

import (
	"fmt"
	"github.com/containerd/containerd"
	"github.com/containerd/containerd/namespaces"
	"golang.org/x/net/context"
	"minik8s/pkg/api_obj"
	"minik8s/pkg/kubelet/container_manager"
)

// todo 需要Master上有DNS服务器
func AddPod(pod *api_obj.Pod) error {
	client, err := containerd.New("/run/containerd/containerd.sock")
	if err != nil {
		fmt.Println("Create Client Failed At line 14")
		return err
	}

	res, err := container_manager.CreatePauseContainer(pod.MetaData.NameSpace,
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
		return err
	}
	defer RmLocalFile(files)
	fmt.Println(files)

	pid, err := GetPodPid(pod.MetaData.NameSpace, containerPauseId)
	if err != nil {
		fmt.Println("Add Pod Failed At line 44 ", err.Error())
		return err
	}

	ns := fmt.Sprintf("/proc/%d/ns/", pid)
	ctx := namespaces.WithNamespace(context.Background(), pod.MetaData.NameSpace)
	for _, container := range pod.Spec.Containers {
		startRes, podId, err_ := container_manager.CreateK8sContainer(ctx, client, &container, pod.MetaData.Name, pod.Spec.Volumes, ns)
		if err_ != nil {
			fmt.Println("Add Pod Failed At Line 53 ", err_.Error())
			return err_
		}

		pid_, err_ := container_manager.StartContainer(ctx, startRes)
		if err_ != nil || pid_ == 0 {
			fmt.Println("Add Pod Failed At Line 58 ")
			return err_
		}

		err_ = GenPodNetConfFile(pod.MetaData.NameSpace, podId)
		if err_ != nil {
			fmt.Println("Add Pod Failed At line 64", err_.Error())
			return err_
		}
	}
	//fmt.Sprintf("%s-pause", pod.MetaData.Name)
	// podIp, err := GetPodIp(pod.MetaData.NameSpace, podId)
	podIp_pause, _ := GetPodIp(pod.MetaData.NameSpace, fmt.Sprintf("%s-pause", pod.MetaData.Name))
	fmt.Println("create pod success!")
	if err != nil {
		fmt.Println("Add Pod Failed At line 72 ", err.Error())
		return err
	}
	pod.PodStatus.PodIP = podIp_pause

	return nil
}
