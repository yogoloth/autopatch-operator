# 1. deploy

## 1.1 build
```bash
make docker-build docker-push IMG=<some-registry>/<project-name>:tag
```
## 1.2 deploy
```bash
make install
make deploy IMG=<some-registry>/<project-name>:tag
```
## 1.3 create cr
````
# create a patch apply to pod with label "run: wangjl-test3"
kubectl apply -f config/samples/sre_v1_patch.yaml

# u can test patches with gq(https://github.com/yogoloth/gq) on pod yaml
````
example patchs:

````
config/samples/sre_v1_patch.yaml:

apiVersion: sre.wangjl.dev/v1
kind: Patch
metadata:
  name: patch-test
  labels:
    run: wangjl-test3
steps:
    - engine: jq
      patch: ".spec.containers[0].env[.spec.containers[0].env|length]={name: \"REQUEST_MEM\",valueFrom:{resourceFieldRef:{containerName: .spec.containers[0].name, divisor: 0,resource: \"requests.memory\"}}}"
    - engine: jq
      patch: ".spec.containers[0].resources.requests={cpu: \"100m\",memory: \"1Gi\"}"


````



## 1.4 create a pod with label

````
example: create a pod with label "run: wangjl-test3"
kubectl run wangjl-test3 nginx

then sre_v1_patch.yaml will patched on this pod 
````
## 1.5 undeploy
````
kubectl delete -f config/samples/sre_v1_patch.yaml
kustomize build config/crd | kubectl delete -f -
````
