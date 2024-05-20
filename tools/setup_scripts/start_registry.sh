ip_address=$(hostname -I | awk '{print $1}')
input_address=$1

if [ -x "$(command -v docker)" ]; then
  echo "docker 已经安装，跳过安装步骤"
else
  apt update
  apt install -y docker.io
fi

if [ -x "$(command -v docker-compose)" ]; then
  echo "docker compose 已经安装"
else
  apt install -y docker-compose
fi

addr=$(pwd)/tools/setup_scripts/docker_compose_files/registry.yml
echo "$addr"

if [ "$ip_address" == "$input_address" ]; then
  if ! docker ps | grep -q "registry"; then
    docker-compose -f "$addr" up -d
  else
    echo "registry 已启动"
  fi
fi

if [ ! -f /etc/docker/daemon.json ]; then
    touch /etc/docker/daemon.json
fi

echo '{
    "insecure-registries": ["'"$input_address:5000"'"]
}' > /etc/docker/daemon.json

systemctl restart docker