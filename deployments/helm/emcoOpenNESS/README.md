#################################################################
# EMCO v2 helm charts
#################################################################

EMCO Helm charts include charts for EMCO microservices along with MongoDb, etcd, Fluentd

Make sure to add the correct repository name in values.yaml of each component with correct image:tag

### Steps to generate and install packages
**1. Run make file to package all the required chart**

`$ make clean`

`$ make all`

Pacakges helm charts in tar.gz format. All packages are in **dist/packages** directory and the package of intrest are:

   File      | Description |
  | ----------- | ----------- |
  | **emco-db-0.1.0.tgz**      | Includes database packages for mongo & etcd       |
  | **emco-services-0.1.0.tgz**   | Includes packages for all EMCO services like orchestrator, ncm, rsync etc        |
  | **emco-tools-0.1.0.tgz**   | Tools like Fluentd to be used with EMCO        |
  | **emco-0.1.0.tgz**   | Includes all charts including database, all services and tools        |


**2. Deploy EMCO Packages for Databases and Services**

`$  helm install --namespace emco emco-db./dist/packages/emco-db-0.1.0.tgz `

`$  helm install --namespace emco emco-services ./dist/packages/emco-services-0.1.0.tgz `

**3. Deploy tools (Optional)**

`$ helm install --namespace emco emco-tools ./dist/packages/emco-tools-0.1.0.tgz `

NOTE: Deploy the Chart emco-0.1.0.tgz to deploy all packages including database, services and tools.

`$ helm install --namespace emco emco ./dist/packages/emco-0.1.0.tgz `


**4. To check logs of the different Microservices check fluentd logs**

`kubectl logs emco-tools-fluentd-0 -n emco | grep orchestrator`


**5. Delete all packages**

`$ helm delete emco-services --purge`

`$ helm delete emco-db --purge`

Optional if tools were installed

`$ helm delete emco-tools --purge`

NOTE: If the Chart emco-0.1.0.tgz was deployed

`$ helm delete emco --purge`

### Known Issues

After deleting the db package and before installing the package again following error happens:

`Error: release emco-db failed: object is being deleted: persistentvolumes "emco-db-emco-etcd-data-0" already exists`

Workarounds:

* remove the  finalizers section using `kubectl edit persistentvolumes emco-db-emco-etcd-data-0`
* or, if appropriate, delete the entire namespace using `kubectl delete namespace emco`
