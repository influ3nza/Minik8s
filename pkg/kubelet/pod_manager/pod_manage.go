package pod_manager

import (
	"fmt"
	"minik8s/pkg/api_obj"
	"minik8s/pkg/api_obj/obj_inner"
	"minik8s/pkg/kubelet/container_manager"
	"os/exec"
	"strings"
	"time"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/namespaces"
	"github.com/containerd/containerd/oci"
	"github.com/opencontainers/runtime-spec/specs-go"
	"golang.org/x/net/context"
)

// AddPod todo 需要Master上有DNS服务器
/*
 * 参数
 *  pod: *api_obj.Pod pod对象
 *
 * 返回
 *  error: 错误信息
 */
func AddPod(pod *api_obj.Pod) error {
	client, err := containerd.New("/run/containerd/containerd.sock")
	if err != nil {
		fmt.Println("Create Client Failed At line 22")
		return err
	}
	defer client.Close()
	ctx := namespaces.WithNamespace(context.Background(), pod.MetaData.NameSpace)

	if pod.MetaData.Labels["func"] != "" {
		fmt.Println(*pod)
		fmt.Println("function ", pod.MetaData.Labels["func"], " Create")
		id, err := container_manager.StartFuncContainer(client, pod.MetaData.NameSpace, pod.Spec.Containers[0].Image.Img, pod.MetaData.Name)
		if err != nil {
			return fmt.Errorf("start Func Pod Failed %s", err.Error())
		}
		pod.MetaData.Annotations["pause"] = id
		fmt.Println("create pod id is ", id)
		ip, err := GetPodIp(pod.MetaData.NameSpace, id)
		if err != nil {
			return fmt.Errorf("get Pod ip Failed %s", err.Error())
		}

		pod.PodStatus.PodIP = ip
		pod.PodStatus.Phase = obj_inner.Running
		return nil
	}

	res, err := container_manager.CreatePauseContainer(ctx, client, pod.MetaData.NameSpace,
		/*fmt.Sprintf("%s-pause", pod.MetaData.Name)*/ pod.MetaData.Name)
	if err != nil {
		fmt.Println("Create Pause Failed At line 30")
		return err
	}
	containerPauseId := strings.TrimSuffix(res, "\n")
	// 截取后面64个字母
	if len(containerPauseId) > 64 {
		containerPauseId = containerPauseId[len(containerPauseId)-64:]
	}
	fmt.Println("Create Pause At AddPod line 38 ", containerPauseId)

	files, err := GetPodNetConfFile(pod.MetaData.NameSpace, containerPauseId)
	if err != nil {
		fmt.Println("Add Pod Failed At line 42", err.Error())
		container_manager.DeletePauseContainer(pod.MetaData.NameSpace, containerPauseId)
		return err
	}
	defer RmLocalFile(files)
	fmt.Println(files)

	pid, err := GetPodPid(pod.MetaData.NameSpace, containerPauseId)
	if err != nil {
		fmt.Println("Add Pod Failed At line 51 ", err.Error())
		container_manager.DeletePauseContainer(pod.MetaData.NameSpace, containerPauseId)
		return err
	}

	ns := fmt.Sprintf("/proc/%d/ns/", pid)
	for _, container := range pod.Spec.Containers {
		startRes, podId, err_ := container_manager.CreateK8sContainer(ctx, client, &container, pod.MetaData.Name, pod.Spec.Volumes, ns)
		if err_ != nil {
			fmt.Println("Add Pod Failed At Line 60 ", err_.Error())
			container_manager.DeletePauseContainer(pod.MetaData.NameSpace, containerPauseId)
			return err_
		}

		pid_, err_ := container_manager.StartContainer(ctx, startRes)
		if err_ != nil || pid_ == 0 {
			fmt.Println("Add Pod Failed At Line 67 ")
			container_manager.DeletePauseContainer(pod.MetaData.NameSpace, containerPauseId)
			return err_
		}

		err_ = GenPodNetConfFile(pod.MetaData.NameSpace, podId, containerPauseId)
		if err_ != nil {
			fmt.Println("Add Pod Failed At line 74", err_.Error())
			container_manager.DeletePauseContainer(pod.MetaData.NameSpace, containerPauseId)
			return err_
		}
	}
	podIpInPause, err := GetPodIp(pod.MetaData.NameSpace, fmt.Sprintf("%s-pause", pod.MetaData.Name))
	fmt.Println("create pod success!")
	if err != nil {
		fmt.Println("Add Pod Failed At line 82 ", err.Error())
		container_manager.DeletePauseContainer(pod.MetaData.NameSpace, containerPauseId)
		return err
	}
	pod.PodStatus.PodIP = podIpInPause
	pod.PodStatus.Phase = obj_inner.Running
	pod.MetaData.Annotations["pause"] = containerPauseId

	return nil
}

// DeletePod 删除Pod
/*
 * 参数
 *  podName: string pod名称
 *  namespace: string 命名空间
 *  pauseId: string pause容器id
 *
 * 返回
 *  error: 错误信息
 */
func DeletePod(podName string, namespace string, pauseId string) error {
	client, err := containerd.New("/run/containerd/containerd.sock")
	ctx := namespaces.WithNamespace(context.Background(), namespace)
	if err != nil {
		fmt.Println("DeletePod Create Client Failed At line 97 ", err.Error())
		return err
	}
	defer client.Close()
	err = container_manager.DeleteContainerInPod(ctx, client, podName, namespace, false)
	if err != nil {
		fmt.Println("DeletePod DeleteContainer Failed At line 103 ", err.Error())
		return err
	}

	err = delCNIRules(pauseId)
	if err != nil {
		fmt.Println("Delete Pod CNI Rules Failed At line 109, ", err.Error())
		return nil
	}
	return nil
}

// StopPod 停止Pod
/*
 * 参数
 *  podName: string pod名称
 *  namespace: string 命名空间
 *
 * 返回
 *  error: 错误信息
 */
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

// ReStartPod 重启Pod
/*
 * 参数
 *  podName: string pod名称
 *  namespace: string 命名空间
 *
 * 返回
 *  error: 错误信息
 */
func ReStartPod(podName string, namespace string) error {
	ctx := namespaces.WithNamespace(context.Background(), namespace)
	client, err := containerd.New("/run/containerd/containerd.sock")
	if err != nil {
		fmt.Println("ReStartPod Create Client Failed At line 135 ", err.Error())
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
	fmt.Println("Restart Pause At restartPod line 150 ", res)
	if err_ != nil {
		fmt.Println("Restart Pause Failed At line 152")
		return err
	}
	containerPauseId = res[0].ID()[:12]

	files, err := GetPodNetConfFile(namespace, containerPauseId)
	if err != nil {
		fmt.Println("restart Pod Failed At line 159", err.Error())
		_ = container_manager.DeletePauseContainer(namespace, containerPauseId)
		return err
	}
	defer func(files []string) {
		_ = RmLocalFile(files)
	}(files)

	pid, err := GetPodPid(namespace, containerPauseId)
	if err != nil {
		fmt.Println("Add Pod Failed At line 169 ", err.Error())
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

				_err_ = GenPodNetConfFile(namespace, found.Container.ID(), containerPauseId)
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
		fmt.Println("Failed At DeleteContainerInPod line 250 ", err.Error())
		return err
	}

	return err
}

// MonitorPodContainers 监控指定 Pod 中的所有容器
/*
 * 参数
 *  podName: string pod名称
 *  namespace: string 命名空间
 *
 * 返回
 *  string: 容器状态
 */
func MonitorPodContainers(podName string, namespace string) string {
	client, err := containerd.New("/run/containerd/containerd.sock")
	ctx := namespaces.WithNamespace(context.Background(), namespace)
	if err != nil {
		fmt.Println("MonitorPodContainers Failed At line 262 Create Client Failed ", err.Error())
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
		fmt.Println("Failed At MonitorPodStatus line 298 ", err.Error())
		return ""
	}
	return res
}

// GetPodMetrics 获取Pod的指标
/*
 * 参数
 *  podName: string pod名称
 *  namespace: string 命名空间
 *
 * 返回
 *  *api_obj.PodMetrics: 一个PodMetrics结构
 */
func GetPodMetrics(podName string, namespace string) *api_obj.PodMetrics {
	client, err := containerd.New("/run/containerd/containerd.sock")
	ctx := namespaces.WithNamespace(context.Background(), namespace)
	if err != nil {
		fmt.Println("GetPodMetrics Failed At line 308 ", err.Error())
		return nil
	}
	defer client.Close()

	podMetric := api_obj.PodMetrics{
		Timestamp:        time.Time{},
		Window:           1 * time.Second,
		ContainerMetrics: []api_obj.ContainerMetrics{},
	}
	walker := container_manager.ContainerWalker{
		Client: client,
		OnFound: func(ctx context.Context, found container_manager.Found) error {
			res, erro := container_manager.GetContainersMetrics(ctx, found.Container, podMetric.Window)
			if erro != nil {
				fmt.Println("Get Metrics Failed At line 323 ", err.Error())
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
		fmt.Println("Get Metrics Failed At line 336 ", err.Error())
	}
	for _, containerM := range podMetric.ContainerMetrics {
		fmt.Println("container metrics is ", containerM)
	}
	fmt.Println("container number is ", count)
	return &podMetric
}

// delCNIRules 删除CNI规则
/*
 * 参数
 *  pauseId: string pause容器id
 *
 * 返回
 *  error: 错误信息
 */
func delCNIRules(pauseId string) error {
	shPath := "./tools/setup_scripts/del_flannel_net.sh"
	_, err := exec.Command(shPath, pauseId).CombinedOutput()
	if err != nil {
		return fmt.Errorf("delete Flannel Rules Failed At 344, %s", err.Error())
	}
	return nil
}
