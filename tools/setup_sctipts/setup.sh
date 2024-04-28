if [ "$EUID" -ne 0 ]
  then echo "请以 root 身份运行此脚本"
  exit
fi

if [ -x "$(command -v etcd)" ]; then
  echo "etcd 已经安装，跳过安装步骤"
else
  # 更新 apt 软件包索引
  apt update

  # 使用 apt 安装 etcd wget
  apt install -y etcd wget

  apt install -y openjdk-17-jre
  apt install -y openjdk-17-jdk
  # 输出安装完成的消息
  echo "etcd wget已成功安装"
fi
systemctl enable etcd

echo "etcd 已成功安装并已启用"

if [ -x "$(command -v containerd)" ]; then
  echo "containerd 已经安装，跳过安装步骤"
else
  apt update

  apt install -y containerd

  echo "containerd 已成功安装"
fi
systemctl enable containerd

echo "containerd 已成功安装"

if ! command -v wget &> /dev/null; then
  echo "请先安装 wget"
  exit
fi

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

mkdir -p /etc/cni/net.d/flannel.conflist
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

echo "flannel.conflist 文件已创建并写入内容"

etcdctl --endpoints "http://127.0.0.1:2379" put /coreos.com/network/config '{"NetWork":"10.2.0.0/16","SubnetMin":"10.2.1.0","SubnetMax": "10.2.20.0","Backend": {"Type": "vxlan"}}'

/opt/flannel/flanneld -etcd-endpoints=http://127.0.0.1:2379 &

#安装
wget https://archive.apache.org/dist/kafka/3.5.1/kafka_2.13-3.5.1.tgz
tar xzf kafka_2.13-3.5.1.tgz
mv kafka_2.13-3.5.1 /usr/local/kafka

# 启动
/usr/local/kafka/bin/zookeeper-server-start.sh /usr/local/kafka/config/zookeeper.properties
/usr/local/kafka/bin/kafka-server-start.sh /usr/local/kafka/config/server.properties

SCRIPTS_ROOT="$(cd "$(dirname "$0")" && pwd)"
. "$SCRIPTS_ROOT/etcd_clear.sh" /
PROJECT_ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
export MINIK8S_PATH="$PROJECT_ROOT"
echo "设置环境变量: MINIK8S_PATH=$MINIK8S_PATH"
# 关闭
# /usr/local/kafka/bin/kafka-server-stop.sh
# /usr/local/kafka/bin/zookeeper-server-stop.sh