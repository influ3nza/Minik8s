package application

import (
	"encoding/json"
	"fmt"
	"minik8s/pkg/api_obj"
	"minik8s/pkg/api_obj/obj_inner"
	"minik8s/pkg/config/apiserver"
	"minik8s/pkg/kubelet/pod_manager"
	"minik8s/pkg/kubelet/util"
	"minik8s/pkg/message"
	"minik8s/pkg/network"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
)

func (server *Kubelet) AddPod(c *gin.Context) {
	pod := &api_obj.Pod{}
	err := c.ShouldBind(pod)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "[kubelet/AddPod] Failed to parse pod",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data": "[kubelet/AddPod] creating pod",
	})

	go func() {
		util.RegisterPod(pod.MetaData.Name, pod.MetaData.NameSpace)
		if !util.Lock(pod.MetaData.Name, pod.MetaData.NameSpace) {
			errPod := api_obj.Pod{
				ApiVersion: "",
				Kind:       "",
				MetaData:   obj_inner.ObjectMeta{},
				Spec:       api_obj.PodSpec{},
				PodStatus: api_obj.PodStatus{
					PodIP: "error",
				},
			}
			errPodJson, _ := json.Marshal(errPod)
			msg := &message.Message{
				Type:    message.POD_CREATE,
				Content: string(errPodJson),
				Backup:  "This happened because of add not finished but deleted",
				Backup2: "",
			}
			server.Producer.Produce(message.TOPIC_ApiServer_FromNode, msg)
		}
		err_ := pod_manager.AddPod(pod)
		util.UnLock(pod.MetaData.Name, pod.MetaData.NameSpace)
		fmt.Println("Pod Pause Id is ", pod.MetaData.Annotations["pause"])
		if err_ != nil {
			pod.PodStatus.PodIP = "error"
			msg := &message.Message{
				Type: message.POD_CREATE,
			}
			errPodJson, err_ := json.Marshal(*pod)
			if err_ != nil {
				errPod := api_obj.Pod{
					ApiVersion: "",
					Kind:       "",
					MetaData:   obj_inner.ObjectMeta{},
					Spec:       api_obj.PodSpec{},
					PodStatus: api_obj.PodStatus{
						PodIP: "error",
					},
				}
				errPodJson, _ = json.Marshal(errPod)
				msg.Content = string(errPodJson)
			} else {
				msg.Content = string(errPodJson)
			}
			server.Producer.Produce(message.TOPIC_ApiServer_FromNode, msg)
			for {
				if ok := util.UnRegisterPod(pod.MetaData.Name, pod.MetaData.NameSpace); ok == 0 || ok == 2 {
					break
				}
			}
			return
		} else {
			msgPod, _ := json.Marshal(*pod)
			msg := &message.Message{
				Type:    message.POD_CREATE,
				Content: string(msgPod),
			}
			server.Producer.Produce(message.TOPIC_ApiServer_FromNode, msg)
			return
		}
	}()
	return
}

func (server *Kubelet) DelPod(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")
	pauseId := c.Param("pause")
	c.JSON(http.StatusOK, gin.H{
		"data": "[kubectl/DelPod] deleting pod",
	})
	key := util.GenerateKey(name, namespace)

	go func() {
		var ok = 1
		for {
			if ok = util.UnRegisterPod(name, namespace); ok == 0 || ok == 2 {
				break
			}
		}
		if _, l := util.RestartingLock.Load(key); l {
			util.RestartingLock.Delete(l)
		}
		msg := &message.Message{
			Type: message.POD_DELETE,
		}

		if ok == 2 {
			msg.Content = message.DEL_POD_NOT_EXIST
			server.Producer.Produce(message.TOPIC_ApiServer_FromNode, msg)
			return
		}

		err := pod_manager.DeletePod(name, namespace, pauseId)
		if err != nil {
			msg.Content = message.DEL_POD_FAILED
			server.Producer.Produce(message.TOPIC_ApiServer_FromNode, msg)
			return
		}

		msg.Content = message.DEL_POD_SUCCESS
		server.Producer.Produce(message.TOPIC_ApiServer_FromNode, msg)
	}()

	return
}

func (server *Kubelet) GetPodMatrix(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")

	ok := util.RLock(name, namespace)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "[kubectl/GetPodMatrix] Pod Not Exist",
		})
		return
	}

	res := pod_manager.GetPodMetrics(name, namespace)
	matrix, err := json.Marshal(*res)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "[kubelet/GetPodMatrix] Marshal Error",
		})
		util.RUnLock(name, namespace)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": string(matrix),
	})
	util.RUnLock(name, namespace)
	return
}

func (server *Kubelet) GetPodStatus() {
	nodename, _ := os.Hostname()
	for {
		time.Sleep(5 * time.Second)
		request, err := network.GetRequest(server.ApiServerAddress + apiserver.API_get_pods_by_node_prefix + nodename)
		if err != nil {
			fmt.Println("Send Get RequestErr ", err.Error())
			continue
		}

		list := gjson.Parse(request).Array()
		for _, p := range list {
			pod := &api_obj.Pod{}
			err = json.Unmarshal([]byte(p.String()), pod)
			if err != nil {
				fmt.Println("Unmarshal Error At GetPodStatus line 179 ", err.Error())
				continue
			}
			if !util.RLock(pod.MetaData.Name, pod.MetaData.NameSpace) {
				fmt.Println("Pod not Exist ", "Name is : ", pod.MetaData.Name, " Ns is : ", pod.MetaData.NameSpace)
				continue
			}

			key := util.GenerateKey(pod.MetaData.Name, pod.MetaData.NameSpace)
			if _, ok := util.RestartingLock.Load(key); ok {
				continue
			}

			res := pod_manager.MonitorPodContainers(pod.MetaData.Name, pod.MetaData.NameSpace)
			if res != obj_inner.Running {
				switch {
				case res == obj_inner.Succeeded || res == obj_inner.Unknown:
					{
						pod.PodStatus.Phase = res
						podJson, _ := json.Marshal(*pod)
						msg := &message.Message{
							Type:    message.POD_UPDATE,
							Content: string(podJson),
							Backup:  "",
							Backup2: "",
						}
						server.Producer.Produce(message.TOPIC_ApiServer_FromNode, msg)
					}
				case res == obj_inner.Failed || res == obj_inner.Terminating || res == obj_inner.Pending:
					{
						pod.PodStatus.Phase = obj_inner.Restarting
						podJson, _ := json.Marshal(pod)
						util.RestartingLock.Store(key, 1)
						go server.Restart(*pod, key)
						msg := &message.Message{
							Type:    message.POD_UPDATE,
							Content: string(podJson),
							Backup:  "",
							Backup2: "",
						}
						server.Producer.Produce(message.TOPIC_ApiServer_FromNode, msg)
					}
				}
			}
			util.RUnLock(pod.MetaData.Name, pod.MetaData.NameSpace)
		}
	}
}

func (server *Kubelet) Restart(pod_ api_obj.Pod, key_ string) {
	i := 1
	for ; i < 4; i++ {
		time.Sleep(5 * time.Second)
		if util.Lock(pod_.MetaData.Name, pod_.MetaData.NameSpace) {
			fmt.Println("Get Lock ", pod_.MetaData.NameSpace, " ", pod_.MetaData.Name)
			err_ := server.PodRestart(&pod_)
			if err_ != nil {
				fmt.Printf("Restart Failed error: %s", err_.Error())
				if i == 3 {
					pod_.PodStatus.Phase = obj_inner.Failed
					podJson_, _ := json.Marshal(pod_)
					msg_ := &message.Message{
						Type:    message.POD_UPDATE,
						Content: string(podJson_),
						Backup:  "",
						Backup2: "",
					}
					server.Producer.Produce(message.TOPIC_ApiServer_FromNode, msg_)
					util.RestartingLock.Delete(key_)
					util.UnLock(pod_.MetaData.Name, pod_.MetaData.NameSpace)
					fmt.Println("Unlock ", pod_.MetaData.NameSpace, " ", pod_.MetaData.Name)
					return
				} else {
					util.UnLock(pod_.MetaData.Name, pod_.MetaData.NameSpace)
					fmt.Println("Unlock ", pod_.MetaData.NameSpace, " ", pod_.MetaData.Name, "Continuing")
					continue
				}
			} else {
				pod_.PodStatus.Phase = obj_inner.Running
				podJson_, _ := json.Marshal(pod_)
				msg_ := &message.Message{
					Type:    message.POD_UPDATE,
					Content: string(podJson_),
					Backup:  "",
					Backup2: "",
				}
				server.Producer.Produce(message.TOPIC_ApiServer_FromNode, msg_)
				util.RestartingLock.Delete(key_)
				util.UnLock(pod_.MetaData.Name, pod_.MetaData.NameSpace)
				fmt.Println("Unlock ", pod_.MetaData.NameSpace, " ", pod_.MetaData.Name, "Restart success")
				return
			}
		} else {
			fmt.Println("Lock Failed Fall Back")
			if _, ok := util.RestartingLock.Load(key_); ok {
				util.RestartingLock.Delete(key_)
			}
			return
		}
	}
}

func (server *Kubelet) PodRestart(pod *api_obj.Pod) error {
	err := pod_manager.DeletePod(pod.MetaData.Name, pod.MetaData.NameSpace, pod.MetaData.Annotations["pause"])
	if err != nil {
		return err
	}

	err = pod_manager.AddPod(pod)
	if err != nil {
		return nil
	}
	return nil
}

// PV
func (k *Kubelet) MountNfs(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "[kubelet/MountNfs] Marshal Error.",
		})
		return
	}

	args := []string{"/mnt/minik8s/" + name}
	_, _ = exec.Command("mkdir", args...).CombinedOutput()
	args = []string{util.IpAddressMas + ":/mnt/minik8s/" + name, "/mnt/minik8s/" + name}
	_, err := exec.Command("mount", args...).CombinedOutput()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "[kubelet/MountNfs] Mount nfs failed, " + err.Error(),
		})
	}
}
