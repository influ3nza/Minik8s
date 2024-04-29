SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

echo "Parent directory of the script is: $SCRIPT_DIR"

flanneld_pid=$(pgrep -f flanneld)

if [ -z "$flanneld_pid" ]; then
  echo "flanneld进程未运行"
else
  # 杀死flanneld进程
  sudo kill -9 "$flanneld_pid"
  echo "已杀死flanneld进程 (PID: $flanneld_pid)"
fi

$SCRIPT_DIR/deletcd.sh

export ETCDCTL_API=3
etcdctl --endpoints "http://127.0.0.1:2379" put /coreos.com/network/config '{"NetWork":"10.2.0.0/16","SubnetMin":"10.2.1.0","SubnetMax": "10.2.20.0","Backend": {"Type": "vxlan"}}'

/opt/flannel/flanneld -etcd-endpoints=http://127.0.0.1:2379 &