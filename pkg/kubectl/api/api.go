package api

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"minik8s/pkg/api_obj"
	"minik8s/pkg/api_obj/obj_inner"
	"minik8s/pkg/config/apiserver"
	"minik8s/pkg/network"
)

// WARN:此函数仅供测试使用。
func ParsePod(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("[ERR/kubectl/parsePod] Failed to open file, err:%v\n", err)
		return err
	}

	content, err := io.ReadAll(file)
	if err != nil {
		fmt.Printf("[ERR/kubectl/parsePod] Failed to read file, err:%v\n", err)
		return err
	}
	pod := &api_obj.Pod{}

	err = yaml.Unmarshal(content, pod)
	if err != nil {
		fmt.Printf("[ERR/kubectl/parsePod] Failed to unmarshal yaml, err:%v\n", err)
		return err
	}

	//WARN: 这里默认为running，便于测试。
	pod.PodStatus.Phase = obj_inner.Pending

	pod_str, err := json.Marshal(pod)
	if err != nil {
		fmt.Printf("[ERR/kubectl/parsePod] Failed to marshal pod, err:%v\n", err)
		return err
	}

	//将请求发送给apiserver
	uri := apiserver.API_server_prefix + apiserver.API_add_pod
	_, err = network.PostRequest(uri, pod_str)
	if err != nil {
		fmt.Printf("[ERR/kubectl/parsePod] Failed to post request, err:%v\n", err)
		return err
	}

	fmt.Printf("[kubectl/parsePod] Send add pod request success!\n")

	return nil
}

// WARN:此函数仅供测试使用。
func ParseNode(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("[ERR/kubectl/parseNode] Failed to open file, err:%v\n", err)
		return err
	}

	content, err := io.ReadAll(file)
	if err != nil {
		fmt.Printf("[ERR/kubectl/parseNode] Failed to read file, err:%v\n", err)
		return err
	}

	node := &api_obj.Node{}
	err = yaml.Unmarshal(content, node)
	if err != nil {
		fmt.Printf("[ERR/kubectl/parseNode] Failed to unmarshal yaml, err:%v\n", err)
		return err
	}

	node_str, err := json.Marshal(node)
	if err != nil {
		fmt.Printf("[ERR/kubectl/parseNode] Failed to marshal pod, err:%v\n", err)
		return err
	}

	//将请求发送给apiserver
	uri := apiserver.API_server_prefix + apiserver.API_add_node
	_, err = network.PostRequest(uri, node_str)
	if err != nil {
		fmt.Printf("[ERR/kubectl/parseNode] Failed to post request, err:%v\n", err)
		return err
	}

	fmt.Printf("[kubectl/parseNode] Send add node request success!\n")

	return nil
}

// WARN:此函数仅供测试使用。
func ParseSrv(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("[ERR/kubectl/parseSrv] Failed to open file, err:%v\n", err)
		return err
	}

	content, err := io.ReadAll(file)
	if err != nil {
		fmt.Printf("[ERR/kubectl/parseSrv] Failed to read file, err:%v\n", err)
		return err
	}

	srv := &api_obj.Service{}
	err = yaml.Unmarshal(content, srv)
	if err != nil {
		fmt.Printf("[ERR/kubectl/parseSrv] Failed to unmarshal yaml, err:%v\n", err)
		return err
	}

	srv_str, err := json.Marshal(srv)
	if err != nil {
		fmt.Printf("[ERR/kubectl/parseSrv] Failed to marshal pod, err:%v\n", err)
		return err
	}

	//将请求发送给apiserver
	uri := apiserver.API_server_prefix + apiserver.API_add_service
	_, err = network.PostRequest(uri, srv_str)
	if err != nil {
		fmt.Printf("[ERR/kubectl/parseSrv] Failed to post request, err:%v\n", err)
		return err
	}

	fmt.Printf("[kubectl/parseSrv] Send add srv request success!\n")

	return nil
}

func SendObjectTo(jsonStr []byte, kind string) error {
	var suffix string
	switch kind {
	case "pod":
		suffix = apiserver.API_add_pod
	case "node":
		suffix = apiserver.API_add_node
	case "service":
		suffix = apiserver.API_add_service
	case "dns":
		suffix = apiserver.API_add_dns
	}

	uri := apiserver.API_server_prefix + suffix
	_, err := network.PostRequest(uri, jsonStr)
	if err != nil {
		fmt.Printf("[ERR/kubectl/apply"+kind+"] Failed to send request, err: %s\n", err.Error())
		return err
	}

	fmt.Printf("[kubectl/apply" + kind + "] Send add " + kind + " request success!\n")

	return nil
}

// TODO:实现有问题。
func DoZip(path string, target string) error {
	// 1. Create a ZIP file and zip.Writer
	f, err := os.Create(target)
	if err != nil {
		return err
	}
	defer f.Close()

	writer := zip.NewWriter(f)
	defer writer.Close()

	// 2. Go through all the files of the source
	return filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 3. Create a local file header
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		// set compression
		header.Method = zip.Deflate

		// 4. Set relative path of a file as the header name
		header.Name, err = filepath.Rel(filepath.Dir(path), p)
		if err != nil {
			return err
		}
		if info.IsDir() {
			header.Name += "/"
		}

		// 5. Create writer for the file header and save content of the file
		headerWriter, err := writer.CreateHeader(header)
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()

		_, err = io.Copy(headerWriter, f)
		return err
	})
}

func DoUnzip(source string, destination string) error {
	// 1. Open the zip file
	reader, err := zip.OpenReader(source)
	if err != nil {
		return err
	}
	defer reader.Close()

	// 2. Get the absolute destination path
	destination, err = filepath.Abs(destination)
	if err != nil {
		return err
	}

	// 3. Iterate over zip files inside the archive and unzip each of them
	for _, f := range reader.File {
		err := unzipFile(f, destination)
		if err != nil {
			return err
		}
	}

	return nil
}

func unzipFile(f *zip.File, destination string) error {
	// 4. Check if file paths are not vulnerable to Zip Slip
	filePath := filepath.Join(destination, f.Name)
	if !strings.HasPrefix(filePath, filepath.Clean(destination)+string(os.PathSeparator)) {
		return fmt.Errorf("invalid file path: %s", filePath)
	}

	// 5. Create directory tree
	if f.FileInfo().IsDir() {
		if err := os.MkdirAll(filePath, os.ModePerm); err != nil {
			return err
		}
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
		return err
	}

	// 6. Create a destination file for unzipped content
	destinationFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
	if err != nil {
		return err
	}
	defer destinationFile.Close()

	// 7. Unzip the content of a file and copy it to the destination file
	zippedFile, err := f.Open()
	if err != nil {
		return err
	}
	defer zippedFile.Close()

	if _, err := io.Copy(destinationFile, zippedFile); err != nil {
		return err
	}
	return nil
}
