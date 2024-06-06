```bash
modprobe ip_vs
sysctl net.ipv4.vs.conntrack=1
ipvsadm -A -t 10.20.0.1:80 -s rr
ipvsadm -a -t 10.20.0.1:80 -r 10.2.20.2:80 -m
iptables -t nat -I POSTROUTING -m ipvs --vaddr 10.20.0.1 --vport 80 -j MASQUERADE
iptables -t nat -A POSTROUTING -m ipvs --vaddr 10.20.0.1 --vport 80 -j MASQUERADE
iptables -t nat -A PREROUTING -p tcp -d 192.168.1.5 --dport 80 -j DNAT --to-destination 10.20.0.1:80
iptables -t nat -A OUTPUT -d 192.168.1.5/32 -p tcp -m tcp --dport 80 -j DNAT --to-destination 10.20.0.1:80
ip addr add 10.20.0.1/24 dev flannel.1
```