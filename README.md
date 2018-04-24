# DATMAS_2018_Implementation

## Incomplete installation:
* Go version 1.9+
* IPFS version 0.4.15
* Tendermint 0.19.0

```
curl -O https://dl.google.com/go/go1.10.1.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.10.1.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin:$GOPATH/bin

go get -u -d github.com/ipfs/go-ipfs
cd $GOPATH/src/github.com/ipfs/go-ipfs
make install
ipfs init

go get -u -d github.com/ipfs/ipfs-cluster
cd $GOPATH/src/github.com/ipfs/ipfs-cluster
make install
ipfs-cluster-service init

go get github.com/tendermint/tendermint/cmd/tendermint
cd $GOPATH/src/github.com/tendermint/tendermint
make get_tools
make get_vendor_deps
make install
tendermint init

cd $GOPATH/src/github.com
mkdir racin
cd racin/
git clone https://github.com/racin/DATMAS_2018_Implementation
cd DATMAS_2018_Implementation/
sh install.sh

# See: IPFS import conflicts
# See: Multiple registrations 

go build main.go
go build client/main.go
go build ipfsproxy/main.go
```

## Running 
Make sure $HOME/.tendermint/config/config.toml:
1. abci = "grpc"
2. create_empty_blocks = false

1. Start tendermint core with "tendermint node"
2. Start tendermint app with "cd $GOPATH/src/github.com/racin/DATMAS_2018_Implementation && ./main"
3. Start IPFS with "ipfs daemon"
4. Start IPFS-cluster with "ipfs-cluster-service"
5. Start IPFS proxy with "cd $GOPATH/src/github.com/racin/DATMAS_2018_Implementation/ipfsproxy && ./main"
6. Run the client with "cd $GOPATH/src/github.com/racin/DATMAS_2018_Implementation/client && ./main"
6a. Example client command: "./main data upload [file] [name] [description]"
6b. After upload is completed, open http://localhost:8080/ipfs/[CID]

## New protobuf:
```
protoc -I=types/ -I=$GOPATH/src -I=$GOPATH/src/github.com/gogo/protobuf/protobuf --go_out=plugins=grpc:types/ 
```

## Common problems:
### IPFS import conflicts:
In the latest versions there is some import conflics with IPFS. See discussion: https://github.com/ipfs/go-ipfs-api/issues/75
A simple fix that works in our case:
Edit: $GOPATH/src/gx/ipfs/QmZ4Qi3GaRbjcx28Sme5eMH7RQjGkt8wHxt2a65oLaeFEV/gogo-protobuf/proto/properties.go
1. Remove log import
2. Function RegisterEnum: Comment out the two panic calls
3. Function RegisterType: Comment out the log call

### Multiple registrations
'http: multiple registrations for /debug/requests'
```
go get -u golang.org/x/net/trace
rm -rf $GOPATH/src/github.com/tendermint/tendermint/vendor/golang.org/x/net/trace
```

## Generate certificate:
```
mkdir $HOME/.bcfs
cd $HOME/.bcfs
echo -ne '\n' | openssl req -x509 -nodes -newkey rsa:4096 -keyout mycert.pem -out mycert.pem
chmod 0600 mycert.pem
openssl rsa -in mycert.pem -pubout > mycert.pub
```

## Get fingerprint of certificate
### From Private key
```
ssh-keygen -yf mycert.pem > mycert_ssh.pub
ssh-keygen -E md5 -lf mycert_ssh.pub | egrep -o '([0-9a-f]{2}:){15}.{2}' | sed -E 's/://g'
```

### From Public key
```
ssh-keygen -E MD5 -lf /dev/stdin <<< $( ssh-keygen -im PKCS8 -f mycert.pub ) | egrep -o '([0-9a-f]{2}:){15}.{2}' | sed -E 's/://g'
```
