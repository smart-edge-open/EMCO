```
Copyright (c) 2021 Intel Corporation
```

# Getting Started with HPA tests
This document describes how to efficiently get started with HPA and run a simple test given as part of the HPA source. The steps are below :

- Make sure that you can running the EMCO Microservices. After the successful installation of EMCO, the following things should be running. For details on EMCO installation refer the EMCO installation docs.

```
NAME                                         READY   STATUS    RESTARTS   AGE
emco-db-emco-mongo-0                         1/1     Running   0          46h
emco-etcd-0                                  1/1     Running   0          46h
emco-services-clm-6c8f87d668-mwwjv           1/1     Running   0          46h
emco-services-dcm-7f7df86679-f6j87           1/1     Running   0          46h
emco-services-dtc-9899c7446-cbvcn            1/1     Running   0          46h
emco-services-gac-58f5b9d8f9-hnwr2           1/1     Running   0          46h
emco-services-hpa-ac-7875cc4c49-6bfkh        1/1     Running   0          46h
emco-services-hpa-plc-569975945f-lpsdk       1/1     Running   0          46h
emco-services-ncm-7644f66766-fxllr           1/1     Running   0          46h
emco-services-nps-84d77859b9-dbvrs           1/1     Running   0          46h
emco-services-orchestrator-8b8cbf989-scvr8   1/1     Running   0          46h
emco-services-ovnaction-596c64648c-5mrgs     1/1     Running   0          46h
emco-services-rsync-78858c74bb-lbvnf         1/1     Running   0          46h
emco-services-sds-7756957965-d5zc8           1/1     Running   0          46h
emco-services-sfc-7cdf8946f-wdbjl            1/1     Running   0          46h
emco-services-sfcclient-77f5cf5588-qdq9x     1/1     Running   0          46h
emco-tools-fluentd-0                         1/1     Running   0          46h
emco-tools-fluentd-7xlqt                     1/1     Running   2          46h
```
- Run HPA test examples using emcoctl

Navigate to ```$EMCO-HOME/src/hpa-plc/examples/emcoctl/hpa```

Inside this directory, use the emcoctl :
```
$EMCO-HOME/bin/emcoctl/emcoctl --config emco-cfg-local.yaml apply -f hpa-test-simple.yaml
```

where $EMCO-HOME is your EMCO home directory

The output shall be ::

```
NAME                               READY   STATUS    RESTARTS   AGE
r1-http-client-57d467c654-5skc6    1/1     Running   0          46h
r1-http-client-57d467c654-7vrgj    1/1     Running   0          46h
r1-http-server-799cc666c-r8hh5     1/1     Running   0          46h
```

