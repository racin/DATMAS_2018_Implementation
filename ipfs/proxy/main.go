package main

import (
	"github.com/ipfs/ipfs-cluster/api/rest/client"
	"github.com/ipfs/ipfs-cluster/api/rest"
	"strings"

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
	"bytes"
)

type Proxy struct {
	client				*client.Client
	localAPIAddr		ma.Multiaddr
	remoteAPIAddr		ma.Multiaddr
	seenTranc			map[string]bool
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

	// Empty Data parameter (Request is JSON of SignedTransaction)
	router.HandleFunc("/isup", proxy.IsUp).Methods("POST")
	router.HandleFunc("/statusall", proxy.StatusAll).Methods("POST")

	// Data parameter contains CID (Request is JSON of SignedTransaction)
	router.HandleFunc("/pinfile", proxy.PinFile).Methods("POST")
	router.HandleFunc("/unpinfile", proxy.UnPinFile).Methods("POST")
	router.HandleFunc("/get", proxy.GetFile).Methods("POST")
	router.HandleFunc("/status", proxy.Status).Methods("POST")

	// Data parameter contains CID, and has additional parameter file which contains the file.
	// Request is multipart/form-data. Transaction is in the transaction parameter
	router.HandleFunc("/addnopin", proxy.AddFileNoPin).Methods("POST")

	// Proof of Storage
	router.HandleFunc("/challenge", proxy.Challenge).Methods("POST")

	srv := &http.Server{
		Handler:      		router,
		Addr:         		conf.IPFSProxyConfig().ListenAddr,
		WriteTimeout: 		rest.DefaultWriteTimeout,
		ReadTimeout:  		rest.DefaultReadTimeout,
		IdleTimeout:		rest.DefaultIdleTimeout,
		ReadHeaderTimeout:	rest.DefaultReadHeaderTimeout,
	}

	if err := srv.ListenAndServe(); err != nil {
		panic("Error setting up IPFS proxy. Error: " + err.Error())
	}
}

func writeResponse(w *http.ResponseWriter, codeType types.CodeType, message string){
	json.NewEncoder(*w).Encode(&types.IPFSReponse{Message:message, Codetype:codeType})
}

func (proxy *Proxy) CheckProxyAccess(txString string, minAccessLevel app.AccessLevel) (*crypto.SignedStruct, types.CodeType, string) {
	tx := &crypto.SignedStruct{}
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

	// Check access rights
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
		return
	}

	// Check if data hash is contained within the transaction.
	fileHash, ok := tx.Data.(string)
	if (fileHash == "" || !ok) {
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
	fileBytes, err := ioutil.ReadAll(fopen)
	if err != nil {
		writeResponse(&w, types.CodeType_BCFSInvalidInput, "Could not get byte array of input file.");
		return
	} else if uplFileHash, err := crypto.IPFSHashData(fileBytes); err != nil {
		writeResponse(&w, types.CodeType_BCFSInvalidInput, "Could not get hash of input file.");
		return
	} else if uplFileHash != fileHash {
		writeResponse(&w, types.CodeType_BCFSInvalidInput, "Data hash parameter does not equal hash of uploaded file.");
		return
	}

	if resStr, err := proxy.client.IPFS().AddNoPin(bytes.NewReader(fileBytes)); err != nil {
		writeResponse(&w, types.CodeType_BCFSInvalidInput, resStr + ". Error: " + err.Error());
	} else {
		writeResponse(&w, types.CodeType_OK, resStr);
		// Add transaction to list of known transactions (message contains hash of tranc)
		proxy.seenTranc[message] = true
	}

}
func (proxy *Proxy) GetProof(cidStr string) *app.StorageChallengeProof{
	proof := &app.StorageChallengeProof{Proof:[]byte("abc")}
	err := proxy.client.IPFS().Get(cidStr, conf.IPFSProxyConfig().TempUploadPath)
	if err != nil {
		return nil
	}


	// Implement PoS here. Now simply check if hash is correct.
	if ipfsHash, err := crypto.IPFSHashFile(conf.IPFSProxyConfig().TempUploadPath + cidStr); err != nil {
		return nil
	} else if ipfsHash != cidStr {
		return nil
	}

	return proof
}
func (proxy *Proxy) Challenge(w http.ResponseWriter, r *http.Request) {
	// Only Consensus access level can execute this method.
	//

}
func (proxy *Proxy) PinFile(w http.ResponseWriter, r *http.Request) {
	txString, err := ioutil.ReadAll(r.Body)
	if err != nil {
		writeResponse(&w, types.CodeType_BCFSInvalidInput, "Missing transaction parameter.");
		return
	}

	// Check access to proxy method
	tx, codeType, message := proxy.CheckProxyAccess(string(txString), app.User)
	if codeType != types.CodeType_OK {
		writeResponse(&w, codeType, message);
		return
	}

	// Check if CID is contained within the transaction.
	cidStr, ok := tx.Data.(string)
	if (!ok) {
		writeResponse(&w, types.CodeType_BCFSInvalidInput, "Missing data hash parameter.");
		return
	}

	err = proxy.client.IPFS().Get(cidStr, conf.IPFSProxyConfig().TempUploadPath)
	if err != nil {
		writeResponse(&w, types.CodeType_BCFSUnknownAddress, "Could not find file with hash. Error: " + err.Error());
		return
	}

	// Pin file.
	b58, err := mh.FromB58String(cidStr)
	if err != nil {
		writeResponse(&w, types.CodeType_InternalError, err.Error());
		return
	}

	if err := proxy.client.Pin(cid2.NewCidV0(b58), -1, -1, ""); err != nil {
		writeResponse(&w, types.CodeType_InternalError, err.Error());
	} else {
		writeResponse(&w, types.CodeType_OK, "File pinned.");
	}
}
func (proxy *Proxy) UnPinFile(w http.ResponseWriter, r *http.Request) {
	// For removing stored data.
}

func (proxy *Proxy) GetFile(w http.ResponseWriter, r *http.Request) {

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
func (proxy *Proxy) Status(w http.ResponseWriter, r *http.Request) {
	txString, err := ioutil.ReadAll(r.Body)
	if err != nil {
		writeResponse(&w, types.CodeType_BCFSInvalidInput, "Missing transaction parameter.");
		return
	}

	// Check access to proxy method
	tx, codeType, message := proxy.CheckProxyAccess(string(txString), app.User)
	if codeType != types.CodeType_OK {
		writeResponse(&w, codeType, message);
		return
	}

	// Check if CID is contained within the transaction.
	cidStr, ok := tx.Data.(string)
	if (!ok) {
		writeResponse(&w, types.CodeType_BCFSInvalidInput, "Missing data hash parameter.");
		return
	}

	b58, err := mh.FromB58String(cidStr)
	if err != nil {
		writeResponse(&w, types.CodeType_InternalError, err.Error());
		return
	}

	cid := cid2.NewCidV0(b58)
	if pininfo, err := proxy.client.Status(cid,false); err != nil {
		writeResponse(&w, types.CodeType_InternalError, err.Error());
	} else {
		writeResponse(&w, types.CodeType_OK, fmt.Sprintf("%+v", pininfo));
		// Add transaction to list of known transactions (message contains hash of tranc)
		proxy.seenTranc[message] = true
	}
}

func (proxy *Proxy) StatusAll(w http.ResponseWriter, r *http.Request) {
	txString, err := ioutil.ReadAll(r.Body)
	if err != nil {
		writeResponse(&w, types.CodeType_BCFSInvalidInput, "Missing transaction parameter.");
		return
	}

	if _, codeType, message := proxy.CheckProxyAccess(string(txString), app.User); codeType != types.CodeType_OK {
		writeResponse(&w, codeType, message);
	} else if pininfo, err := proxy.client.StatusAll(false); err != nil {
		writeResponse(&w, types.CodeType_InternalError, err.Error());
	} else {
		writeResponse(&w, types.CodeType_OK, fmt.Sprintf("%+v", pininfo));

		// Add transaction to list of known transactions (message contains hash of tranc)
		proxy.seenTranc[message] = true
	}
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

	/*
	err = proxy.IPFSAddFile("/home/gob/racin.txt")
	if err != nil {
		panic(err.Error())
	}*/
}

func (proxy *Proxy) IPFSAddFile(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}

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


func apiMAddr(a *rest.API) ma.Multiaddr {
	listen, _ := a.HTTPAddress()
	hostPort := strings.Split(listen, ":")

	addr, _ := ma.NewMultiaddr(fmt.Sprintf("/ip4/127.0.0.1/tcp/%s", hostPort[1]))
	return addr
}

func (proxy *Proxy) GetAccessList() (*conf.AccessList){
	return conf.GetAccessList(conf.IPFSProxyConfig().BasePath + conf.IPFSProxyConfig().AccessList)
}