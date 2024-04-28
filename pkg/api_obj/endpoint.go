package api_obj

type Endpoints struct {
	UUID          string     `json:"UUID"`
	ServiceIp     string     `json:"ServiceIp"`
	EndpointArray []Endpoint `json:"EndpointArray"`
}

type Endpoint struct {
	//官方文档使用Cartesian Product进行ip集合和port集合的合并。
	//我们这里采用最原始的枚举方法。
	//where do you go?
	PodName string `json:"PodName"`
	PodIP   string `json:"PodIP"`
	PodPort string `json:"PodPort"`

	//where do you come from?
	ServicePort string `json:"ServicePort"`

	//if it hadn't been for Cotton Eye Joe,
	//i'd been married long time age,
	//where 'd you come from where 'd you go?
	//where 'd you come from Cotton Eye Joe?
}
