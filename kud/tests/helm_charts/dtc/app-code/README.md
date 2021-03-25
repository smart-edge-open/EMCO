**Steps to deploy a Go application on kubernetes with helm:**

(1) Compile the code under app-code
```
    cd kud/tests/helm_charts/dtc/app-code/http-server/
    go build http-server.go
    cd ../http-client/
    go build http-client.go
```

(2) Build the docker images
```
    cd ../
    docker build -t httptest-base-server -f Dockerfile_server .
    docker build -t httptest-base-client -f Dockerfile_client .
```

(3) Tag the images with docker registry
```
    docker tag httptest-base-server:latest <docker-registry-url>/my-custom-httptest-server:1.1
    docker tag httptest-base-client:latest <docker-registry-url>/my-custom-httptest-client:1.1
    Note: Bump up the version if you change the code
```

(4) Push these images to docker registry
```
    docker push <docker-registry-url>/my-custom-httptest-server:1.1
    docker push <docker-registry-url>/my-custom-httptest-client:1.1
```

(5) Modify the helm files (values.yaml, service.yaml and deployment.yaml) accordingly in the folder kud/tests/helm_charts/dtc/http-client and kud/tests/helm_charts/dtc/http-server
    Note: The NodePort in values.yaml is the port exposed by the service running on K8s. Also update the tag of the image to be downloaded if required

**Testing locally using helm commands:**

(6) [This is for your testing purpose to check whether the helm chart is getting deployed successfully]. Run helm install command to deploy this app on the K8s
```
    cd kud/tests/helm_charts/dtc/
    tar -czvf http-server.tgz http-server/
    server app deployed on North cluster: helm install --kubeconfig </root/.kube/config_north> http-server.tgz http-server
    tar -czvf http-client.tgz http-client/
    client app deployed on South Cluster: helm install --kubeconfig </root/.kube/config_south> http-client.tgz http-client
```

(7) To get the node IP and service ports
```
     North Cluster:
     kubectl --kubeconfig </root/.kube/config_north> get nodes --namespace <installed namespace> -o jsonpath="{.items[0].status.addresses[0].address}"
     kubectl --kubeconfig </root/.kube/config_north> get svc http-service --namespace <installed namespace> -o jsonpath="{.spec.ports[0].nodePort}"

     South Cluster:
     kubectl --kubeconfig </root/.kube/config_south> get nodes --namespace <installed namespace> -o jsonpath="{.items[0].status.addresses[0].address}"
```
(8)  Uninstall the client and server
```
    helm uninstall --kubeconfig </root/.kube/config_south> http-client.tgz
    helm uninstall --kubeconfig </root/.kube/config_north> http-server.tgz
```
**Test DTC feature using the emcoctl-tests test scripts:**

(9) Follow the readme instructions given under kud/emcoctl-tests/ and deploy DTC feature.

**Cleanup:**

(10) Delete the built images from app-code folder

```
    rm -rf app-code/http-client/http-client
    rm -rf app-code/http-server/http-server
```


