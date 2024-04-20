package api_obj

import(
	"time"
)

type Condition string

const (
	Ready              Condition = "Ready"
	MemoryPressure     Condition = "MemoryPressure"
	DiskPressure       Condition = "DiskPressure"
	PIDPressure		   Condition = "PIDPressure"	   
	NetworkUnavailable Condition = "NetworkUnavailable"
)

type Address struct {
	HostName	string			`json:"hostname" yaml:"hostname"`
	ExternalIp	string			`json:"externalIp" yaml:"externalIp"`
	InternalIp	string			`json:"internalIp" yaml:"InternalIp"`
}

type NodeMetadata struct{
	UUID		string			`json:"uuid" yaml:"uuid"`
	Name		string			`json:"name" yaml:"name"`			
}

type NodeStatus struct{
	Addresses	Address				`json:"addresses" yaml:"addresses"`
	Condition	Condition			`json:"condition" yaml:"condition"`
	Capacity	map[string]string	`json:"capacity" yaml:"capacity"`
	Allocatable	map[string]string	`json:"allocatable" yaml:"allocatable"`
	UpdateTime	time.Time			`json:"updateTime" yaml:"updateTime"`
}


type Node struct{
	APIVersion		string			`json:"apiVersion" yaml:"apiVersion"`
	NodeMetadata	NodeMetadata	`json:"metadata" yaml:"metadata"`
	Kind			string			`json:"kind" yaml:"kind"`
	NodeStatus		NodeStatus		`json:"status" yaml:"status"`
}

func (n *Node) GetExternalIp() string {
	return n.NodeStatus.Addresses.ExternalIp
}

func (n *Node) GetInternelIp() string {
	return n.NodeStatus.Addresses.InternalIp
}

func (n *Node) GetHostName() string {
	return n.NodeStatus.Addresses.HostName
}

func (n *Node) GetAPIVersion() string {
	return n.APIVersion
}

func (n *Node) GetKind() string {
	return n.Kind
}

func (n *Node) GetUUID() string {
	return n.NodeMetadata.UUID
}

func (n *Node) GetName() string {
	return n.NodeMetadata.Name
}

func (n *Node) GetCapacity() map[string]string {
	return n.NodeStatus.Capacity
}

func (n *Node) GetAllocatable() map[string]string {
	return n.NodeStatus.Allocatable
}

func (n *Node) GetUpdateTime() time.Time {
	return n.NodeStatus.UpdateTime
}

func (n *Node) GetCondition() Condition {
	return n.NodeStatus.Condition
}



