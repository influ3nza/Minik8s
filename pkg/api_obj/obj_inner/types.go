package obj_inner

type Quantity string

type ObjectMeta struct {
	Name        string            `json:"name" yaml:"name"`
	NameSpace   string            `json:"nameSpace" yaml:"nameSpace"`
	Labels      map[string]string `json:"labels" yaml:"labels"`
	Annotations map[string]string `json:"annotations" yaml:"annotations"`
	UUID        string            //todo 是否需要UUID，Name唯一吗，唯一的话需要检测冲突，不唯一则记录UUID作为标识
	// answer is yes, 4 ppt give it @k8s-2.pptx 18. But how to generate?
}

type Image struct {
	Img           string `json:"imgName" yaml:"imgName"`
	ImgPullPolicy string `json:"imgPullPolicy" yaml:"imgPullPolicy"`
}

type EntryPoint struct {
	Command    []string `json:"command" yaml:"command"`
	Args       []string `json:"args" yaml:"args"`
	WorkingDir string   `json:"workingDir" yaml:"workingDir"`
}

type EnvVar struct {
	Name  string `json:"name" yaml:"name"`
	Value string `json:"value" yaml:"value"`
}

type Volume struct {
	Name string `json:"name" yaml:"name"`
	Type string `json:"type" yaml:"type"`
	Path string `json:"path" yaml:"path"`
}

type VolumeMount struct {
	MountPath string `json:"mountPath" yaml:"mountPath"`
	SubPath   string `json:"subPath" yaml:"subPath"`
	Name      string `json:"name" yaml:"name"`
	ReadOnly  bool   `json:"readOnly" yaml:"readOnly"`
}

type ResourceRequirements struct {
	Limits   map[string]Quantity `json:"limits" yaml:"limits"`
	Requests map[string]Quantity `json:"requests" yaml:"requests"`
}

type ContainerPort struct {
	ContainerPort int32  `json:"containerPort" yaml:"containerPort"`
	HostIP        string `json:"hostIP" yaml:"hostIP"`
	HostPort      int32  `json:"hostPort" yaml:"hostPort"`
	Name          string `json:"name" yaml:"name"`
	Protocol      string `json:"protocol" yaml:"protocol"`
}

const (
	Pending     = "PodPending"
	Running     = "PodRunning"
	Succeeded   = "PodSucceeded"
	Failed      = "PodFailed"
	Unknown     = "PodUnknown"
	Terminating = "PodTerminating"
)
