kubectl delete -f sfc-with-virtual-and-provider-network.yaml
kubectl delete -f nginx-right-deployment.yaml
kubectl delete -f nginx-left-deployment.yaml
kubectl delete -f sdewan-multiple-network.yaml
kubectl delete -f ngfw.yaml 
kubectl delete -f slb-multiple-network.yaml 
kubectl delete -f sfc-virutal-network.yaml 
kubectl delete -f namespace-left.yaml
kubectl delete -f namespace-right.yaml