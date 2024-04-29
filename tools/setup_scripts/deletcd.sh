substring=/coreos.com/network

echo "Deleting etcd storage with prefix: $substring"

# 获取 etcd 中所有键值
keys=$(etcdctl get --prefix / | awk -F'[:]' '{print $1}')

# 循环删除包含指定字符串的键值
for key in $keys; do
  if [[ $key == *$substring* ]]; then
    echo  "Delete key: $key"
    etcdctl del $key
  fi
done

echo "Done!"