# Go implementaion of client/server gRPC communication.

## Generate code from protobuf
Go and protoc must be installed and available in $PATH.

To generate code from Protobuf files run:
```shell
make gen_proto
```
The directory with the generated code is added to [.gitignore](../.gitignore).
It should not be commited.

## Run the client/server
```shell
make build

./out/go_server
./out/go_client
```
or
```shell
make gen_proto

go run server/main.go
go run client/main.go
```

