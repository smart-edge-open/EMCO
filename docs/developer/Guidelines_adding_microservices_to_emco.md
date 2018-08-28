```
SPDX-License-Identifier: Apache-2.0
Copyright (c) 2019-2020 Intel Corporation
```
Here are the guidelines that need to be followed whenever we add a new microservice to EMCO code.

Following are the list of files that need to be added/updated followed by thorough testing with both local and helm install.

###  Dockerfile:
• Create new docker file with name Docker.<service name> under root/build/docker folder

### Makefile : 
• Update Build, Pre-compile, compile-container, tidy, clean targets

### Docker-compose.yml:
• Look for the correct ports to be added as part of docker service and update file accordingly for new service in root/deployments/docker/docker-compose.yml file


### Config.json for new microservice:
• Avoid adding hardcoded ips inside root/src/<service>/config.json

### Helm Templates:
• Under deployments/helm/emcoCI/templates
-> Update configmap.yaml with config values from src/service/config.json file
-> Update deployment.yaml with correct ports , volume mounts, binary file details
-> Update service.yaml with correct ports 

### Scripts:

In scripts/deploy_emco.sh, add a new line to push the new service's image to the container registry:

```
 push_to_registry <service-name> ${TAG}

```
### EMCO-CFG configs:
Under root/src/tools/emcoctl/examples

* Update emco-cfg.yaml, emco-cfg-local.yaml and emco-cfg-remote.yaml for new service,host and port


* After updating the code, please use the 
  [local install tutorial](../user/Tutorial_Local_Install.md) and the
  [Helm install tutorial](../user/Tutorial_Helm.md) to build and deploy
  the new microservice.

* After installing locally , update the output of **docker ps** in the [Local Install tutorial](../user/Tutorial_Local_Install.md).

* After installing remotely using helm, update the output of **kubectl get all -n emco** in the [Helm Install tutorial](../user/Tutorial_Helm.md)
