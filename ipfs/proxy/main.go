package main

import (
	"github.com/ipfs/ipfs-cluster/api/rest/client"
	"github.com/ipfs/ipfs-cluster/api/rest"
	"strings"
	//libp2p "github.com/libp2p/go-libp2p"
	//pnet "github.com/libp2p/go-libp2p-pnet"
	ma "github.com/multiformats/go-multiaddr"
	conf "github.com/racin/DATMAS_2018_Implementation/configuration"
	"fmt"
	"os"
	"net/http"
	"github.com/gorilla/mux"
	"github.com/racin/DATMAS_2018_Implementation/types"
	"encoding/json"
	cid2 "github.com/ipfs/go-cid"
	mh "github.com/multiformats/go-multihash"
)

type Proxy struct {
	client				*client.Client
	localAPIAddr		ma.Multiaddr
	remoteAPIAddr		ma.Multiaddr
}

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

func (proxy *Proxy) StartHTTPAPI(){
	router := mux.NewRouter()
	router.HandleFunc("/addnopin", proxy.AddFileNoPin).Methods("POST")
	router.HandleFunc("/pinfile/{cid}", proxy.PinFile).Methods("GET")
	router.HandleFunc("/remove/{cid}", proxy.RemoveFile).Methods("DELETE")
	router.HandleFunc("/isup", proxy.IsUp).Methods("GET")
	router.HandleFunc("/get/{cid}", proxy.GetFile).Methods("GET")
	router.HandleFunc("/status/{cid}", proxy.Status).Methods("GET")
	router.HandleFunc("/statusall", proxy.StatusAll).Methods("GET")
	if err := http.ListenAndServe(conf.IPFSProxyConfig().ListenAddr, router); err != nil {
		panic("Error setting up IPFS proxy. Error: " + err.Error())
	}
}

func (proxy *Proxy) AddFileNoPin(w http.ResponseWriter, r *http.Request) {

}
func (proxy *Proxy) PinFile(w http.ResponseWriter, r *http.Request) {

}
func (proxy *Proxy) RemoveFile(w http.ResponseWriter, r *http.Request) {

}
func (proxy *Proxy) IsUp(w http.ResponseWriter, r *http.Request) {
	var resp *types.IPFSReponse

	if proxy.client.IPFS().IsUp() {
		resp = &types.IPFSReponse{Message:"true"}
	} else {
		resp = &types.IPFSReponse{Message:"false"}
	}
	json.NewEncoder(w).Encode(resp)
}
func (proxy *Proxy) GetFile(w http.ResponseWriter, r *http.Request) {

}
func (proxy *Proxy) Status(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	var resp *types.IPFSReponse

	b58, err := mh.FromB58String(params["cid"])
	if err != nil {
		resp = &types.IPFSReponse{Error:err.Error()}
	}

	cid := cid2.NewCidV0(b58)

	if pininfo, err := proxy.client.Status(cid,false); err != nil {
		resp = &types.IPFSReponse{Error:err.Error()}
	} else {
		resp = &types.IPFSReponse{Message:fmt.Sprintf("%+v", pininfo)}
	}
	json.NewEncoder(w).Encode(resp)
}
func (proxy *Proxy) StatusAll(w http.ResponseWriter, r *http.Request) {
	var resp *types.IPFSReponse
	if pininfo, err := proxy.client.StatusAll(false); err != nil {
		resp = &types.IPFSReponse{Error:err.Error()}
	} else {
		resp = &types.IPFSReponse{Message:fmt.Sprintf("%+v", pininfo)}
	}
	json.NewEncoder(w).Encode(resp)
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

func apiMAddr(a *rest.API) ma.Multiaddr {
	listen, _ := a.HTTPAddress()
	hostPort := strings.Split(listen, ":")

	addr, _ := ma.NewMultiaddr(fmt.Sprintf("/ip4/127.0.0.1/tcp/%s", hostPort[1]))
	return addr
}


func main() {
	conf.LoadIPFSProxyConfig()
	localAPIAddr, _ := ma.NewMultiaddr(rest.DefaultHTTPListenAddr)
	remoteAPIAddr, _ := ma.NewMultiaddr(conf.IPFSProxyConfig().ListenAddr)
	proxy := &Proxy {
		client: getClient(localAPIAddr),
		localAPIAddr:localAPIAddr,
		remoteAPIAddr:remoteAPIAddr,
	}

	fmt.Println("Starting IPFS proxy API")
	proxy.StartHTTPAPI()

	/*
	err = proxy.IPFSAddFile("/home/gob/racin.txt")
	if err != nil {
		panic(err.Error())
	}*/
}

func (proxy *Proxy) StartUploadHandler(){
	/*http.HandleFunc(conf.AppConfig().UploadEndpoint, app.UploadHandler)
	if err := http.ListenAndServe(app.uploadAddr, nil); err != nil {
		panic("Error setting up upload handler. Error: " + err.Error())
	}*/
}

func (proxy *Proxy) IPFSAddFile(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	//result, err := proxy.client.IPFS().Unpin()
	result, err := proxy.client.IPFS().Add(file)
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