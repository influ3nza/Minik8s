./deletecd.sh

export ETCDCTL_API=3
etcdctl --endpoints "http://127.0.0.1:2379" put /coreos.com/network/config '{"NetWork":"10.2.0.0/16","SubnetMin":"10.2.1.0","SubnetMax": "10.2.20.0","Backend": {"Type": "vxlan"}}'

/opt/flannel/flanneld -etcd-endpoints=http://127.0.0.1:2379 &