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
	"github.com/racin/DATMAS_2018_Implementation/app"
	"encoding/json"
	cid2 "github.com/ipfs/go-cid"
	mh "github.com/multiformats/go-multihash"
	"github.com/racin/DATMAS_2018_Implementation/crypto"
	"io/ioutil"
)

type Proxy struct {
	client				*client.Client
	localAPIAddr		ma.Multiaddr
	remoteAPIAddr		ma.Multiaddr
	seenTranc			map[string]bool
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

// Simple check to prevent Replay attacks. Not reliable.
func (proxy *Proxy) HasSeenTranc(trancHash string) bool{
	if _, ok := proxy.seenTranc[trancHash]; ok {
		return true;
	}
	return false;
}

func (proxy *Proxy) StartHTTPAPI(){
	router := mux.NewRouter()
	router.HandleFunc("/addnopin", proxy.AddFileNoPin).Methods("POST")
	router.HandleFunc("/pinfile/{cid}", proxy.PinFile).Methods("POST")
	router.HandleFunc("/unpinfile/{cid}", proxy.UnPinFile).Methods("POST")
	router.HandleFunc("/isup", proxy.IsUp).Methods("POST")
	router.HandleFunc("/get/{cid}", proxy.GetFile).Methods("POST")
	router.HandleFunc("/status/{cid}", proxy.Status).Methods("POST")
	router.HandleFunc("/statusall", proxy.StatusAll).Methods("POST")
	if err := http.ListenAndServe(conf.IPFSProxyConfig().ListenAddr, router); err != nil {
		panic("Error setting up IPFS proxy. Error: " + err.Error())
	}
}

func writeResponse(w *http.ResponseWriter, codeType types.CodeType, message string){
	json.NewEncoder(*w).Encode(&types.IPFSReponse{Message:message, Codetype:codeType})
}

func (proxy *Proxy) CheckProxyAccess(txString string, minAccessLevel app.AccessLevel) (*app.SignedTransaction, types.CodeType, string) {
	tx := &app.SignedTransaction{}
	if err := json.Unmarshal([]byte(txString), tx); err != nil {
		return nil, types.CodeType_BCFSInvalidInput, "Could not Marshal transaction"
	}

	// Check for replay attack
	txHash := tx.Hash()
	if proxy.HasSeenTranc(txHash) {
		return nil, types.CodeType_BadNonce, "Could not process transaction. Possible replay attack."
	}

	// Check identity access
	identity, ok := app.GetAccessList("ipfs").Identities[tx.Identity];
	if !ok {
		return nil, types.CodeType_Unauthorized, "Could not get access list"
	}


	// Verify signature in transaction. (Temp disabled)
	/*if ok, msg := app.VerifySignature(conf.IPFSProxyConfig().BasePath + conf.IPFSProxyConfig().PublicKeys + identity.PublicKey,
		txHash, tx.Signature); !ok {
		return nil, types.CodeType_BCFSInvalidSignature, msg
	}*/

	// Check if uploader is allowed to upload data.
	if identity.AccessLevel < minAccessLevel {
		return nil, types.CodeType_Unauthorized, "Insufficient access level"
	}

	return tx, types.CodeType_OK, txHash
}

func (proxy *Proxy) AddFileNoPin(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(104857600) // Up to 100MB stored in memory.
	if err != nil {
		writeResponse(&w, types.CodeType_InternalError, err.Error());
		return
	}
	formdata := r.MultipartForm
	txString, ok := formdata.Value["transaction"]
	if !ok {
		writeResponse(&w, types.CodeType_BCFSInvalidInput, "Missing transaction parameter.");
		return
	}

	// Check access to proxy method
	tx, codeType, message := proxy.CheckProxyAccess(txString[0], app.User)
	if codeType != types.CodeType_OK {
		writeResponse(&w, codeType, message);
	}

	// Check if data hash is contained within the transaction.
	fileHash, ok := tx.Data.(string)
	if (!ok) {
		writeResponse(&w, types.CodeType_BCFSInvalidInput, "Missing data hash parameter.");
		return
	}

	files, ok := formdata.File["file"]
	if !ok || len(files) > 1 {
		writeResponse(&w, types.CodeType_BCFSInvalidInput, "File parameter should contain exactly one file.");
		return // Missing files or more than one file
	}

	file := files[0]
	fopen, err := file.Open()
	defer fopen.Close()
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}

	// Check if the hash of the upload file equals the hash contained in the transaction
	if fileBytes, err := ioutil.ReadAll(fopen); err != nil {
		writeResponse(&w, types.CodeType_BCFSInvalidInput, "Could not get byte array of input file.");
		return
	} else if uplFileHash, err := crypto.IPFSHashData(fileBytes); err != nil {
		writeResponse(&w, types.CodeType_BCFSInvalidInput, "Could not get hash of input file.");
		return
	} else if uplFileHash != fileHash {
		writeResponse(&w, types.CodeType_BCFSInvalidInput, "Hash of input file not present in transaction.");
		return
	}

	if resStr, err := proxy.client.IPFS().AddNoPin(fopen); err != nil {
		writeResponse(&w, types.CodeType_BCFSInvalidInput, resStr + ". Error: " + err.Error());
	} else {
		writeResponse(&w, types.CodeType_OK, resStr);
	}

}
func (proxy *Proxy) PinFile(w http.ResponseWriter, r *http.Request) {

}
func (proxy *Proxy) UnPinFile(w http.ResponseWriter, r *http.Request) {

}
/**
Simply checks if the IPFS service is up. Does not need to protected.
 */
func (proxy *Proxy) IsUp(w http.ResponseWriter, r *http.Request) {
	if proxy.client.IPFS().IsUp() {
		writeResponse(&w, types.CodeType_OK, "true");
	} else {
		writeResponse(&w, types.CodeType_OK, "false");
	}
}
func (proxy *Proxy) GetFile(w http.ResponseWriter, r *http.Request) {

}
func (proxy *Proxy) Status(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	b58, err := mh.FromB58String(params["cid"])
	if err != nil {
		writeResponse(&w, types.CodeType_InternalError, err.Error());
		return
	}

	cid := cid2.NewCidV0(b58)

	if pininfo, err := proxy.client.Status(cid,false); err != nil {
		writeResponse(&w, types.CodeType_InternalError, err.Error());
	} else {
		writeResponse(&w, types.CodeType_OK, fmt.Sprintf("%+v", pininfo));
	}
}

func (proxy *Proxy) StatusAll(w http.ResponseWriter, r *http.Request) {
	txString, err := ioutil.ReadAll(r.Body)
	if err != nil {
		writeResponse(&w, types.CodeType_BCFSInvalidInput, "Missing transaction parameter.");
		return
	}
	_, codeType, message := proxy.CheckProxyAccess(string(txString), app.User);
	if codeType != types.CodeType_OK {
		writeResponse(&w, codeType, message);
		return
	}
	if pininfo, err := proxy.client.StatusAll(false); err != nil {
		writeResponse(&w, types.CodeType_InternalError, err.Error());
	} else {
		writeResponse(&w, types.CodeType_OK, fmt.Sprintf("%+v", pininfo));
	}

	// Add transaction to list of known transactions (message contains hash of tranc)
	proxy.seenTranc[message] = true
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
		seenTranc:make(map[string]bool),
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