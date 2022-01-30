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
		docker run -e TERM=color -it --rm --privileged teleport-exec-test \
			go test -race -v ./...

ping: base_image
		docker run -e TERM=color -it --rm --privileged teleport-exec-test \
			go run -race cmd/containerize/containerize.go ping -O 8.8.8.8
