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
	router.HandleFunc("/unpinfile/{cid}", proxy.UnPinFile).Methods("DELETE")
	router.HandleFunc("/isup", proxy.IsUp).Methods("GET")
	router.HandleFunc("/get/{cid}", proxy.GetFile).Methods("GET")
	router.HandleFunc("/status/{cid}", proxy.Status).Methods("GET")
	router.HandleFunc("/statusall", proxy.StatusAll).Methods("GET")
	if err := http.ListenAndServe(conf.IPFSProxyConfig().ListenAddr, router); err != nil {
		panic("Error setting up IPFS proxy. Error: " + err.Error())
	}
}

func writeUploadResponse(w *http.ResponseWriter, codeType types.CodeType, message string){
	//json.NewEncoder(*w).Encode(&types.ResponseUpload{Message:message, Codetype:codeType})
	byteArr, _ := json.Marshal(&types.ResponseUpload{Message:message, Codetype:codeType})
	fmt.Fprintf(*w, "%s", byteArr)
}

func (proxy *Proxy) AddFileNoPin(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(104857600) // Up to 100MB stored in memory.
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}
	formdata := r.MultipartForm // ok, no problem so far, read the Form data

	txString, ok := formdata.Value["transaction"]
	if !ok {
		writeUploadResponse(&w, types.CodeType_BCFSInvalidInput, "Missing transaction parameter.");
		return
	}

	tx := &app.SignedTransaction{}
	if err := json.Unmarshal([]byte(txString[0]), tx); err != nil {
		writeUploadResponse(&w, types.CodeType_BCFSInvalidInput, "Could not Marshal transaction");
		return
	}

	// Check identity access
	identity, ok := app.GetAccessList().Identities[tx.Identity];
	if !ok {
		writeUploadResponse(&w, types.CodeType_Unauthorized, "Could not get access list");
		return
	}

	// Verify signature in transaction.
	if ok, msg := app.VerifySignature(&identity, tx); !ok {
		writeUploadResponse(&w, types.CodeType_BCFSInvalidSignature, msg);
		return
	}

	// Check if uploader is allowed to upload data.
	if identity.AccessLevel < 1 {
		writeUploadResponse(&w, types.CodeType_Unauthorized, "Insufficient access level");
		return
	}

	// Check if data hash is contained within the transaction.
	fileHash, ok := tx.Data.(string)
	if (!ok) {
		writeUploadResponse(&w, types.CodeType_BCFSInvalidInput, "Missing data hash parameter.");
		return
	}

	// Check if data hash is already in the list of uploads pending
	if _, ok := app.tempUploads[fileHash]; !ok {
		writeUploadResponse(&w, types.CodeType_BCFSInvalidInput, "Data hash not in the list of pending uploads.");
		return // Data hash not in the list of pending uploads
	}

	files, ok := formdata.File["file"]
	if !ok || len(files) > 1 {
		writeUploadResponse(&w, types.CodeType_BCFSInvalidInput, "File parameter should contain exactly one file.");
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
		writeUploadResponse(&w, types.CodeType_BCFSInvalidInput, "Could not get byte array of input file.");
		return
	} else if uplFileHash, err := crypto.IPFSHashData(fileBytes); err != nil {
		writeUploadResponse(&w, types.CodeType_BCFSInvalidInput, "Could not get hash of input file.");
		return
	} else if uplFileHash != fileHash {
		writeUploadResponse(&w, types.CodeType_BCFSInvalidInput, "Hash of input file not present in transaction.");
		return
	}

	out, err := os.Create("/tmp/" + file.Filename)
	defer out.Close()
	if err != nil {
		writeUploadResponse(&w, types.CodeType_Unauthorized, "Unable to create the file for writing. Check your write access privilege");
		return
	}

	_, err = io.Copy(out, fopen) // file not files[i] !

	if err != nil {
		writeUploadResponse(&w, types.CodeType_InternalError, err.Error());
		return
	}

	writeUploadResponse(&w, types.CodeType_OK, "Files uploaded successfully : " + file.Filename);
}
func (proxy *Proxy) PinFile(w http.ResponseWriter, r *http.Request) {

}
func (proxy *Proxy) UnPinFile(w http.ResponseWriter, r *http.Request) {

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