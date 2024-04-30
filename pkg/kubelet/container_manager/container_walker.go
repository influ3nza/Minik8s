// Package container_manager Walker! this file is inspired by nerdctl, when watching pod - containers
package container_manager

import (
	"fmt"
	"github.com/containerd/containerd"
	"golang.org/x/net/context"
	"minik8s/pkg/api_obj/obj_inner"
)

func ListContainers(client *containerd.Client, ctx context.Context, filter ...string) ([]containerd.Container, error) {
	res, err := client.Containers(ctx, filter...)
	if err != nil {
		return nil, err
	}
	return res, nil
}

type Found struct {
	Container containerd.Container
	Filter    map[string]string
	Idx       int
	Count     int
}

type OnFound func(ctx context.Context, found Found) error

type ContainerWalker struct {
	Client  *containerd.Client
	OnFound OnFound
}

// Walk pod
func (walker *ContainerWalker) Walk(ctx context.Context, filter map[string]string) (int, error) {
	var filters []string
	for label, value := range filter {
		filters = append(filters, fmt.Sprintf("labels.%q==%s", label, value))
	}
	res, err := ListContainers(walker.Client, ctx, filters...)
	if err != nil {
		return -1, err
	}
	matchCount := len(res)
	for i, c := range res {
		f := Found{
			Container: c,
			Filter:    filter,
			Idx:       i,
			Count:     matchCount,
		}
		if e := walker.OnFound(ctx, f); e != nil {
			return -1, e
		}
	}
	return matchCount, nil
}

func (walker *ContainerWalker) WalkStatus(ctx context.Context, filter map[string]string) (string, error) {
	var filters []string
	for label, value := range filter {
		filters = append(filters, fmt.Sprintf("labels.%q==%s", label, value))
	}
	res, err := ListContainers(walker.Client, ctx, filters...)
	if err != nil {
		return obj_inner.Unknown, err
	}

	matchCount := len(res)
	if matchCount == 0 {
		return obj_inner.Pending, nil
	}

	founds := []Found{}
	for i, c := range res {
		f := Found{
			Container: c,
			Filter:    filter,
			Idx:       i,
			Count:     matchCount,
		}
		founds = append(founds, f)
	}

	ifFinished := false
	for _, found := range founds {
		e := walker.OnFound(ctx, found)
		if e != nil {
			if e.Error() == obj_inner.Failed {
				return obj_inner.Failed, nil
			} else if e.Error() == obj_inner.Succeeded {
				ifFinished = true
			} else if e.Error() == obj_inner.Pending {
				return obj_inner.Pending, nil
			} else {
				return obj_inner.Unknown, e
			}
		}
	}
	if ifFinished {
		return obj_inner.Succeeded, nil
	}

	return obj_inner.Running, nil
}
