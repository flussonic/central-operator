
# central-operator

## Description

The central-operator deploys central's instances and performs monitoring and management of them.

The functions of the operator include:

- Deploying instances of Central
- Performing database migrations
- Detecting Flussonics in the cluster and automatically adding them to Central

## Getting Started

**Steps to launch a cluster with multiple instances of Central and Flussonic.**

Apply the Central and Media Server operators so that Kubernetes recognizes the CRDs and runs operator instances waiting for the corresponding resources to appear in the cluster:

```sh
kubectl apply -f https://flussonic.github.io/media-server-operator/latest/operator.yaml
kubectl apply -f https://flussonic.github.io/central-operator/latest/operator.yaml
```

Next, label the nodes accordingly and add the necessary secrets:

```sh
kubectl label nodes node_name flussonic.com/streamer=true
kubectl create secret generic flussonic-license \
    --from-literal=license_key="${LICENSE_KEY}" \
    --from-literal=edit_auth="root:password"
```

\* node_name - the name of the node where Flussonic should run. Flussonic instances will be deployed on each node labeled accordingly.

Then, add the required custom resources so that the operators can deploy and provision the corresponding standard Kubernetes resources:

```sh
kubectl apply -f https://raw.githubusercontent.com/flussonic/central-operator/master/config/samples/ingress.yaml
kubectl apply -f https://raw.githubusercontent.com/flussonic/central-operator/master/config/samples/media_v1alpha1_mediaserver.yaml
kubectl apply -f https://raw.githubusercontent.com/flussonic/central-operator/master/config/samples/media_v1alpha1_central.yaml
kubectl apply -f https://raw.githubusercontent.com/flussonic/central-operator/master/config/samples/postgres.yaml
kubectl apply -f https://raw.githubusercontent.com/flussonic/central-operator/master/config/samples/redis.yaml
```

Note: Currently, Central works correctly only with Nginx as the ingress controller (required by the agent). If another controller is used by default, remove it from Kubernetes and apply the ingress-nginx manifests:

```sh
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/cloud/deploy.yaml
kubectl apply -f https://raw.githubusercontent.com/flussonic/central-operator/master/config/samples/nginx_ingress_class.yaml
```

If all the above steps are completed successfully, the cluster should include PostgreSQL, Redis, two Central instances, and one Flussonic instance per each labeled node.

## For developers

The project was generated using [operator-sdk](https://sdk.operatorframework.io/).

To generate the documentation, update the `version` in the Makefile to the current one and run:

```sh
make operator.yaml
```

To launch the cluster on multipass and k3s, run:

```sh
make mp-start
```

The script is configured to build the controller locally and transfer it to the nodes.

If you need to run the cluster using the controller image from Docker Hub, execute `mp-start.sh` directly.

## License

Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
