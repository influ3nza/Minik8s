### Kubelet environment

1. etcd
    ```bash
    $ sudo apt install etcd
    $ sudo systemctl enable etcd

    etcd will listen 127.0.0.1:2379 & 2380 by default
    ```
2. nerdctl
    ```bash
    $ sudo apt install containerd
    $ cd /home/$yourname/Desktop/
    $ wget https://github.com/containerd/nerdctl/releases/download/v1.7.5/nerdctl-1.7.5-linux-amd64.tar.gz
    # Extract the archive to a path like /usr/local/bin
    ```
3. network use flannel CNI
    ```bash
    $ cd /home/$yourname/Desktop/
    $ wget https://github.com/flannel-io/flannel/releases/download/v0.25.1/flannel-v0.25.1-linux-amd64.tar.gz
    $ sudo mkdir -p /opt/flannel
    $ sudo chmod 777 -R /opt/flannel
    # Extract the archive to /opt/flannel

    $ cd /home/$yourname/Desktop/
    $ wget https://github.com/containernetworking/plugins/releases/download/v0.9.1/cni-plugins-linux-amd64-v0.9.1.tgz
    $ sudo mkdir -p /opt/cni/bin
    # Extract the archive to /opt/cni/bin

    $ sudo vim /etc/cni/net.d/flannel.conflist
    ```
    insert text below 
    ```text
    {
        "name": "flannel",
        "cniVersion": "0.3.1",
        "plugins": [
            {
                "type": "flannel",
                "delegate": {
                    "isDefaultGateway": true
                }
            },
            {
                "type": "portmap",
                "capabilities": {
                    "portMappings": true
                }
            }
        ]
    }
    ```

    ```bash
    $ sudo nerdctl network ls
    NETWORK ID      NAME       FILE
                    flannel    /etc/cni/net.d/flannel.conflist
    17f29b073143    bridge     /etc/cni/net.d/nerdctl-bridge.conflist
                    host       
                    none

    $ etcdctl --endpoints "http://127.0.0.1:2379" put /coreos.com/network/config '{"NetWork":"10.2.0.0/16","SubnetMin":"10.2.1.0","SubnetMax": "10.2.20.0","Backend": {"Type": "vxlan"}}'

    $ sudo /opt/flannel/flanneld -etcd-endpoints=http://127.0.0.1:2379 & 

    # background
    ```
    now you can try
    ```bash
    $ sudo nerdctl run -d --net flannel redis
    ```