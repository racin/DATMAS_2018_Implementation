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
echo -ne '\n' | openssl req -x509 -nodes -newkey rsa:4196 -keyout mycert.pem -out mycert.pem
openssl rsa -in mycert.pem -pubout > mycert.pub
```

## Get fingerprint of certificate
### From Private key
```
ssh-keygen -yf mycert_test.pem > mycert_test_ssh.pub
ssh-keygen -E md5 -lf mycert_test_ssh.pub | egrep -o '([0-9a-f]{2}:){15}.{2}' | sed -E 's/://g'
```

### From Public key
```
ssh-keygen -E MD5 -lf /dev/stdin <<< $( ssh-keygen -im PKCS8 -f mycert_test.pub ) | egrep -o '([0-9a-f]{2}:){15}.{2}' | sed -E 's/://g'
```
