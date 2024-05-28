package function

import (
	"fmt"
	"log"
	"minik8s/pkg/config/apiserver"
	"minik8s/pkg/network"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
)

type FileWatcher struct {
	Watcher *fsnotify.Watcher
	Mutex   sync.Mutex
}

var Watcher *FileWatcher = &FileWatcher{}

func init() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Println("[Serverless/FileWatcher] Failed To Start FileWatcher")
		return
	}
	Watcher.Watcher = watcher
	Watcher.Mutex = sync.Mutex{}
}

func (fw *FileWatcher) FileWatch() {
	done := make(chan bool)
	go func() {
		defer close(done)
		for {
			select {
			case ev := <-fw.Watcher.Events:
				{
					//判断事件发生的类型，如下5种
					// Create 创建
					// Write 写入
					// Remove 删除
					// Rename 重命名
					// Chmod 修改权限
					if ev.Op&fsnotify.Create == fsnotify.Create {
						log.Println("创建文件 : ", ev.Name)
					}
					if ev.Op&fsnotify.Write == fsnotify.Write {
						log.Println("写入文件 : ", ev.Name)
						pathParts := strings.Split(ev.Name, string(filepath.Separator))
						if len(pathParts) >= 2 {
							fName := pathParts[len(pathParts)-2]
							uri := apiserver.API_server_prefix + apiserver.API_exec_function_prefix + fName
							_, err := network.GetRequest(uri)
							if err != nil {
								fmt.Printf("[ERR/FileWatch] Failed to trigger function, %s\n", err.Error())
								return
							}
							fmt.Printf("[FileWatcher] Send exec function success.\n")
						}
					}
					if ev.Op&fsnotify.Remove == fsnotify.Remove {
						log.Println("删除文件 : ", ev.Name)
					}
					if ev.Op&fsnotify.Rename == fsnotify.Rename {
						log.Println("重命名文件 : ", ev.Name)
					}
					if ev.Op&fsnotify.Chmod == fsnotify.Chmod {
						log.Println("修改权限 : ", ev.Name)
					}
				}
			case err := <-fw.Watcher.Errors:
				{
					log.Println("error : ", err)
					return
				}
			}
		}
	}()
	<-done
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func (fw *FileWatcher) AddWatchFile(fileName string) error {
	fw.Mutex.Lock()
	defer fw.Mutex.Unlock()
	if fileExists(fileName) {
		err := fw.Watcher.Add(fileName)
		if err != nil {
			fmt.Println("[Serverless/FileWatcher] Failed To Add Watch File")
			return err
		}
		return nil
	} else {
		fmt.Println("[Serverless/FileWatcher] Failed To Add Watch File, File Not Exist")
		return fmt.Errorf("no Such File")
	}
}

func (fw *FileWatcher) DelWatchFile(fileName string) error {
	fw.Mutex.Lock()
	defer fw.Mutex.Unlock()
	err := fw.Watcher.Remove(fileName)
	if err != nil {
		fmt.Println("[Serverless/FileWatcher] Failed To Rm Watch File, ", err.Error())
		return err
	}

	return nil
}
