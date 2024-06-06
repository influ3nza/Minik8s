package container_manager

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"minik8s/pkg/api_obj"
	"minik8s/pkg/api_obj/obj_inner"
	"strconv"

	"github.com/containerd/containerd/oci"
	"github.com/docker/go-units"
)

// ParseResources parses the resource requirements of a container
/*
 * 参数
 *  requirements: obj_inner.ResourceRequirements
 *
 * 返回
 *  []oci.SpecOpts: 一个oci.SpecOpts类型的切片
 *  error: 错误信息
 */
func ParseResources(requirements obj_inner.ResourceRequirements) ([]oci.SpecOpts, error) {
	var opts []oci.SpecOpts
	var period = uint64(100000)
	var (
		reqCpu int64  = 0
		reqMem uint64 = 0
		limCpu int64  = 0
		limMem uint64 = 0
	)
	fmt.Println(requirements)
	//parse all string to numbers
	if res, ok := requirements.Requests[obj_inner.CPU_REQUEST]; ok {
		parseRes, err := strconv.ParseFloat(string(res), 64)
		if err == nil {
			reqCpu = int64(parseRes * 100000.0)
			// fmt.Println(reqCpu)
			// opts = append(opts, oci.WithCPUCFS(quota, period))
		} else {
			return nil, fmt.Errorf("parse cpu requests error")
		}
	}

	if res, ok := requirements.Requests[obj_inner.MEMORY_REQUEST]; ok {
		parseRes, err := units.RAMInBytes(string(res))
		if err == nil {
			reqMem = uint64(parseRes)
			// fmt.Println(reqMem)
		} else {
			return nil, fmt.Errorf("parse mem requests error")
		}
	}

	if res, ok := requirements.Limits[obj_inner.CPU_LIMIT]; ok {
		parseRes, err := strconv.ParseFloat(string(res), 64)
		if err == nil {
			limCpu = int64(parseRes * 100000.0)
			// fmt.Println(limCpu)
		} else {
			return nil, fmt.Errorf("parse cpu limits error")
		}
	}

	if res, ok := requirements.Limits[obj_inner.MEMORY_LIMIT]; ok {
		parseRes, err := units.RAMInBytes(string(res))
		if err == nil {
			limMem = uint64(parseRes)
			// fmt.Println(reqMem)
		} else {
			return nil, fmt.Errorf("parse mem limits error")
		}
	}

	// fmt.Printf("req is mem %d, cpu %d, lim is mem %d, cpu %d", reqMem, reqCpu, limMem, limCpu)
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
	println("set opt is :", len(opts))
	return opts, nil
}

// convertEnv converts the environment variables of a container
/*
 * 参数
 *  container: *api_obj.Container
 *
 * 返回
 *  []string: 一个字符串切片
 */
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

// convertMounts converts the volume mounts of a container
/*
 * 参数
 *  volumes: []obj_inner.Volume
 *  container: *api_obj.Container
 *
 * 返回
 *  []VolumeMap: 一个VolumeMap切片
 *  error: 错误信息
 */
func convertMounts(volumes []obj_inner.Volume, container *api_obj.Container) ([]VolumeMap, error) {
	var mounts []VolumeMap
	if container.VolumeMounts != nil {
		for _, volumeMount := range container.VolumeMounts {
			for _, volume := range volumes {
				if volumeMount.Name == volume.Name {
					mounts = append(mounts, VolumeMap{
						Host_:      volume.Path,           // /var/lib
						Container_: volumeMount.MountPath, // /home
						Subdir_:    volumeMount.SubPath,   // /config
						Type_:      volume.Type,
					})
				}
			}
		}
		return mounts, nil
	}
	return nil, errors.New("convert Mounts Error")
}

// GenerateUUIDForContainer generates a UUID for a container
/*
 * 返回
 *  string: 一个字符串
 *  error: 错误信息
 */
func GenerateUUIDForContainer() (string, error) {
	bytesLength := IDLength / 2
	b := make([]byte, bytesLength)
	n, err := rand.Read(b)
	if err != nil {
		fmt.Println("Failed At GenerateUUIDForContainer line 120 ", err.Error())
		return "", err
	}
	if n != bytesLength {
		fmt.Printf("expected %d bytes, got %d bytes\n", bytesLength, n)
		return "", nil
	}
	return hex.EncodeToString(b), nil
}
