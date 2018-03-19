# DATMAS_2018_Implementation

## Incomplete installation:
* Go version 1.9+
* IPFS version 0.4.13+
* Tendermint 0.16.0 +

```
git clone git@github.com:racin/DATMAS_2018_Implementation
go get -u github.com/golang/protobuf/protoc-gen-go
go get github.com/gogo/protobuf/protoc-gen-gofast
go get github.com/gogo/protobuf/gogoproto
protoc -I=types/ -I=$GOPATH/src -I=$GOPATH/src/github.com/gogo/protobuf/protobuf --go_out=plugins=grpc:. types.proto
go build main.go
```
