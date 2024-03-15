#!/bin/sh

set -x

multipass delete central
multipass delete streamer1
multipass delete streamer2
multipass purge
rm -f k3s.yaml
