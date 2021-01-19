# Wordpress-mysql Operator

## Overview

This Wordpress-mysql operator over Kuberentes. Created using the [operator-sdk](https://github.com/operator-framework/operator-sdk) project. Deploys wordpress using on sql via a [custom resource](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/).

## Prerequisites

- [docker][docker_tool] version 17.03+
- [kubectl][kubectl_tool] v1.14.1+
- [operator_sdk][operator_install]
- Access to a Kubernetes v1.14.1+ cluster, I used [minikube][minikube_install]

## Getting Started

### Cloning the repository

Checkout this Memcached Operator repository

```
git clone https://github.com/talbensimhon/redHat-operators.git
cd redhat-operators/wordpress-mysql-operator
```

### Building the operator     

Build the Wordpress-mysql operator image and push it to a public registry, such as quay.io:     

```sh
export USERNAME=<docker hub username>     
export IMG=$USERNAME/wordpress-mysql-operator:v1alpha1      
docker push IMG=$IMG      
```
The current version can be found at IMG=talbensimhon71/wordpress-mysql-operator:v1alpha1 in dockerHub 
**NOTE** To allow the cluster pull the image the repository needs to be set as public or you must configure an image pull secret.     


### Run the operator

Deploy the project to the cluster. Set `IMG` with `make deploy` to use the image you just pushed:

### Create a `Wordpress-mysql` resources

Apply the sample Custom Resource:

```sh
kubectl apply -f config/samples/wordpressyaml -n wordpress-mysql-operator-system
```

Run the following command to verify that the installation was successful:

```sh
$ kubectl get all -n wordpress-mysql-operator-system

NAME                                                               READY   STATUS    RESTARTS   AGE
pod/wordpress-mysql-operator-controller-manager-58b94bccff-f9558   2/2     Running   0          73m

NAME                                                                  TYPE        CLUSTER-IP   EXTERNAL-IP   PORT(S)    AGE
service/wordpress-mysql-operator-controller-manager-metrics-service   ClusterIP   10.97.1.22   <none>        8443/TCP   73m

NAME                                                          READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/wordpress-mysql-operator-controller-manager   1/1     1            1           73m

NAME                                                                     DESIRED   CURRENT   READY   AGE
replicaset.apps/wordpress-mysql-operator-controller-manager-58b94bccff   1         1         1       73m
```

### Cleanup

To leave the operator, but remove the wordpress sample pods, delete the
CR.

```sh
kubectl delete -f config/samples/wordpress.yaml -n wordpress-mysql-operator-system
```

To clean up everything:

```sh
make undeploy
```
### Troubleshooting

Run the following command to check the operator logs.

```sh
kubectl logs deployment.apps/wordpress-mysql-operator-controller-manager -n wordpress-mysql-operator-system -c manager
```

### Notes
I had too many issues with pushing the image.
I did docker login, but still I was unauthorized.
```sh
unauthorized: access to the requested resource is not authorized
```

For reference, I've created the CR manually by runing 
```sh
 kubectl apply -k config/samples/KubeSamples
```
Test it was done
```sh
kubectl get services
NAME              TYPE           CLUSTER-IP       EXTERNAL-IP   PORT(S)        AGE
kubernetes        ClusterIP      10.96.0.1        <none>        443/TCP        28h
wordpress         LoadBalancer   10.102.193.105   <pending>     80:30135/TCP   44s
wordpress-mysql   ClusterIP      None             <none>        3306/TCP       44s
Tals-MBP:wordpress-mysql-operator talbensimhon$ kubectl get deployments
NAME              READY   UP-TO-DATE   AVAILABLE   AGE
wordpress         1/1     1            1           69s
wordpress-mysql   1/1     1            1           69s
Tals-MBP:wordpress-mysql-operator talbensimhon$ kubectl get pvc
NAME             STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
mysql-pv-claim   Bound    pvc-d2612987-e4ea-4699-ab57-5faac37f982c   20Gi       RWO            standard       114s
wp-pv-claim      Bound    pvc-92305463-2e6a-484d-8b0a-0476bd41e886   20Gi       RWO            standard       114s
Tals-MBP:wordpress-mysql-operator talbensimhon$ kubectl get secret
NAME                    TYPE                                  DATA   AGE
default-token-r76rg     kubernetes.io/service-account-token   3      28h
mysql-pass-66f5f4tctg   Opaque                                1      2m5s
Tals-MBP:wordpress-mysql-operator talbensimhon$ 

```

The operator is not doing it yet. Still need to debug


[kubectl_tool]: https://kubernetes.io/docs/tasks/tools/install-kubectl/
[docker_tool]: https://docs.docker.com/install/
[operator_install]: https://sdk.operatorframework.io/docs/install-operator-sdk/
[minikube_install]: https://minikube.sigs.k8s.io/docs/start/
