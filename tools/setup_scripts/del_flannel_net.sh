#!/bin/bash

# 检查参数是否提供
if [ $# -ne 1 ]; then
    echo "Usage: $0 <keyword>"
    exit 1
fi

keyword="$1"

# 列出当前所有的nat表规则
rules=$(iptables-save | grep "$keyword")

while IFS= read -r rule; do
    # 删除匹配的规则
    rule=$(echo "$rule" | cut -c 3-)
    srv="iptables -t nat -D $rule"
    eval "$srv"

done <<< "$rules"