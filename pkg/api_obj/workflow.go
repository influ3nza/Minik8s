package api_obj

import (
	"minik8s/pkg/api_obj/obj_inner"
)

type WorkflowNodeType string
type CompareType string

const (
	WF_Func WorkflowNodeType = "func"
	WF_Fork WorkflowNodeType = "fork"
)

const (
	WF_ByNumEqual    CompareType = "NumE"
	WF_ByNumGreater  CompareType = "NumG"
	WF_ByNumLess     CompareType = "NumL"
	WF_ByNumNotEqual CompareType = "NumNE"
	WF_ByStrEqual    CompareType = "StrE"
	WF_ByStrNotEqual CompareType = "StrNE"
)

type WF_FuncSpec struct {
	Namespace string `json:"namespace" yaml:"namespace"`
	Name      string `json:"name" yaml:"name"`
	Next      string `json:"next" yaml:"next"`
}

type WF_ForkSpec struct {
	Variable  string      `json:"variable" yaml:"variable"`
	CompareBy CompareType `json:"compareType" yaml:"compareType"`
	Next      string      `json:"next" yaml:"next"`
}

type WorkflowNode struct {
	Name      string           `json:"name" yaml:"name"`
	Type      WorkflowNodeType `json:"type" yaml:"type"`
	FuncSpec  WF_FuncSpec      `json:"funcSpec" yaml:"funcSpec"`
	ForkSpecs []WF_ForkSpec    `json:"forkSpecs" yaml:"forkSpecs"`
}

type WorkflowSpec struct {
	StartCoeff string         `json:"startCoeff" yaml:"startCoeff"`
	StartNode  string         `json:"startNode" yaml:"startNode"`
	Nodes      []WorkflowNode `json:"nodes" yaml:"nodes"`
}

type Workflow struct {
	ApiVersion string               `json:"apiVersion" yaml:"apiVersion"`
	Kind       string               `json:"kind" yaml:"kind"`
	MetaData   obj_inner.ObjectMeta `json:"metaData" yaml:"metadata"`
	Spec       WorkflowSpec         `json:"spec" yaml:"spec"`
}
