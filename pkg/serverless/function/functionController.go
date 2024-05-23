package function

import (
	"encoding/json"
	"fmt"
	"minik8s/pkg/api_obj"
	"minik8s/pkg/api_obj/obj_inner"
	"minik8s/pkg/config/apiserver"
	"minik8s/pkg/network"
)

type FunctionController struct {
}

func (fc *FunctionController) GenerateFunction(f *api_obj.Function) error {
	err := CreateImage(f)
	if err != nil {
		return fmt.Errorf("create Img Failed At GF, %s", err.Error())
	}

	err = fc.GenerateReplicaset(f)
	if err != nil {
		return fmt.Errorf("gen replica Failed At GF, %s", err.Error())
	}

	//todo record function

	return nil
}

func CreateNewFunctionControllerInstance() FunctionController {
	return FunctionController{}
}

func (fc *FunctionController) GetFunctionPodIps(f *api_obj.Function) {
	
}

func (fc *FunctionController) GenerateReplicaset(f *api_obj.Function) error {
	imgName := serverDns + ":5000/" + f.Metadata.Name + ":latest"
	replica := &api_obj.ReplicaSet{
		ApiVersion: "v1",
		Kind:       "replicaset",
		MetaData: obj_inner.ObjectMeta{
			Name:      "todo" + f.Metadata.Name,
			NameSpace: f.Metadata.NameSpace,
			Labels:    map[string]string{},
		},
		Spec: api_obj.ReplicaSetSpec{
			Replicas: 0,
			Selector: map[string]string{},
			Template: api_obj.PodTemplate{
				Metadata: obj_inner.ObjectMeta{
					Name:      f.Metadata.Name, //todo need to be different
					NameSpace: f.Metadata.NameSpace,
					Labels:    map[string]string{},
				},
				Spec: api_obj.PodSpec{
					Containers: []api_obj.Container{
						{
							Name: f.Metadata.Name,
							Image: obj_inner.Image{
								Img:           imgName,
								ImgPullPolicy: "Always",
							},
						}, // one and only one container
					},
					Volumes:       nil,
					NodeName:      "",
					NodeSelector:  nil,
					RestartPolicy: "",
				},
			},
		},
		Status: api_obj.ReplicaSetStatus{
			Replicas:      0,
			ReadyReplicas: 0,
			Status:        "",
			Message:       "",
		},
	}

	replicaJson, err := json.Marshal(replica)
	if err != nil {
		return fmt.Errorf("marshal Replica Failed, %s", err.Error())
	}
	_, err = network.PostRequest("", replicaJson)
	if err != nil {
		return fmt.Errorf("network Failed, %s", err.Error())
	}

	return nil
}

func (fc *FunctionController) DeleteFunction(f *api_obj.Function) error {
	replicName := "todo" + f.Metadata.Name
	url := apiserver.API_server_prefix + apiserver.API_delete_replicaset_prefix + replicName
	_, err := network.DelRequest(url)
	if err != nil {
		return fmt.Errorf("send Delete Rep Failed, %s", err.Error())
	}

	//todo record

	err = DeleteImage(f.Metadata.Name)
	if err != nil {
		fmt.Printf("no Such Img or Delete Img error, %s", err.Error())
	}
	return nil
}
