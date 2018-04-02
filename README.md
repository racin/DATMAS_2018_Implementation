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
./main -abci=grpc
```

## Generate certificate:
```
mkdir $HOME/.bcfs
cd $HOME/.bcfs
openssl req -x509 -nodes -newkey rsa:4196 -keyout mycert.pem -out mycert.pem (All questions can be skipped)
openssl rsa -in mycert.pem -pubout > mycert.pub
```

## Get fingerprint of certificate
```
openssl rsa -pubin -inform PEM -in mycert_test.pub -pubout -outform PEM | openssl md5 -c

ssh-keygen -E MD5 -lf /dev/stdin <<< $( ssh-keygen -im PKCS8 -f mycert_test.pub ) | sed s/[0-9]+ MD5://g
```
