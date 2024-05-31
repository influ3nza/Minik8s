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

var pod = api_obj.Pod{
	ApiVersion: "1",
	Kind:       "pod",
	MetaData: obj_inner.ObjectMeta{
		Name:      "testpod",
		NameSpace: "test2",
		Labels: map[string]string{
			"testlabel": "podlabel",
		},
		Annotations: map[string]string{},
		UUID:        "",
	},
	Spec: api_obj.PodSpec{
		Containers: []api_obj.Container{
			/*{
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
						Name:  "MYSQL_ROOT_PASSWORD",
						Value: "123456",
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
						obj_inner.MEMORY_LIMIT: obj_inner.Quantity("500MiB"),
					},
					Requests: map[string]obj_inner.Quantity{
						obj_inner.CPU_REQUEST:    obj_inner.Quantity("0.25"),
						obj_inner.MEMORY_REQUEST: obj_inner.Quantity("100MiB"),
					},
				},
			},*/{
				Name: "testName1",
				Image: obj_inner.Image{
					Img:           "docker.io/library/ubuntu:latest",
					ImgPullPolicy: "Always",
				},
				EntryPoint: obj_inner.EntryPoint{
					WorkingDir: "/",
					Command:    []string{"tail"},
					Args:       []string{"-f", "/dev/null"},
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
						Name:  "MYSQL_ROOT_PASSWORD",
						Value: "123456",
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
						obj_inner.CPU_REQUEST:    obj_inner.Quantity("2"),
						obj_inner.MEMORY_REQUEST: obj_inner.Quantity("1GiB"),
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

func main() {
	testCreateMonitor()
	// testFunc()
}

func testFunc() {
	pod1 := &api_obj.Pod{
		ApiVersion: "v1",
		Kind:       "pod",
		MetaData: obj_inner.ObjectMeta{
			Name:      "qwer",
			NameSpace: "q",
			Labels: map[string]string{
				"func": "bar",
			},
			Annotations: map[string]string{},
		},
		Spec: api_obj.PodSpec{
			Containers: []api_obj.Container{
				{
					Name: "container",
					Image: obj_inner.Image{
						Img:           "my-registry.io:5000/bar",
						ImgPullPolicy: "Always",
					},
				},
			},
		},
		PodStatus: api_obj.PodStatus{},
	}

	res := pod_manager.AddPod(pod1)
	if res != nil {
		fmt.Println("Create Func Pod Failed ", res.Error())
	}
	fmt.Println(pod1.PodStatus.PodIP + "  " + pod1.MetaData.Annotations["pause"])
	client, err := containerd.New("/run/containerd/containerd.sock")
	if err != nil {
		fmt.Println("Create Client Failed At Main")
	}
	ctx := namespaces.WithNamespace(context.Background(), pod1.MetaData.NameSpace)
	// filter := "labels.podName==" + pod.MetaData.Name
	containers, err := container_manager.ListContainers(client, ctx)
	// containers[0].Labels(ctx)
	if len(containers) > 0 {
		str, _ := containers[0].Labels(ctx)
		fmt.Println(str)
	}
	if err != nil {
		return
	}
	fmt.Printf("container length is %d", len(containers))
	for i := 0; i < 1; i++ {
		time.Sleep(2 * time.Second)
		str := pod_manager.MonitorPodContainers(pod1.MetaData.Name, pod1.MetaData.NameSpace)
		fmt.Println("Monitor Pod is ", str)
	}

	err = pod_manager.DeletePod(pod1.MetaData.Name, pod1.MetaData.NameSpace, pod1.MetaData.Annotations["pause"])
	if err != nil {
		fmt.Println("Delete Func Pod Failed")
	}
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
}

func testCreateMonitor() {
	client, err := containerd.New("/run/containerd/containerd.sock")
	defer client.Close()
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
	fmt.Println("Register success, ", pod.MetaData.Labels["pause"])
	str := make(chan string)
	//wg.Add(2)
	//go func() {
	//	for {
	//		if util.Lock(pod.MetaData.Name, pod.MetaData.NameSpace) {
	//			res := pod_manager.MonitorPodContainers(pod.MetaData.Name, pod.MetaData.NameSpace)
	//			// fmt.Println("Monitor Pod is ", res)
	//			if strings.Contains(res, "stopped") {
	//				pod_manager.DeletePod(pod.MetaData.Name, pod.MetaData.NameSpace, pod.MetaData.Annotations["pause"])
	//			}
	//			util.UnLock(pod.MetaData.Name, pod.MetaData.NameSpace)
	//		} else {
	//			break
	//		}
	//		time.Sleep(2 * time.Second)
	//	}
	//	// wg.Done()
	//}()

	go func() {
		// time.Sleep(120 * time.Second)
		for {
			pod_manager.GetPodMetrics(pod.MetaData.Name, pod.MetaData.NameSpace)
			time.Sleep(5 * time.Second)
		}

		//if res != nil {
		//	id1 := res.ContainerMetrics[0].Name
		//	force, err_ := util.RmForce(pod.MetaData.NameSpace, id1)
		//	if err_ != nil {
		//		fmt.Println("Force Err ", force)
		//		return
		//	}
		//}
		//for {
		//	if ok := util.UnRegisterPod(pod.MetaData.Name, pod.MetaData.NameSpace); ok == 0 {
		//		fmt.Println("UnRegister Success")
		//		break
		//	} else if ok == 2 {
		//		fmt.Println("UnRegister NonExist")
		//	}
		//}
		//err = pod_manager.DeletePod(pod.MetaData.Name, pod.MetaData.NameSpace, pod.MetaData.Labels["pause"])
		//if err != nil {
		//	fmt.Println("Main Failed At line 268 ", err.Error())
		//}
		//str <- "finish"
		// wg.Done()
	}()
	// wg.Wait()
	<-str
}
