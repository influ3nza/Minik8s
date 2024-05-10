if [ "$EUID" -ne 0 ]
  then echo "请以 root 身份运行此脚本"
  exit
fi

# 更新 apt 软件包索引
apt update

# 使用 apt 安装 wget
apt install -y wget

wget https://github.com/etcd-io/etcd/releases/download/v3.4.32/etcd-v3.4.32-linux-amd64.tar.gz
tar -xzvf etcd-v3.4.32-linux-amd64.tar.gz
cd etcd-v3.4.32-linux-amd64
cp etcd /usr/local/bin
cp etcdctl /usr/local/bin
cd ..

apt install -y openjdk-17-jre
apt install -y openjdk-17-jdk
# 输出安装完成的消息
echo "etcd wget已成功安装"

if [ "$1" == "node" ]; then
  nohup etcd --data-dir="/var/lib/etcd/default.etcd" --listen-client-urls="http://192.168.1.13:2379,http://localhost:2379"  --advertise-client-urls="http://192.168.1.13:2379,http://localhost:2379" & > etcd.log
else
    # 参数不匹配时的处理
    etcd &
fi

echo "etcd 已成功安装"

if [ -x "$(command -v containerd)" ]; then
  echo "containerd 已经安装，跳过安装步骤"
else
  apt update

  apt install -y containerd

  echo "containerd 已成功安装"
fi
systemctl enable containerd

echo "containerd 已成功安装"

wget https://github.com/containerd/nerdctl/releases/download/v1.7.5/nerdctl-1.7.5-linux-amd64.tar.gz
if [ $? -ne 0 ]; then
  echo "下载 nerdctl 失败"
  exit
fi

tar -xzvf nerdctl-1.7.5-linux-amd64.tar.gz -C /usr/local/bin/

if [ $? -ne 0 ]; then
  echo "解压 nerdctl 失败"
  exit
fi
rm nerdctl-1.7.5-linux-amd64.tar.gz

echo "nerdctl 已成功下载并解压到 /usr/local/bin"

wget https://github.com/flannel-io/flannel/releases/download/v0.25.1/flannel-v0.25.1-linux-amd64.tar.gz

# 检查下载是否成功
if [ $? -ne 0 ]; then
  echo "下载 flannel 失败"
  exit
fi

# 创建目录 /opt/flannel
mkdir -p /opt/flannel
# 修改目录权限
chmod 777 -R /opt/flannel
# 解压 flannel 到 /opt/flannel
tar -xzvf flannel-v0.25.1-linux-amd64.tar.gz -C /opt/flannel
# 检查解压是否成功
if [ $? -ne 0 ]; then
  echo "解压 flannel 失败"
  exit
fi
# 删除下载的压缩文件
rm flannel-v0.25.1-linux-amd64.tar.gz
echo "flannel 已成功下载并解压到 /opt/flannel 目录，并授予适当的权限"

wget https://github.com/containernetworking/plugins/releases/download/v0.9.1/cni-plugins-linux-amd64-v0.9.1.tgz
if [ $? -ne 0 ]; then
  echo "下载 flannel 失败"
  exit
fi

mkdir -p /opt/cni/bin
tar -xzvf cni-plugins-linux-amd64-v0.9.1.tgz -C /opt/cni/bin

touch -p /etc/cni/net.d/flannel.conflist
cat << EOF > /etc/cni/net.d/flannel.conflist
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
EOF

nerdctl network ls
echo "flannel.conflist 文件已创建并写入内容"

export ETCDCTL_API=3
etcdctl --endpoints "http://localhost:2379" put /coreos.com/network/config '{"NetWork":"10.2.0.0/16","SubnetMin":"10.2.1.0","SubnetMax": "10.2.20.0","Backend": {"Type": "vxlan"}}'

if [ "$1" == "node" ]; then
  nohup /opt/flannel/flanneld -etcd-endpoints=http://192.168.1.13:2379 & > flannel.log
else
    # 参数不匹配时的处理
    nohup /opt/flannel/flanneld -etcd-endpoints=http://localhost:2379 &
fi

#安装
wget https://archive.apache.org/dist/kafka/3.5.1/kafka_2.13-3.5.1.tgz
tar xzf kafka_2.13-3.5.1.tgz
mv kafka_2.13-3.5.1 /usr/local/kafka

# 启动
nohup /usr/local/kafka/bin/zookeeper-server-start.sh /usr/local/kafka/config/zookeeper.properties & > zookeeper.out
nohup /usr/local/kafka/bin/kafka-server-start.sh /usr/local/kafka/config/server.properties & > kafka.out

SCRIPTS_ROOT="$(cd "$(dirname "$0")" && pwd)"
#. "$SCRIPTS_ROOT/etcd_clear.sh" /
PROJECT_ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
export MINIK8S_SCRIPTS_PATH="$SCRIPTS_ROOT"
export MINIK8S_PATH="$PROJECT_ROOT"
echo "设置环境变量: MINIK8S_PATH=$MINIK8S_PATH"
cat /etc/cni/net.d/flannel.conflist
# 关闭
# /usr/local/kafka/bin/kafka-server-stop.sh
# /usr/local/kafka/bin/zookeeper-server-stop.sh
