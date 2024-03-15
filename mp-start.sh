#!/bin/bash

set -ex

if [ -f env ]; then
    set -a
    source ./env
    set +a
fi

if [ -z "$LICENSE_KEY" ]; then
    read -p "Enter Flussonic license key: "  LICENSE_KEY
fi

multipass launch --name central --cpus 1 --memory 4096M --disk 5G lts
multipass launch --name streamer1 --cpus 1 --memory 4096M --disk 5G lts
multipass launch --name streamer2 --cpus 1 --memory 4096M --disk 5G lts

multipass exec central -- sudo /bin/sh -c 'curl -sfL https://get.k3s.io | sh -'

plane_ip=$(multipass info central | grep -i ip | awk '{print $2}')
token=$(multipass exec central sudo cat /var/lib/rancher/k3s/server/node-token)

multipass exec streamer1 -- sudo /bin/sh -c "curl -sfL https://get.k3s.io | K3S_URL=https://${plane_ip}:6443 K3S_TOKEN=${token} sh -"
multipass exec streamer2 -- sudo /bin/sh -c "curl -sfL https://get.k3s.io | K3S_URL=https://${plane_ip}:6443 K3S_TOKEN=${token} sh -"

rm -f k3s.yaml
multipass exec central sudo cat /etc/rancher/k3s/k3s.yaml |sed "s/127.0.0.1/${plane_ip}/" > k3s.yaml
chmod 0400 k3s.yaml
export KUBECONFIG=`pwd`/k3s.yaml

kubectl create secret generic flussonic-license \
    --from-literal=license_key="${LICENSE_KEY}" 


kubectl label nodes streamer1 flussonic.com/streamer=true
kubectl label nodes streamer2 flussonic.com/streamer=true
kubectl label nodes central flussonic.com/central=true

kubectl apply -f https://flussonic.github.io/media-server-operator/latest/operator.yaml
kubectl apply -f ./docs/latest/operator.yaml
