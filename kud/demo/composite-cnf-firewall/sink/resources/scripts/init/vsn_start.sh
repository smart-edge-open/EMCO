#!/bin/bash

apt-get update
apt-get install -y sudo curl net-tools iproute2 inetutils-ping wget darkstat unzip

echo "provision interfaces"

echo "add route entries"
ip route add $unprotectedPrivateNetCidr via $vfwProtectedPrivateNetIp

echo "update darkstat configuration"
sed -i "s/START_DARKSTAT=.*/START_DARKSTAT=yes/g;s/INTERFACE=.*/INTERFACE=\"-i eth1\"/g" /etc/darkstat/init.cfg

echo "start darkstat"

darkstat -i eth1

echo "done"
sleep infinity
