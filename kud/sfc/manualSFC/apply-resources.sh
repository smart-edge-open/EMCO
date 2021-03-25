## do the following if completely resetting to updated CNI
# kubectl apply -f /home/vagrant/git/icn-nodus/deploy/ovn4nfv-k8s-plugin.yaml
kubectl apply -f namespace-left.yaml 
kubectl apply -f namespace-right.yaml 
sleep 5
kubectl apply -f sfc-virutal-network.yaml 
sleep 5
kubectl apply -f slb-multiple-network.yaml 
sleep 5
kubectl apply -f ngfw.yaml 
sleep 5
kubectl apply -f sdewan-multiple-network.yaml 
sleep 5
kubectl apply -f nginx-left-deployment.yaml
sleep 5
kubectl apply -f nginx-right-deployment.yaml
#sleep 5
#kubectl apply -f sfc-with-virtual-and-provider-network.yaml 
