SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

echo "Parent directory of the script is: $SCRIPT_DIR"

flanneld_pid=$(pgrep -f flanneld)

if [ -z "$flanneld_pid" ]; then
  echo "flanneld进程未运行"
else
  # 杀死flanneld进程
  systemctl stop flannel
  echo "已杀死flanneld进程 (PID: $flanneld_pid)"
fi

$SCRIPT_DIR/deletcd.sh

etcdctl --endpoints "http://127.0.0.1:2379" put /coreos.com/network/config '{"NetWork":"10.2.0.0/16","SubnetMin":"10.2.1.0","SubnetMax": "10.2.254.0","Backend": {"Type": "vxlan"}}'

systemctl restart flannel