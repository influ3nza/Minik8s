### 答辩视频说明文档

> 本文档对于Minik8s答辩时演示的视频内容做出导览性的解释说明

#### 0-start.mp4
 - 演示支持多机Minik8s。
 - 支持Node抽象，用户可以通过kubectl将新的Node加入集群，并通过kubectl查看Node的运行情况。

#### 1-pod.mp4
 - 演示pod的创建和删除。
 - 多机使用时，pod可以被调度策略调度到不同的节点。
 - pod内容器可以通过localhost互相访问。
 - pod内不同容器共享文件。
 - 用户可以通过kubectl查看pod的运行情况。

#### 2-srv.mp4
- 演示用户可以通过kubectl创建service。
 - 演示service可以将多个pod纳入管理。
 - service支持多个pod的通信，集群内的pod可以通过虚拟ip访问其他节点上的pod提供的service。
 - 集群外的宿主机可以访问service。
 - 自己的机器上可以通过nodeport访问service。
 - 用户可以通过kubectl查看service。

#### 3-rep.mp4
 - 演示用户可以通过kubectl创建replicaset。
 - replicaset可以将多个pod平均分布到多个node上。
 - replicaset创建的pod可以被service正常分发流量。
 - 主动删除replicaset中的pod，可以自动回复数量。
 - 用户可以通过kubectl查看replicaset。

#### 4-hpafinal.mp4
 - 演示用户可以通过kubectl创建hpa。
 - 增加或减少pod负载，hpa可以自动改变replica数量。
 - 增加的pod可以被service正常流量分发，且分布在不同节点。
 - 支持监控cpu负载和memory负载。
 - 用户可以通过kubectl查看hpa。

#### 5-dns.mp4
 - 演示用户可以通过kubectl创建dns。
 - dns拥有多个路径，匹配多个service。
 - pod内部可以通过域名访问service。
 - 宿主机上可以通过域名访问service。
 - 同一域名下的不同子路径可以访问不同service。
 - dns删除后不再能访问service。

#### 6-fault.mp4
 - pod和service在启动后，关闭控制面，关闭的过程中pod均可以正常运行。
 - 重启之后用户仍然可以通过kubectl查看pod和service。
 - 重启之后仍然可以访问pod和service。

#### 7-serverless1.mp4
 - 演示用户可以通过kubectl创建function。
 - 演示function的http触发和事件绑定触发。
 - 演示function的冷启动和自动缩容（scale-to-0）。
 - 用户可以通过kubectl查看function。

#### 8-workflow.mp4
 - 演示用户可以通过kubectl创建workflow。
 - 演示workflow支持分支。
 - 演示workflow的运行和结果输出。

#### 9-complex.mp4
 - 演示复杂函数链。
 - 演示复杂函数链支持其他函数链的调用

#### 10-func-update.mp4
 - 演示用户可以通过kubectl更新函数。

#### 11-jmeter.mp4
 - 演示性能测试，支持并发量为20的函数调用请求，并可以正确扩容。

#### 12-pv-pvc.mp4
 - 演示用户可以通过kubectl创建pv，pvc。
 - 演示pv支持静态创建和动态创建。
 - 演示pv可以将文件在多台机器上共享，并且一个pod关闭时，自动与pv解除绑定。在新pod创建时，仍能访问之前的文件。

#### 13-monitor.mp4
 - 演示node的添加
 - 演示pod的添加与暴露指标
 - pod更改指标
 - pod删除后停止监控
 - grafana图形化界面查看node指标
 - 删除node 停止监控
