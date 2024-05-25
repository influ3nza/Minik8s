todo：弄明白那些rsa，ssh是干什么用的

目前想法就是

1.用户通过yaml文件创建一个job，这个要在pkg/kubectl/cmd/cmd_apply.go下添加一个case，然后调用SendObjectTo函数来发送给apiserver请求，创建一个job，并保存在etcd中；

2.然后jobController去监控job的情况，然后根据对应的job的信息去创建pod，这里我想维护一个变量来判断当前job的状态（类似于未处理或者已处理），这样来避免对一个job重复创建pod；

3.（todo）接下来就是在pod里面提交cuda脚本的步骤了，怎么做还不知道，好像是job的yaml文件里面就要写cuda编译脚本命令，

基本想法是：一旦创建了job并且放置在etcd中，就创建另外一个controller来负责向交我算提交，然后调用run函数，来然后run函数内具体执行提交逻辑，提交完，在run函数内根据jobID来监控（通过循环）这个job，如果结束或者出错，就return，结束的话还要将result存储成文件然后发送到apiserver，然后apiserver持久化到etcd中（更新对应job的jobfile）

# 5.23

那我要pod干什么？

好像可以不用根据job创建pod，我直接对着etcd的job的信息，直接执行步骤3的controller
