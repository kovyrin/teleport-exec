cwd := $(shell pwd)
use_tty := $(shell [ -t 0 ] && echo "-it")
docker_run := docker run -e TERM=color $(use_tty) --rm --privileged --cgroupns=host -v /sys/fs/cgroup:/sys/fs/cgroup:rw teleport-exec-test
#--------------------------------------------------------------------------------------------------
all: protoc

#--------------------------------------------------------------------------------------------------
protoc: remote_exec/remote_exec.pb.go remote_exec/remote_exec_grpc.pb.go

remote_exec/remote_exec.pb.go: remote_exec/remote_exec.proto
		protoc --go_out=. --go_opt=paths=source_relative remote_exec/remote_exec.proto

remote_exec/remote_exec_grpc.pb.go: remote_exec/remote_exec.proto
		protoc --go-grpc_out=. --go-grpc_opt=paths=source_relative remote_exec/remote_exec.proto

#--------------------------------------------------------------------------------------------------
base_image: protoc
		docker build -t teleport-exec-test .

test: base_image
		 $(docker_run) go test -race -v ./...

lint:
		docker run -e TERM=color $(use_tty) --rm -v $(cwd):/app -w /app golangci/golangci-lint:v1.44.0 golangci-lint run