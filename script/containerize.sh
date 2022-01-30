#!/bin/bash

set -e

if [[ "$1" == "" ]]; then
  echo "Usage: $0 command args"
  exit 1
fi

make base_image
docker run -e TERM=color -it --rm --privileged teleport-exec-test go run -race cmd/containerize/containerize.go $*
