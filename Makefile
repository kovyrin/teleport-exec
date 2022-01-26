cwd := $(shell pwd)
use_tty := $(shell [ -t 0 ] && echo "-it")
#--------------------------------------------------------------------------------------------------
all: protoc

#--------------------------------------------------------------------------------------------------
protoc: remote_exec/remote_exec.pb.go remote_exec/remote_exec_grpc.pb.go

remote_exec/remote_exec.pb.go: remote_exec/remote_exec.proto
		protoc --go_out=. --go_opt=paths=source_relative remote_exec/remote_exec.proto

remote_exec/remote_exec_grpc.pb.go: remote_exec/remote_exec.proto
		protoc --go-grpc_out=. --go-grpc_opt=paths=source_relative remote_exec/remote_exec.proto

#--------------------------------------------------------------------------------------------------
test:
		docker build -t teleport-exec-test -f Dockerfile.test .
		docker run -e TERM=color $(use_tty) --rm --privileged teleport-exec-test

lint:
		docker run -e TERM=color $(use_tty) --rm -v $(cwd):/app -w /app golangci/golangci-lint:v1.44.0 golangci-lint run
