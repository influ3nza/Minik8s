package container_manager

import (
	"SE3356/pkg/api_obj"
	"SE3356/pkg/api_obj/obj_inner"
	"errors"
	"fmt"
	"github.com/containerd/containerd/oci"
	"github.com/docker/go-units"
	"strconv"
)

func ParseResources(requirements obj_inner.ResourceRequirements) ([]oci.SpecOpts, error) {
	var opts []oci.SpecOpts
	var period = uint64(100000)
	var (
		reqCpu int64  = 0
		reqMem uint64 = 0
		limCpu int64  = 0
		limMem uint64 = 0
	)

	//parse all string to numbers
	if res, ok := requirements.Requests[obj_inner.CPU_REQUEST]; ok {
		parseRes, err := strconv.ParseFloat(string(res), 64)
		if err == nil {
			reqCpu = int64(parseRes * 100000.0)
			// opts = append(opts, oci.WithCPUCFS(quota, period))
		} else {
			return nil, fmt.Errorf("parse cpu requests error")
		}
	}

	if res, ok := requirements.Requests[obj_inner.MEMORY_REQUEST]; ok {
		parseRes, err := units.RAMInBytes(string(res))
		if err == nil {
			reqMem = uint64(parseRes)
		} else {
			return nil, fmt.Errorf("parse mem requests error")
		}
	}

	if res, ok := requirements.Limits[obj_inner.CPU_LIMIT]; ok {
		parseRes, err := strconv.ParseFloat(string(res), 64)
		if err == nil {
			limCpu = int64(parseRes * 100000.0)
		} else {
			return nil, fmt.Errorf("parse cpu limits error")
		}
	}

	if res, ok := requirements.Limits[obj_inner.MEMORY_LIMIT]; ok {
		parseRes, err := units.RAMInBytes(string(res))
		if err == nil {
			limMem = uint64(parseRes)
		} else {
			return nil, fmt.Errorf("parse mem limits error")
		}
	}

	if reqCpu == 0 && reqMem == 0 && limCpu == 0 && limMem == 0 {
		return opts, nil
	}

	if reqCpu > limCpu {
		reqCpu = limCpu
	}

	if reqMem > limMem {
		reqMem = limMem
	}

	if limCpu != 0 {
		opts = append(opts, oci.WithCPUCFS(limCpu, period))
	}
	if limMem != 0 {
		opts = append(opts, oci.WithMemoryLimit(limMem))
	}
	return opts, nil
}

func convertEnv(container *api_obj.Container) []string {
	var envs = []string{}
	if container.Env != nil && len(container.Env) > 0 {
		for _, env := range container.Env {
			str := env.Name + "=" + env.Value
			envs = append(envs, str)
		}
	}
	return envs
}

func convertMounts(podSpec *api_obj.PodSpec, container *api_obj.Container) ([]VolumeMap, error) {
	var mounts = []VolumeMap{}
	if container.VolumeMounts != nil {
		for _, volumeMount := range container.VolumeMounts {
			for _, volume := range podSpec.Volumes {
				if volumeMount.Name == volume.Name {
					mounts = append(mounts, VolumeMap{
						Host_:      volume.Path,
						Container_: volumeMount.MountPath,
						Subdir_:    volumeMount.SubPath,
						Type_:      volume.Type,
					})
				}
			}
		}
		return mounts, nil
	}
	return nil, errors.New("Convert Mounts Error")
}
