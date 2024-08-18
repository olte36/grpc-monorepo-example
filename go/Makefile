GOPATH=$(shell go env GOPATH)

clean:
	rm -rf ./genproto

gen_proto: clean
	go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.34.2
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.5.1
	protoc --proto_path=../ --go_out=./ --go-grpc_out=./ \
		--plugin=${GOPATH}/bin/protoc-gen-go \
		--plugin=${GOPATH}/bin/protoc-gen-go-grpc \
		../api/*.proto

build: gen_proto
	go build -o ./out/go_client ./client
	go build -o ./out/go_server ./server
