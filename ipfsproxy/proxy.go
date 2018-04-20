package ipfsproxy

import (
	"github.com/ipfs/ipfs-cluster/api/rest/client"
	"github.com/ipfs/ipfs-cluster/api/rest"

	ma "github.com/multiformats/go-multiaddr"
	conf "github.com/racin/DATMAS_2018_Implementation/configuration"

	"net/http"
	"github.com/gorilla/mux"
	"github.com/racin/DATMAS_2018_Implementation/types"
	"encoding/json"
	"github.com/racin/DATMAS_2018_Implementation/crypto"
)

type Proxy struct {
	client				*client.Client
	localAPIAddr		ma.Multiaddr
	remoteAPIAddr		ma.Multiaddr
	seenTranc			map[string]bool

	privKey				*crypto.Keys
	identity			*conf.Identity
	fingerprint			string
}

func NewProxy() *Proxy {
	conf.LoadIPFSProxyConfig()
	localAPIAddr, _ := ma.NewMultiaddr(rest.DefaultHTTPListenAddr)
	remoteAPIAddr, _ := ma.NewMultiaddr(conf.IPFSProxyConfig().ListenAddr)

	proxy := &Proxy {
		client: getClient(localAPIAddr),
		localAPIAddr:localAPIAddr,
		remoteAPIAddr:remoteAPIAddr,
		seenTranc:make(map[string]bool),
	}

	// Load private keys in order to later digitally sign transactions
	if myPrivKey, err := crypto.LoadPrivateKey(conf.IPFSProxyConfig().BasePath + conf.IPFSProxyConfig().PrivateKey); err != nil {
		panic("Could not load private key. Error: " + err.Error())
	} else if fp, err := crypto.GetFingerprint(myPrivKey); err != nil{
		panic("Could not get fingerprint of private key.")
	} else {
		proxy.fingerprint = fp;
		proxy.privKey = myPrivKey
		proxy.identity = proxy.GetAccessList().Identities[fp]
	}

	return proxy
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

func (proxy *Proxy) CheckProxyAccess(txString string, openForTypes... conf.NodeType) (*types.Transaction, types.CodeType, string) {
	stx := &crypto.SignedStruct{Base: &types.Transaction{}}
	var tx types.Transaction
	var ok bool = false
	if err := json.Unmarshal([]byte(txString), stx); err != nil {
		return nil, types.CodeType_BCFSInvalidInput, "Could not Marshal transaction. Error: " + err.Error()
	} else if tx, ok = stx.Base.(types.Transaction); !ok {
		return nil, types.CodeType_BCFSInvalidInput, "Could not Marshal transaction."
	}

	// Check for replay attack
	txHash := crypto.HashStruct(tx)
	if proxy.HasSeenTranc(txHash) {
		return nil, types.CodeType_BadNonce, "Could not process transaction. Possible replay attack."
	}

	// Get signers identity and public key
	signer, pubKey := proxy.GetIdentityPublicKey(tx.Identity)
	if signer == nil {
		return nil, types.CodeType_Unauthorized, "Could not get access list"
	}

	// Check if public key exists and if message is signed.
	if pubKey == nil {
		return nil, types.CodeType_BCFSInvalidSignature, "Could not locate public key"
	} else if !stx.Verify(pubKey) {
		return nil, types.CodeType_BCFSInvalidSignature, "Could not verify signature"
	}

	// Check access rights
	for allowedType := range openForTypes {
		if signer.Type == openForTypes[allowedType] {
			return &tx, types.CodeType_OK, txHash
		}
	}

	return nil, types.CodeType_Unauthorized, "Insufficient access level"
}

func (proxy *Proxy) GetAccessList() (*conf.AccessList){
	return conf.GetAccessList(conf.IPFSProxyConfig().BasePath + conf.IPFSProxyConfig().AccessList)
}

func (proxy *Proxy) GetIdentityPublicKey(ident string) (identity *conf.Identity, pubkey *crypto.Keys){
	return crypto.GetIdentityPublicKey(ident, proxy.GetAccessList(), conf.IPFSProxyConfig().BasePath + conf.IPFSProxyConfig().PublicKeys)
}