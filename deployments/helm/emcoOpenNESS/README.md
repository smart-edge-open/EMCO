#################################################################
# EMCO v2 helm charts
#################################################################

EMCO Helm charts include charts for EMCO microservices along with MongoDb, etcd, Fluentd

**NOTE: Make sure to add the correct repository name in common/values.yaml**
For Ex: repository: amr-registry-pre.caas.intel.com/emco/
Make sure to add "/" in the end while defining the repository 

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

Create namespace "emco"

`$  kubectl create ns emco `

Helm install

`$  helm install --namespace emco emco-db ./dist/packages/emco-db-0.1.0.tgz `

`$  helm install --namespace emco emco-services ./dist/packages/emco-services-0.1.0.tgz `

**3. Deploy tools (Optional)**

`$ helm install --namespace emco emco-tools ./dist/packages/emco-tools-0.1.0.tgz `

NOTE: Deploy the Chart emco-0.1.0.tgz to deploy all packages including database, services and tools.

`$ helm install --namespace emco emco ./dist/packages/emco-0.1.0.tgz `

**4. Database Authentication**

By default, user name / password authentication is enabled for both the MongoDB and etcd data stores.
If no parameters are provided, the helm installation will select arbitrary passwords, which are stored
in the secrets: `emco-etcd` and `emco-mongo`

The passwords may be overridden on install by setting the following parameters:

- `global.db.rootPassword` overrides the MongoDB root password (required when authentication is used)
- `global.db.emcoPassword` overrides the MongoDB EMCO user password
- `global.contextdb.rootPassword` overrides the etcd root password

To disable database authentication, install with `global.disableDbAuth=true`

Examples:

The first two examples show how to override and disable `etcd` and `mongo` persistence.  See the *Known Issues* section
at the end of the document for more information on how to handle issues with database authentication and
database persistence enabled.

Install with database authentication enabled and default (random) passwords, and disable persistence:

`$ helm install --namespace emco --set mongo.persistence.enabled=false --set etcd.persistence.enabled=false emco ./dist/packages/emco-0.1.0.tgz`

Install with database and services separately with database authentication enabled and default (random) passwords, and disable persistence:

`$ helm install --namespace emco --set mongo.persistence.enabled=false --set etcd.persistence.enabled=false emco-db ./dist/packages/emco-db-0.1.0.tgz`
`$ helm install --namespace emco emco-services ./dist/packages/emco-services-0.1.0.tgz`

Install with database authentication enabled and default (random) passwords:

`$ helm install --namespace emco emco ./dist/packages/emco-0.1.0.tgz`

Install with database authentication enabled and override passwords:

`$ helm install --namespace emco --set global.db.rootPassword=abc --set global.db.emcoPassword=def --set global.contextdb.rootPassword=xyz emco ./dist/packages/emco-0.1.0.tgz`

Install databases and services separately with database authentication enabled and override passwords:

`$ helm install --namespace emco --set global.db.rootPassword=abc --set global.db.emcoPassword=def --set global.contextdb.rootPassword=xyz emco-db ./dist/packages/emco-db-0.1.0.tgz`

`$ helm install --namespace emco emco-services ./dist/packages/emco-services-0.1.0.tgz`

Install with database authentication disabled:

`$ helm install --namespace emco --set global.disableDbAuth=true emco ./dist/packages/emco-0.1.0.tgz`

**5. To check logs of the different Microservices check fluentd logs**

`kubectl logs emco-tools-fluentd-0 -n emco | grep orchestrator`


**6. Delete all packages**

Delete namespace "emco"

`$  kubectl delete ns emco `

Helm uninstall

`$ helm delete emco-services --purge`

`$ helm delete emco-db --purge`

Optional if tools were installed

`$ helm delete emco-tools --purge`

NOTE: If the Chart emco-0.1.0.tgz was deployed

`$ helm delete emco --purge`

### Known Issues

#### Errors after deleting the database package

After deleting the db package and before installing the package again following error happens:

`Error: release emco-db failed: object is being deleted: persistentvolumes "emco-db-emco-etcd-data-0" already exists`

Workarounds:

* remove the  finalizers section using `kubectl edit persistentvolumes emco-db-emco-etcd-data-0`
* or, if appropriate, delete the entire namespace using `kubectl delete namespace emco`

#### Known issues with database authentication

If EMCO has been installed with database authentication disabled (such as with a pre-release version of the 21.03 release) or
with authentication enabled, and then EMCO is uninstalled and subsequently reinstalled with a different database
authentication configuration (e.g. enabled and/or different passwords), it is possible that either the database services
will not become ready or the EMCO services will not authenticate with the databases properly.

In the case of `etcd`, the pod may not become fully ready and logs indicate authentication errors.
In the case of 'mongo', the mongo pods will come up, but EMCO services logs indicate that the EMCO services are having an authentication error with
the mongo database.

Workarounds:

- The quickest and simplest workaround is to uninstall EMCO and remove the host database data directories in the volume mount points.  Then reinstall.
  Using the default values in the EMCO helm charts, these will be as follows if db package was installed as helm release `emco-db`:
    - `sudo rm -r /dockerdata-nfs/emco-db/emco/mongo/data`
    - `sudo rm -r /dockerdata-nfs/emco-db/emco/etcd/data-0`
    - Or, if the db package was installed with the all-in-one emco-0.1.0.tgz chart as helm release `emco`
    - `sudo rm -r /dockerdata-nfs/emco/emco/mongo/data`
    - `sudo rm -r /dockerdata-nfs/emco/emco/etcd/data-0`
- On a re-installation, use the same database authentication passwords that were used in the previous installation.  These can be found from the secrets
  `emco-mongo` and `emco-etcd` (before the previous installation is uninstalled)
- Disable database persistence on installation:
    - For example: `$ helm install --namespace emco --set mongo.persistence.enabled=false --set etcd.persistence.enabled=false emco-db ./dist/packages/emco-db-0.1.0.tgz`
