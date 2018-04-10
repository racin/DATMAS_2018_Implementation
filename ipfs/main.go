package main

import (
	"github.com/ipfs/ipfs-cluster/api/rest/client"
	"github.com/ipfs/ipfs-cluster/api/rest"
	"strings"
	libp2p "github.com/libp2p/go-libp2p"
	pnet "github.com/libp2p/go-libp2p-pnet"
	ma "github.com/multiformats/go-multiaddr"
	"fmt"
	"context"
)

func init() {
	apiMAddr, _ := ma.NewMultiaddr("/ip4/127.0.0.1/tcp/0")
	cfg := &client.Config{
		APIAddr:           apiMAddr,
		DisableKeepAlives: true,
	}
	var secret [32]byte
	prot, err := pnet.NewV1ProtectorFromBytes(&secret)
	h, err := libp2p.New(
		context.Background(),
		libp2p.ListenAddrs(apiMAddr),
		libp2p.PrivateNetwork(prot),
	)
	if err != nil {
		panic(err)
	}
	rest, err := rest.NewAPIWithHost(cfg, h)
}
func apiMAddr(a *rest.API) ma.Multiaddr {
	listen, _ := a.HTTPAddress()
	hostPort := strings.Split(listen, ":")

	addr, _ := ma.NewMultiaddr(fmt.Sprintf("/ip4/127.0.0.1/tcp/%s", hostPort[1]))
	return addr
}


func main() {

}