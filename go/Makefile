
gen_proto:
	../../../../protoc/bin/protoc --proto_path=../ \
		--go_out=./ --go-grpc_out=./ \
		--plugin=../../../../go/bin/protoc-gen-go \
		--plugin=../../../../go/bin/protoc-gen-go-grpc \
		../api/test.proto