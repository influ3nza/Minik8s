package pod_manager

import (
	"fmt"
	"github.com/containerd/containerd"
	"github.com/containerd/containerd/namespaces"
	"github.com/containerd/containerd/oci"
	"github.com/opencontainers/runtime-spec/specs-go"
	"golang.org/x/net/context"
	"minik8s/pkg/api_obj"
	"minik8s/pkg/api_obj/obj_inner"
	"minik8s/pkg/kubelet/container_manager"
	"strings"
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
	if err != nil {
		fmt.Println("Create Pause Failed At line 20")
		return err
	}
	containerPauseId = strings.TrimSuffix(res, "\n")
	// 截取后面64个字母
	if len(containerPauseId) > 64 {
		containerPauseId = containerPauseId[len(containerPauseId)-64:]
	}
	fmt.Println("Create Pause At AddPod line 23 ", containerPauseId)

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

func StopPod(podName string, namespace string) error {
	ctx := namespaces.WithNamespace(context.Background(), namespace)
	client, err := containerd.New("/run/containerd/containerd.sock")
	if err != nil {
		fmt.Println("StopPod Create Client Failed At line 104 ", err.Error())
		return err
	}
	defer client.Close()
	err = container_manager.StopContainerInPod(ctx, client, namespace, podName)
	if err != nil {
		fmt.Println("StopPod DeleteContainer Failed At line 110 ", err.Error())
		return err
	}
	return nil
}

func ReStartPod(podName string, namespace string) error {
	ctx := namespaces.WithNamespace(context.Background(), namespace)
	client, err := containerd.New("/run/containerd/containerd.sock")
	if err != nil {
		fmt.Println("ReStartPod Create Client Failed At line 122 ", err.Error())
		return err
	}
	defer client.Close()

	pauseFileter := []string{fmt.Sprintf("labels.%q==%s", "nerdctl/name", fmt.Sprintf("%s-pause", podName))}
	res, err := container_manager.ListContainers(client, ctx, pauseFileter...)
	if err != nil {
		return fmt.Errorf("cannot Find Pause Container %s", err.Error())
	} else if len(res) != 1 {
		return fmt.Errorf("pause Number Not Match %d", len(res))
	}

	err_ := container_manager.ReStartPauseContainer(namespace, res[0].ID())
	containerPauseId := ""
	fmt.Println("Restart Pause At restartPod line 137 ", res)
	if err_ != nil {
		fmt.Println("Restart Pause Failed At line 139")
		return err
	}
	containerPauseId = res[0].ID()[:12]

	files, err := GetPodNetConfFile(namespace, containerPauseId)
	if err != nil {
		fmt.Println("restart Pod Failed At line 146", err.Error())
		_ = container_manager.DeletePauseContainer(namespace, containerPauseId)
		return err
	}
	defer func(files []string) {
		_ = RmLocalFile(files)
	}(files)

	pid, err := GetPodPid(namespace, containerPauseId)
	if err != nil {
		fmt.Println("Add Pod Failed At line 161 ", err.Error())
		_ = container_manager.DeletePauseContainer(namespace, containerPauseId)
		return err
	}

	ns := fmt.Sprintf("/proc/%d/ns/", pid)
	fmt.Println("linux ns is ", pid)
	walker := &container_manager.ContainerWalker{
		Client: client,
		OnFound: func(ctx context.Context, found container_manager.Found) error {
			oldTask, _err_ := found.Container.Task(ctx, nil)
			if _err_ != nil {
				fmt.Println("Failed to get task:", _err_.Error())
				return _err_
			}

			// 获取任务的状态
			status, _err_ := oldTask.Status(ctx)
			if _err_ != nil {
				fmt.Println("Failed to get task status:", _err_)
				return _err_
			}

			fmt.Println(found.Container.ID(), " ", status)
			if status.Status == "stopped" && status.ExitStatus != 0 {
				if _, _err_ = oldTask.Delete(ctx); _err_ != nil {
					fmt.Println("Delete OldTask Failed ", _err_.Error())
				}

				var linuxNamespaces = map[string]string{
					"pid":     ns + "pid",
					"network": ns + "net",
					"ipc":     ns + "ipc",
					"uts":     ns + "uts",
				}
				var ociLinuxNsOpt []oci.SpecOpts
				for nameSpaceType, nameSpace := range linuxNamespaces {
					linuxNsSpace := specs.LinuxNamespace{
						Type: specs.LinuxNamespaceType(nameSpaceType),
						Path: nameSpace,
					}
					ociLinuxNsOpt = append(ociLinuxNsOpt, oci.WithLinuxNamespace(linuxNsSpace))
				}

				var processArgs []string
				processArgs = append(processArgs, "/bin/bash")

				entryOci := []oci.SpecOpts{oci.WithProcessCwd("/")}
				entryOci = append(entryOci, ociLinuxNsOpt...)
				entryOci = append(entryOci, oci.WithProcessArgs(processArgs...))
				updateOpts := []containerd.UpdateContainerOpts{
					containerd.UpdateContainerOpts(containerd.WithNewSpec(entryOci...)),
				}

				_err_ = found.Container.Update(ctx, updateOpts...)
				if _err_ != nil {
					return fmt.Errorf("update LinuxNs Error %s", _err_.Error())
				}

				_, _err_ = container_manager.StartContainer(ctx, found.Container)
				if err != nil {
					return fmt.Errorf("reStart Failed At line 465 %s", _err_.Error())
				}

				_err_ = GenPodNetConfFile(namespace, containerPauseId)
				if err != nil {
					return fmt.Errorf("reStart Failed At line 471 %s", _err_.Error())
				}

				return nil
			} else {
				return nil
			}
		},
	}

	filter := map[string]string{
		"podName": podName,
	}
	_, err = walker.WalkRestart(ctx, filter)
	if err != nil {
		fmt.Println("Failed At DeleteContainerInPod line 263 ", err.Error())
		return err
	}

	return err
}

// MonitorPodContainers 监控指定 Pod 中的所有容器
func MonitorPodContainers(podName string, namespace string) string {
	client, err := containerd.New("/run/containerd/containerd.sock")
	ctx := namespaces.WithNamespace(context.Background(), namespace)
	if err != nil {
		fmt.Println("MonitorPodContainers Failed At line 121 Create Client Failed ", err.Error())
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
				} else if u == 143 {
					return fmt.Errorf(obj_inner.Terminating)
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
		fmt.Println("Failed At MonitorPodStatus line 157 ", err.Error())
		return ""
	}
	return res
}

func GetPodMetrics(podName string, namespace string) *api_obj.PodMetrics {
	client, err := containerd.New("/run/containerd/containerd.sock")
	ctx := namespaces.WithNamespace(context.Background(), namespace)
	if err != nil {
		fmt.Println("GetPodMetrics Failed At line 167 ", err.Error())
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
				fmt.Println("Get Metrics Failed At line 182 ", err.Error())
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
		fmt.Println("Get Metrics Failed At line 195 ", err.Error())
	}
	for _, containerM := range podMetric.ContainerMetrics {
		fmt.Println("container metrics is ", containerM)
	}
	fmt.Println("container number is ", count)
	return &podMetric
}
