# Container-based vFW application

## Summary

This is the containerized version of the composite-firewall where virtlet
VMs are replaced with Docker containers.

## Deployment

Please follow [the guide][1] of the composite-firewall app to deploy the
application. Please note that all path parameters need to be replaced
accordingly. Since virtlet is not required anymore, it can be removed from
the Vagrantfile before constructing edge clusters using kud.

It is necessary to enable hugepages for the app. If libvirt is used as the
vagrant provider, it is neccessary to modify the config file for that. Below
is the example of the necessary modifications before deployment.

```diff
diff --git a/kud/hosting_providers/vagrant/Vagrantfile b/kud/hosting_providers/vagrant/Vagrantfile
--- a/kud/hosting_providers/vagrant/Vagrantfile
+++ b/kud/hosting_providers/vagrant/Vagrantfile
@@ -27,7 +27,7 @@ File.open(File.dirname(__FILE__) + "/inventory/hosts.ini", "w") do |inventory_fi
   nodes.each do |node|
     inventory_file.puts("#{node['name']}\tansible_ssh_host=#{node['ip']} ansible_ssh_port=22")
   end
-  ['kube-master', 'kube-node', 'etcd', 'ovn-central', 'ovn-controller', 'virtlet', 'cmk'].each do|group|
+  ['kube-master', 'kube-node', 'etcd', 'ovn-central', 'ovn-controller'].each do|group|
     inventory_file.puts("\n[#{group}]")
     nodes.each do |node|
       if node['roles'].include?("#{group}")
@@ -74,6 +74,7 @@ Vagrant.configure("2") do |config|
     v.cpu_mode = 'host-passthrough'
     v.management_network_address = "192.168.121.0/27"
     v.random_hostname = true
+    v.memorybacking :hugepages
   end

   sync_type = "virtualbox"

diff --git a/kud/hosting_providers/vagrant/installer.sh b/kud/hosting_providers/vagrant/installer.sh
--- a/kud/hosting_providers/vagrant/installer.sh
+++ b/kud/hosting_providers/vagrant/installer.sh
@@ -159,14 +159,14 @@ function install_addons {
     ansible-playbook $verbose -i $kud_inventory -e "base_dest=$HOME" $kud_playbooks/configure-kud.yml | sudo tee $log_folder/setup-kud.log
     # The order of KUD_ADDONS is important: some plugins (sriov, qat)
     # require nfd to be enabled.
-    for addon in ${KUD_ADDONS:-topology-manager virtlet ovn4nfv nfd sriov qat optane cmk}; do
+    for addon in ${KUD_ADDONS:-ovn4nfv}; do
         echo "Deploying $addon using configure-$addon.yml playbook.."
         ansible-playbook $verbose -i $kud_inventory -e "base_dest=$HOME" $kud_playbooks/configure-${addon}.yml | sudo tee $log_folder/setup-${
     done
     echo "Run the test cases if testing_enabled is set to true."
     if [[ "${testing_enabled}" == "true" ]]; then
         failed_kud_tests=""
-        for addon in ${KUD_ADDONS:-multus topology-manager virtlet ovn4nfv nfd sriov qat optane cmk}; do
+        for addon in ${KUD_ADDONS:-multus ovn4nfv }; do
             pushd $kud_tests
             bash ${addon}.sh || failed_kud_tests="${failed_kud_tests} ${addon}"
             popd

```

If virtualbox is used as the vagrant provider, hugepages can be enabled
inside the vagrant VM. It's can be done by running:
$ sudo sysctl -w vm.nr_hugepages=1024

or adding hugepages to sysctl for the permament effect.
$ echo "vm.nr_hugepages=1024" | sudo tee -a /etc/sysctl.conf

Also, please note that in case of virtualbox, it is necessary to make sure
the versions of VirtualBox installed in the host and VBoxGuestAdditions installed
in the vagrant VM through vagrant-vbguest are the same. The app has been tested
with VirtualBox v6.1.18 & v6.1.22

[1]: ../../tests/README-composite-vfw.md
[2]: https://github.com/onap/multicloud-k8s/blob/master/kud/hosting_providers/vagrant/Vagrantfile
