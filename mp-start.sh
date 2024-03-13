#!/bin/bash

set -ex

if [ -f env ]; then
    set -a
    source ./env
    set +a
fi

multipass launch --name central --cpus 1 --memory 4096M --disk 5G lts
multipass exec central -- sudo /bin/sh -c 'curl -sfL https://get.k3s.io | sh -'

rm -f k3s.yaml
plane_ip=$(multipass info central | grep -i ip | awk '{print $2}')

multipass exec central sudo cat /etc/rancher/k3s/k3s.yaml |sed "s/127.0.0.1/${plane_ip}/" > k3s.yaml
chmod 0400 k3s.yaml
export KUBECONFIG=`pwd`/k3s.yaml
