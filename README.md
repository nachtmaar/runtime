# Runtime

### Development

#### Test

```
make test
```
### Setup development environment (mac)

start a beefy minikube
```@sh
minikube start --memory=12288 --cpus=4 \\n  --kubernetes-version=v1.12.0 \\n  --vm-driver=hyperkit \\n  --disk-size=30g \\n  --extra-config=apiserver.enable-admission-plugins="LimitRanger,NamespaceExists,NamespaceLifecycle,ResourceQuota,ServiceAccount,DefaultStorageClass,MutatingAdmissionWebhook"
```

install istio
```
kubectl apply --filename https://raw.githubusercontent.com/knative/serving/v0.5.2/third_party/istio-1.0.7/istio-crds.yaml &&\ncurl -L https://raw.githubusercontent.com/knative/serving/v0.5.2/third_party/istio-1.0.7/istio.yaml \\n  | sed 's/LoadBalancer/NodePort/' \\n  | kubectl apply --filename -
```

install knative
```
kubectl apply --selector knative.dev/crd-install=true \\n--filename https://github.com/knative/serving/releases/download/v0.5.2/serving.yaml \\n--filename https://github.com/knative/build/releases/download/v0.5.0/build.yaml \\n--filename https://github.com/knative/eventing/releases/download/v0.5.0/release.yaml \\n--filename https://github.com/knative/eventing-sources/releases/download/v0.5.0/eventing-sources.yaml \\n--filename https://github.com/knative/serving/releases/download/v0.5.2/monitoring.yaml \\n--filename https://raw.githubusercontent.com/knative/serving/v0.5.2/third_party/config/build/clusterrole.yaml
```

install knative part2
```
kubectl apply --filename https://github.com/knative/serving/releases/download/v0.5.2/serving.yaml \\n--filename https://github.com/knative/build/releases/download/v0.5.0/build.yaml \\n--filename https://github.com/knative/eventing/releases/download/v0.5.0/release.yaml \\n--filename https://github.com/knative/eventing-sources/releases/download/v0.5.0/eventing-sources.yaml \\n--filename https://github.com/knative/serving/releases/download/v0.5.2/monitoring.yaml \\n--filename https://raw.githubusercontent.com/knative/serving/v0.5.2/third_party/config/build/clusterrole.yaml
```

modify config/samples/config.yaml to include your docker.io credentials (base64 encoded)
apply the configuration

`kubectl apply -f config/samples/config.yaml`

#### Install the CRD to a local Kubernetes cluster

```
make install
```
#### Build and run the manager
```
make run
```

# Create a docker image

```
make docker-build IMG=<img-name>
```

# Push the docker image to a configured container registry

```
make docker-push IMG=<img-name>
```

#### Run the examples
```
kubectl apply -f config/samples/runtime_v1alpha1_function.yaml
```

access the function
```
curl -v -H "Host: $(kubectl get ksvc sample --no-headers | awk '{print $2}')" http://$(minikube ip):$(kubectl get svc istio-ingressgateway --namespace istio-system --output 'jsonpath={.spec.ports[?(@.port==80)].nodePort}')
```
