#!/bin/bash

apt-get update
apt-get install -y sudo curl net-tools iproute2 wget

curl -s https://packagecloud.io/install/repositories/fdio/release/script.deb.sh | sudo bash

export VPP_VER=19.01.2-release
apt-get install -y vpp=$VPP_VER vpp-lib=$VPP_VER

apt-get install -y vpp-plugins=$VPP_VER

if [ -e /run/vpp/cli-vpp1.sock ]; then
    rm /run/vpp/cli-vpp1.sock
fi

#	root@vpktgen:/# taskset -p --cpu-list 1
#	pid 1's current affinity list: 1,2,29

corelist=`taskset -p -c 1 |cut -d : -f 2 | sed 's/^ *//' | sed 's/ *$//'`
#extract master core
mastercoreidx=`echo $corelist | cut -d , -f 1`
#extract worker cores
workercorelist=`echo $corelist | sed -E 's/^[0-9]*,//'`

echo 'start... vpp'
vpp unix {cli-listen /run/vpp/cli-vpp1.sock} api-segment { prefix vpp1 } \
    cpu { main-core $mastercoreidx  corelist-workers $workercorelist }

echo 'wait vpp be up ...'
while [ ! -e /run/vpp/cli-vpp1.sock ]; do
    sleep 1;
done

echo 'configure vpp ...'

HWADDR1=$(ifconfig eth1 |grep ether | tr -s ' ' | cut -d' ' -f 3)
IPADDR1=$(ifconfig eth1 |grep inet | tr -s ' ' | cut -d' ' -f 3)

vppctl -s /run/vpp/cli-vpp1.sock show ver
vppctl -s /run/vpp/cli-vpp1.sock show threads

vppctl -s /run/vpp/cli-vpp1.sock create host-interface name eth1 hw-addr $HWADDR1

vppctl -s /run/vpp/cli-vpp1.sock set int state host-eth1 up

vppctl -s /run/vpp/cli-vpp1.sock show int
vppctl -s /run/vpp/cli-vpp1.sock show hardware

vppctl -s /run/vpp/cli-vpp1.sock set int ip address host-eth1 $unprotectedPrivateNetCidr

vppctl -s /run/vpp/cli-vpp1.sock show int addr

vppctl -s /run/vpp/cli-vpp1.sock ip route add $protectedPrivateNetCidr via $vfwUnprotectedPrivateNetIp

vppctl -s /run/vpp/cli-vpp1.sock show ip fib

echo "provision streams"
### pktgen config
vppctl -s /run/vpp/cli-vpp1.sock loop create
vppctl -s /run/vpp/cli-vpp1.sock set int ip address loop0 11.22.33.1/24
vppctl -s /run/vpp/cli-vpp1.sock set int state loop0 up

cd /opt

mkdir /home/root
cat <<EOF> /home/root/stream_fw_udp1_loop0
packet-generator new {
	  name fw_udp1
	  rate 10
	  node ip4-input
	  size 64-64
	  no-recycle
      interface loop0
	  data {
		UDP: $IPADDR1 -> $vsnProtectedPrivateNetIp
		UDP: 15320 -> 8080
		length 128 checksum 0 incrementing 1
	  }
	}
EOF

vppctl -s /run/vpp/cli-vpp1.sock  exec /home/root/stream_fw_udp1_loop0

vppctl -s /run/vpp/cli-vpp1.sock packet-generator enable

vppctl -s /run/vpp/cli-vpp1.sock show packet-generator

vppctl -s /run/vpp/cli-vpp1.sock show int

echo "done"
sleep infinity
