package main

import (
	"github.com/ipfs/ipfs-cluster/api/rest/client"
	"github.com/ipfs/ipfs-cluster/api/rest"
	"strings"
	//libp2p "github.com/libp2p/go-libp2p"
	//pnet "github.com/libp2p/go-libp2p-pnet"
	ma "github.com/multiformats/go-multiaddr"
	"fmt"
	"os"
)

func init() {
	// Expose APIs:
	/*
	IPFS
		Add file
		Get file
		IsUp !!



		IPFS-Cluster
		StatusAll
		Status(CID)
	 */
	 // Check AccessLevel
	 // Relay response
}

func testClientHTTP(apiAddr ma.Multiaddr) *client.Client {

	cfg := &client.Config{
		APIAddr: apiAddr,
		DisableKeepAlives: true,
	}
	c, err := client.NewClient(cfg)
	if err != nil {
		panic(err.Error())
	}

	return c
}

func apiMAddr(a *rest.API) ma.Multiaddr {
	listen, _ := a.HTTPAddress()
	hostPort := strings.Split(listen, ":")

	addr, _ := ma.NewMultiaddr(fmt.Sprintf("/ip4/127.0.0.1/tcp/%s", hostPort[1]))
	return addr
}


func main() {
	apiAddr, _ := ma.NewMultiaddr(rest.DefaultHTTPListenAddr)
	c := testClientHTTP(apiAddr)
	pininfo, err := c.StatusAll(false)
	if err != nil {
		panic(err.Error())
	}

	c.IPFS().
		fmt.Printf("%+v\n", pininfo)
	err = IPFSAddFile(c,"/home/gob/racin.txt")
	if err != nil {
		panic(err.Error())
	}
}

func IPFSAddFile(c *client.Client, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	result, err := c.IPFS().Add(file)
	fmt.Println(result)
	return err
}
func GetAPI() *rest.API {
	//logging.SetDebugLogging()
	//apiMAddr, _ := ma.NewMultiaddr(rest.DefaultHTTPListenAddr)

	cfg := &rest.Config{}
	cfg.Default()

	api, err := rest.NewAPI(cfg)
	if err != nil {
		panic(err.Error())
	}
	return api
	/*
		var secret [32]byte
		prot, err := pnet.NewV1ProtectorFromBytes(&secret)
		if err != nil {
			panic(err.Error())
		}

		h, err := libp2p.New(
			context.Background(),
			libp2p.ListenAddrs(apiMAddr),
			libp2p.PrivateNetwork(prot),
		)
		if err != nil {
			t.Fatal(err)
		}

		rest, err := rest.NewAPIWithHost(cfg, h)
		if err != nil {
			t.Fatal("should be able to create a new Api: ", err)
		}

		rest.SetClient(test.NewMockRPCClient(t))
		return rest*/
}