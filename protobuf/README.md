## Data Transfer: Google Protocol Buffers

### Building protobuf for Go
This requires that $GOPATH/bin is in your PATH.
```sh
go get -u github.com/golang/protobuf/protoc-gen-go
protoc --go_out=generated-go StatusMessages.proto
```
