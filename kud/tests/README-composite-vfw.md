# Notes on running the composite vFW test case
There are two versions of the composite vFW test case:
1 - ../demo/composite-firewal: this demo uses virtlet VMs
2 - ../demo/composite-cnf-firewall: this demo uses containers

# Infrastructure
As written, the vfw-test.sh script assumes 3 clusters
1 - the cluster in which the EMCO microservices are running
2 - two edge clusters in which the vFW will be instantiated

The edge cluster in which vFW will be instantiated should be KUD clusters.

## Containerized test case
If the containerized test case is the target, virtlet is not required anymore.
It can be removed from the Vagrantfile before constructing edge clusters.

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

# Edge cluster preparation

For status monitoring support, the 'monitor' docker image must be built and
deployed.

In multicloud-k8s repo:
	cd multicloud-k8s/src/monitor
 	docker build -f build/Dockerfile . -t monitor
	<tag and push docker image to dockerhub ...>

Deploy monitor program in each cluster (assumes multicloud-k8s repo is present in cloud)
	# one time setup per cluster - install the CRD
	cd multicloud-k8s/src/monitor/deploy
	kubectl apply -f crds/k8splugin_v1alpha1_resourcebundlestate_crd.yaml
	
	# one time setup per cluster
	# update yaml files with correct image
	# (cleanup first, if monitor was already installed - see monitor-cleanup.sh)
	cd multicloud-k8s/src/monitor/deploy
	monitor-deploy.sh


# Preparation of the vFW Composit Application

## Prepare the Composite vFW Application Charts and Profiles

The following steps are for the virtlet-based test case. If the containerized version
is the target, the folder should be changed to "EMCO/kud/demo/composite-cnf-firewall"

1. In the EMCO/kud/demo/composite-firewall directory, prepare the 3 helm
   charts for the vfw.

   tar cvf packetgen.tar packetgen
   tar cvf firewall.tar firewall
   tar cvf sink.tar sink
   gzip *.tar

2. Prepare the profile file (same one will be used for all profiles in this demo)

   tar cvf profile.tar manifest.yaml override_values.yaml
   gzip profile.tar

## Set up environment variables for the vfw-test.sh script

The vfw-test.sh script expects a number of files to be provided via environment
variables.

Change directory to EMCO/kud/tests

1.  Edge cluster kubeconfig files - the script expects 2 of these

    export kubeconfigfile=<path to first cluster kube config file>
    export kubeconfigfile2=<path to second cluster kube config file>

    for example:  export kubeconfigfile=/home/vagrant/multicloud-k8s/cluster-configs/config-edge01


2.  Composite app helm chart files (as prepared above)

    export packetgen_helm_path=../demo/composite-firewall/packetgen.tar.gz
    export firewall_helm_path=../demo/composite-firewall/firewall.tar.gz
    export sink_helm_path=../demo/composite-firewall/sink.tar.gz

3.  Composite profile application profiles (as prepared above)

    export packetgen_profile_targz=../demo/composite-firewall/profile.tar.gz
    export firewall_profile_targz=../demo/composite-firewall/profile.tar.gz
    export sink_profile_targz=../demo/composite-firewall/profile.tar.gz

4.  Modify the script to address the EMCO cluster

    Modifiy the urls at the top part of the script to point to the
    cluster IP address of the EMCO cluster.

    That is, modify the IP address 10.10.10.6 to the correct value for
    your environment.

    Note also that the node ports used in the following are based on the values
    defined in multicloud-k8s/deployments/kubernetes/onap4k8s.yaml

        base_url_clm=${base_url_clm:-"http://10.10.10.6:31856/v2"}
        base_url_ncm=${base_url_ncm:-"http://10.10.10.6:32737/v2"}
        base_url_orchestrator=${base_url_orchestrator:-"http://10.10.10.6:31298/v2"}
        base_url_ovnaction=${base_url_ovnaction:-"http://10.10.10.6:31181/v2"}


# Run the vfw-test.sh

The rest of the data needed for the test is present in the script.

1.  Invoke API calls to create the data
    
    vfw-test.sh create

    This does all of the data setup
    - registers clusters
    - registers controllers
    - sets up the composite app and profile
    - sets up all of the intents

2.  Query results (optional)

    vfw-test.sh get

3.  Apply the network intents

    For the vFW test, the 3 networks used by the vFW are created by using network intents.
    Both virtual and provider networks are used.

    vfw-test.sh apply

    On the edge clusters, check to see the networks were created:

    kubectl get network
    kubectl get providernetwork

4.  Instantiate the vFW

    vfw-test.sh instantiate

    This will instantiate the vFW on the two edge clusters (as defined by the generic
    placement intent).

5. Status query

   vfw-test.sh status

6. Terminate
   Terminate will remove the resources from the clusters and delete the internal
   composite application information in the etcd base AppContext.
   The script will do it for both the deployment intent group (i.e. the vfW composite
   app) and the network intents.

   In principle, after runnin terminate, the 'apply' and 'instantiate' commands could
   be invoked again to re-insantiate the networks and the vFW composite app.

   vfw-test.sh terminate

7. Delete the data
   After running 'terminate', the 'delete' command can be invoked to remove all
   the data created.  This should leave the system back in the starting state -
   begin with point #1 above to start again.

   vfw-test.sh delete
