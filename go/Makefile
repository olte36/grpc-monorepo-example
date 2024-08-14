GOPATH=$(shell go env GOPATH)

gen_proto: clean
	go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.34.2
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.5.1
	protoc --proto_path=../ --go_out=./ --go-grpc_out=./ \
		--plugin=${GOPATH}/bin/protoc-gen-go \
		--plugin=${GOPATH}/bin/protoc-gen-go-grpc \
		../api/*.proto

clean:
	rm -rf ./genproto
