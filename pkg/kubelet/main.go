package main

import (
	"context"
	"fmt"
	"github.com/containerd/containerd"
	"github.com/containerd/containerd/namespaces"
	"minik8s/pkg/api_obj"
	"minik8s/pkg/api_obj/obj_inner"
	"minik8s/pkg/kubelet/container_manager"
	"minik8s/pkg/kubelet/pod_manager"
	"minik8s/pkg/kubelet/util"
	"strconv"
	"sync"
	"time"
)

var wg sync.WaitGroup

func main() {
	testCreateMonitor()
}

func testLock() {
	lockin := func(str string) {
		util.RegisterPod(str, str)
		for {
			if util.Lock(str, str) {
				time.Sleep(1 * time.Second)
				res := util.UnLock(str, str)
				if res == false {
					fmt.Println("Lock But Unlock Error, Means Implement Wrong")
				}
			} else {
				break
			}
		}
		wg.Done()
	}

	lockout := func(str string) {
		time.Sleep(5 * time.Second)
		for {
			if res := util.UnRegisterPod(str, str); res == 0 {
				break
			} else if res == 2 {
				fmt.Println("No Such Key ", str+":"+str)
			}
		}
		wg.Done()
	}
	wg.Add(16)
	for i := 1; i < 9; i++ {
		test_ := "test" + strconv.Itoa(i)
		go lockin(test_)
		go lockout(test_)
	}
	wg.Wait()
}

func testReStart() {
	client, err := containerd.New("/run/containerd/containerd.sock")
	defer client.Close()
	pod := api_obj.Pod{
		ApiVersion: "1",
		Kind:       "pod",
		MetaData: obj_inner.ObjectMeta{
			Name:      "testpod",
			NameSpace: "test2",
			Labels: map[string]string{
				"testlabel": "podlabel",
			},
			Annotations: nil,
			UUID:        "",
		},
		Spec: api_obj.PodSpec{
			Containers: []api_obj.Container{
				{
					Name: "testubuntu",
					Image: obj_inner.Image{
						Img:           "docker.io/library/ubuntu:latest",
						ImgPullPolicy: "Always",
					},
					EntryPoint: obj_inner.EntryPoint{
						WorkingDir: "/",
					},
					Ports: []obj_inner.ContainerPort{
						{
							ContainerPort: 0,
							HostIP:        "0.0.0.0",
							HostPort:      0,
							Name:          "no name",
							Protocol:      "TCP",
						},
					},
					Env: []obj_inner.EnvVar{
						{
							Name:  "env1",
							Value: "env1Value",
						},
					},
					VolumeMounts: []obj_inner.VolumeMount{
						{
							MountPath: "/home",
							SubPath:   "config",
							Name:      "testMount",
							ReadOnly:  false,
						},
					},
					Resources: obj_inner.ResourceRequirements{
						Limits: map[string]obj_inner.Quantity{
							obj_inner.CPU_LIMIT:    obj_inner.Quantity("0.5"),
							obj_inner.MEMORY_LIMIT: obj_inner.Quantity("200MiB"),
						},
						Requests: map[string]obj_inner.Quantity{
							obj_inner.CPU_REQUEST:    obj_inner.Quantity("0.25"),
							obj_inner.MEMORY_REQUEST: obj_inner.Quantity("100MiB"),
						},
					},
				}, {
					Name: "testName1",
					Image: obj_inner.Image{
						Img:           "docker.io/library/ubuntu:latest",
						ImgPullPolicy: "Always",
					},
					EntryPoint: obj_inner.EntryPoint{
						WorkingDir: "/",
					},
					Ports: []obj_inner.ContainerPort{
						{
							ContainerPort: 0,
							HostIP:        "0.0.0.0",
							HostPort:      0,
							Name:          "no name",
							Protocol:      "TCP",
						},
					},
					Env: []obj_inner.EnvVar{
						{
							Name:  "env2",
							Value: "env2Value",
						},
					},
					VolumeMounts: []obj_inner.VolumeMount{
						{
							MountPath: "/home",
							SubPath:   "config1",
							Name:      "testMount",
							ReadOnly:  false,
						},
					},
					Resources: obj_inner.ResourceRequirements{
						Limits: map[string]obj_inner.Quantity{
							obj_inner.CPU_LIMIT:    obj_inner.Quantity("0.5"),
							obj_inner.MEMORY_LIMIT: obj_inner.Quantity("200MiB"),
						},
						Requests: map[string]obj_inner.Quantity{
							obj_inner.CPU_REQUEST:    obj_inner.Quantity("0.25"),
							obj_inner.MEMORY_REQUEST: obj_inner.Quantity("100MiB"),
						},
					},
				},
			},
			Volumes: []obj_inner.Volume{
				{
					Name: "testMount",
					Type: "",
					Path: "/testOOO",
				},
			},
			NodeName:      "",
			NodeSelector:  nil,
			RestartPolicy: "",
		},
	}
	fmt.Println(pod.Spec.Containers[0].Name)

	ctx := namespaces.WithNamespace(context.Background(), pod.MetaData.NameSpace)
	if err != nil {
		fmt.Println("Create Client Failed : ", err.Error())
	}

	containers, err := container_manager.ListContainers(client, ctx)
	if err != nil {
		fmt.Println("List Containers Failed : ", err)
	}

	if len(containers) > 0 {
		fmt.Println("There should be no containers in \"test\"")
	}

	err = pod_manager.AddPod(&pod)
	if err != nil {
		fmt.Println("Main Failed At line 197 ", err.Error())
	}
	fmt.Println("Pod Ip is ", pod.PodStatus.PodIP)

	util.RegisterPod(pod.MetaData.Name, pod.MetaData.NameSpace)
	fmt.Println("Register success")

	//err = pod_manager.StopPod(pod.MetaData.Name, pod.MetaData.NameSpace)
	//if err != nil {
	//	fmt.Println("Stopped Failed At Main ", err.Error())
	//}
	//
	//for i := 0; i < 3; i++ {
	//	err = pod_manager.ReStartPod(pod.MetaData.Name, pod.MetaData.NameSpace)
	//	if err != nil {
	//		fmt.Println(err.Error())
	//		continue
	//	} else {
	//		break
	//	}
	//}
	//
	//if err != nil {
	//	_ = pod_manager.DeletePod(pod.MetaData.Name, pod.MetaData.NameSpace)
	//	fmt.Println("Restart Failed")
	//}
}

func testCreateMonitor() {
	client, err := containerd.New("/run/containerd/containerd.sock")
	defer client.Close()
	pod := api_obj.Pod{
		ApiVersion: "1",
		Kind:       "pod",
		MetaData: obj_inner.ObjectMeta{
			Name:      "testpod",
			NameSpace: "test2",
			Labels: map[string]string{
				"testlabel": "podlabel",
			},
			Annotations: nil,
			UUID:        "",
		},
		Spec: api_obj.PodSpec{
			Containers: []api_obj.Container{
				{
					Name: "testubuntu",
					Image: obj_inner.Image{
						Img:           "docker.io/library/ubuntu:latest",
						ImgPullPolicy: "Always",
					},
					EntryPoint: obj_inner.EntryPoint{
						WorkingDir: "/",
					},
					Ports: []obj_inner.ContainerPort{
						{
							ContainerPort: 0,
							HostIP:        "0.0.0.0",
							HostPort:      0,
							Name:          "no name",
							Protocol:      "TCP",
						},
					},
					Env: []obj_inner.EnvVar{
						{
							Name:  "env1",
							Value: "env1Value",
						},
					},
					VolumeMounts: []obj_inner.VolumeMount{
						{
							MountPath: "/home",
							SubPath:   "config",
							Name:      "testMount",
							ReadOnly:  false,
						},
					},
					Resources: obj_inner.ResourceRequirements{
						Limits: map[string]obj_inner.Quantity{
							obj_inner.CPU_LIMIT:    obj_inner.Quantity("0.5"),
							obj_inner.MEMORY_LIMIT: obj_inner.Quantity("200MiB"),
						},
						Requests: map[string]obj_inner.Quantity{
							obj_inner.CPU_REQUEST:    obj_inner.Quantity("0.25"),
							obj_inner.MEMORY_REQUEST: obj_inner.Quantity("100MiB"),
						},
					},
				}, {
					Name: "testName1",
					Image: obj_inner.Image{
						Img:           "docker.io/library/ubuntu:latest",
						ImgPullPolicy: "Always",
					},
					EntryPoint: obj_inner.EntryPoint{
						WorkingDir: "/",
					},
					Ports: []obj_inner.ContainerPort{
						{
							ContainerPort: 0,
							HostIP:        "0.0.0.0",
							HostPort:      0,
							Name:          "no name",
							Protocol:      "TCP",
						},
					},
					Env: []obj_inner.EnvVar{
						{
							Name:  "env2",
							Value: "env2Value",
						},
					},
					VolumeMounts: []obj_inner.VolumeMount{
						{
							MountPath: "/home",
							SubPath:   "config1",
							Name:      "testMount",
							ReadOnly:  false,
						},
					},
					Resources: obj_inner.ResourceRequirements{
						Limits: map[string]obj_inner.Quantity{
							obj_inner.CPU_LIMIT:    obj_inner.Quantity("0.5"),
							obj_inner.MEMORY_LIMIT: obj_inner.Quantity("200MiB"),
						},
						Requests: map[string]obj_inner.Quantity{
							obj_inner.CPU_REQUEST:    obj_inner.Quantity("0.25"),
							obj_inner.MEMORY_REQUEST: obj_inner.Quantity("100MiB"),
						},
					},
				},
			},
			Volumes: []obj_inner.Volume{
				{
					Name: "testMount",
					Type: "",
					Path: "/testOOO",
				},
			},
			NodeName:      "",
			NodeSelector:  nil,
			RestartPolicy: "",
		},
	}
	fmt.Println(pod.Spec.Containers[0].Name)

	ctx := namespaces.WithNamespace(context.Background(), pod.MetaData.NameSpace)
	if err != nil {
		fmt.Println("Create Client Failed : ", err.Error())
	}

	containers, err := container_manager.ListContainers(client, ctx)
	if err != nil {
		fmt.Println("List Containers Failed : ", err)
	}

	if len(containers) > 0 {
		fmt.Println("There should be no containers in \"test\"")
	}

	err = pod_manager.AddPod(&pod)
	if err != nil {
		fmt.Println("Main Failed At line 197 ", err.Error())
	}
	fmt.Println("Pod Ip is ", pod.PodStatus.PodIP)

	util.RegisterPod(pod.MetaData.Name, pod.MetaData.NameSpace)
	fmt.Println("Register success")

	wg.Add(2)
	go func() {
		for {
			if util.Lock(pod.MetaData.Name, pod.MetaData.NameSpace) {
				res := pod_manager.MonitorPodContainers(pod.MetaData.Name, pod.MetaData.NameSpace)
				fmt.Println("Monitor Pod is ", res)
				util.UnLock(pod.MetaData.Name, pod.MetaData.NameSpace)
			} else {
				break
			}
			time.Sleep(2 * time.Second)
		}
		wg.Done()
	}()

	go func() {
		//time.Sleep(4 * time.Second)
		//res := pod_manager.GetPodMetrics(pod.MetaData.Name, pod.MetaData.NameSpace)
		//if res != nil {
		//	id1 := res.ContainerMetrics[0].Name
		//	force, err_ := util.RmForce(pod.MetaData.NameSpace, id1)
		//	if err_ != nil {
		//		fmt.Println("Force Err ", force)
		//		return
		//	}
		//}
		time.Sleep(4 * time.Second)
		for {
			if ok := util.UnRegisterPod(pod.MetaData.Name, pod.MetaData.NameSpace); ok == 0 {
				fmt.Println("UnRegister Success")
				break
			} else if ok == 2 {
				fmt.Println("UnRegister NonExist")
			}
		}
		//err = pod_manager.DeletePod(pod.MetaData.Name, pod.MetaData.NameSpace)
		//if err != nil {
		//	fmt.Println("Main Failed At line 202 ", err.Error())
		//}
		wg.Done()
	}()
	wg.Wait()
}
