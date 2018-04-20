package ipfsproxy

import (
	"github.com/ipfs/ipfs-cluster/api/rest/client"
	"github.com/ipfs/ipfs-cluster/api/rest"
	"strings"

	ma "github.com/multiformats/go-multiaddr"
	conf "github.com/racin/DATMAS_2018_Implementation/configuration"
	"fmt"
	"net/http"
	"github.com/racin/DATMAS_2018_Implementation/types"
	"encoding/json"
)



func writeResponse(w *http.ResponseWriter, codeType types.CodeType, message string){
	json.NewEncoder(*w).Encode(&types.IPFSReponse{Message:[]byte(message), Codetype:codeType})
}

func getClient(apiAddr ma.Multiaddr) *client.Client {
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

func main() {
	conf.LoadIPFSProxyConfig()
	localAPIAddr, _ := ma.NewMultiaddr(rest.DefaultHTTPListenAddr)
	remoteAPIAddr, _ := ma.NewMultiaddr(conf.IPFSProxyConfig().ListenAddr)
	proxy := &Proxy {
		client: getClient(localAPIAddr),
		localAPIAddr:localAPIAddr,
		remoteAPIAddr:remoteAPIAddr,
		seenTranc:make(map[string]bool),
	}

	fmt.Println("Starting IPFS proxy API")
	proxy.StartHTTPAPI()
}

func GetAPI() *rest.API {
	cfg := &rest.Config{}
	cfg.Default()

	api, err := rest.NewAPI(cfg)
	if err != nil {
		panic(err.Error())
	}
	return api
}

func apiMAddr(a *rest.API) ma.Multiaddr {
	listen, _ := a.HTTPAddress()
	hostPort := strings.Split(listen, ":")

	addr, _ := ma.NewMultiaddr(fmt.Sprintf("/ip4/127.0.0.1/tcp/%s", hostPort[1]))
	return addr
}