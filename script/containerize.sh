#!/bin/bash

set -e

if [[ "$1" == "" ]]; then
  echo "Usage: $0 command args"
  exit 1
fi

make base_image
# Need to have it privileged and mount the cgroups in rw mode because of https://github.com/docker/for-mac/issues/6073
docker run -e TERM=color -it --rm --privileged --cgroupns=host -v /sys/fs/cgroup:/sys/fs/cgroup:rw teleport-exec-test go run -race cmd/containerize/containerize.go "$@"
