package app

import (
	"github.com/racin/DATMAS_2018_Implementation/crypto"
	"github.com/racin/DATMAS_2018_Implementation/types"
	abci "github.com/tendermint/abci/types"
	"fmt"
	conf "github.com/racin/DATMAS_2018_Implementation/configuration"
	"net/http"
	"time"
	rpcClient "github.com/tendermint/tendermint/rpc/client"
)

type Application struct {
	abci.BaseApplication

	info 				string

	tempUploads 		map[string]bool
	seenTranc 			map[string]bool

	IpfsHttpClient		*http.Client
	TMRpcClients		map[string]rpcClient.Client

	privKey				*crypto.Keys
	identity			*conf.Identity
	fingerprint			string

	prevailingBlock		map[string]int64
	nextBlockHeight		int64
}

func NewApplication() *Application {
	app := &Application{info: conf.AppConfig().Info,
		tempUploads: make(map[string]bool), seenTranc: make(map[string]bool),
		IpfsHttpClient: &http.Client{Timeout: time.Duration(conf.AppConfig().IpfsProxyTimeoutSeconds) * time.Second},
		prevailingBlock: make(map[string]int64), nextBlockHeight:1}

	// Load private keys in order to later digitally sign transactions
	if myPrivKey, err := crypto.LoadPrivateKey(conf.AppConfig().BasePath + conf.AppConfig().PrivateKey); err != nil {
		panic("Could not load private key. Error: " + err.Error())
	} else if fp, err := crypto.GetFingerprint(myPrivKey); err != nil{
		panic("Could not get fingerprint of private key.")
	} else {
		app.fingerprint = fp;
		app.privKey = myPrivKey
		app.identity = app.GetAccessList().Identities[fp]
	}

	app.setupTMRpcClients()
	return app
}

func (app *Application) Info(abci.RequestInfo) (resInfo abci.ResponseInfo) {
	fmt.Println("Info trigger");
	return abci.ResponseInfo{Data: app.info}
}

func (app *Application) CheckTx(txBytes []byte) abci.ResponseCheckTx {
	stx, tx, err := types.UnmarshalTransaction(txBytes)
	if err != nil {
		return abci.ResponseCheckTx{Code: uint32(types.CodeType_InternalError), Log: err.Error()}
	}

	signer, pubKey := app.GetIdentityPublicKey(tx.Identity)
	if signer == nil {
		return abci.ResponseCheckTx{Code: uint32(types.CodeType_Unauthorized), Log: "Could not get access list"}
	}

	// Check if public key exists and if message is signed.
	if pubKey == nil {
		return abci.ResponseCheckTx{Code: uint32(types.CodeType_BCFSInvalidSignature), Log: "Could not locate public key"}
	} else if !stx.Verify(pubKey) {
		return abci.ResponseCheckTx{Code: uint32(types.CodeType_BCFSInvalidSignature), Log: "Could not verify signature"}
	}

	switch tx.Type {
	case types.TransactionType_UploadData:
		{
			fmt.Println("CheckTx_UploadData")
			return *app.CheckTx_UploadData(signer, tx)
		}
	case types.TransactionType_RemoveData:
		{
			fmt.Println("CheckTx_RemoveData")
			return *app.CheckTx_RemoveData(signer, tx)
		}
	case types.TransactionType_VerifyStorage:
		{
			fmt.Println("CheckTx_VerifyStorage")
			return *app.CheckTx_VerifyStorage(signer, tx)
		}
	case types.TransactionType_ChangeContentAccess:
		{
			fmt.Println("CheckTx_ChangeContentAccess")
			return *app.CheckTx_ChangeContentAccess(signer, tx)
		}
	default:
		{
			return abci.ResponseCheckTx{Code: uint32(types.CodeType_BCFSInvalidInput), Info: "Unknown transaction type."};
		}
	}
}

func (app *Application) DeliverTx(txBytes []byte)  abci.ResponseDeliverTx {
	stx, tx, err := types.UnmarshalTransaction(txBytes)
	if err != nil {
		return abci.ResponseDeliverTx{Code: uint32(types.CodeType_InternalError), Log: err.Error()}
	}

	signer, pubKey := app.GetIdentityPublicKey(tx.Identity)
	if signer == nil {
		return abci.ResponseDeliverTx{Code: uint32(types.CodeType_Unauthorized), Log: "Could not get access list"}
	}

	// Check if public key exists and if message is signed.
	if pubKey == nil {
		return abci.ResponseDeliverTx{Code: uint32(types.CodeType_BCFSInvalidSignature), Log: "Could not locate public key"}
	} else if !stx.Verify(pubKey) {
		return abci.ResponseDeliverTx{Code: uint32(types.CodeType_BCFSInvalidSignature), Log: "Could not verify signature"}
	}
	switch tx.Type {
	case types.TransactionType_UploadData:
		{
			fmt.Println("DeliverTx_UploadData")
			return *app.DeliverTx_UploadData(signer, tx)
		}
	case types.TransactionType_RemoveData:
		{
			fmt.Println("DeliverTx_RemoveData")
			return *app.DeliverTx_RemoveData(signer, tx)
		}
	case types.TransactionType_VerifyStorage:
		{
			fmt.Println("DeliverTx_VerifyStorage")
			return *app.DeliverTx_VerifyStorage(signer, tx)
		}
	case types.TransactionType_ChangeContentAccess:
		{
			fmt.Println("DeliverTx_ChangeContentAccess")
			return *app.DeliverTx_ChangeContentAccess(signer, tx)
		}
	default:
		{
			return abci.ResponseDeliverTx{Code: uint32(types.CodeType_BCFSInvalidInput), Info: "Unknown transaction type."};
		}
	}
}


func (app *Application) Commit() abci.ResponseCommit { //types.Result {
	app.nextBlockHeight++
	fmt.Printf("Commit trigger. Next Block height: %v\n",app.nextBlockHeight);
	// TODO: Why will it endlessly create new blocks if this is put as Data ?
	/*b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(app.nextBlockHeight))*/
	return abci.ResponseCommit{/*Data:b*/}
}

func (app *Application) Query(reqQuery abci.RequestQuery) (abci.ResponseQuery) {
	switch reqQuery.Path {
	// Newsample is not in use.
	/*case "/newsample":
		{
			return *app.Query_Newsample(reqQuery)
		}*/
	case "/challenge":
		{
			return *app.Query_Challenge(reqQuery)
		}
	case "/prevailingheight":
		{
			return *app.Query_PrevailingHeight(reqQuery)
		}
	default:
		{
			return abci.ResponseQuery{Code: uint32(types.CodeType_BCFSInvalidInput), Log: "Invalid query path."}
		}
	}
}

func (app *Application) GetAccessList() (*conf.AccessList){
	return conf.GetAccessList(conf.AppConfig().BasePath + conf.AppConfig().AccessList)
}

func (app *Application) GetIdentityPublicKey(ident string) (identity *conf.Identity, pubkey *crypto.Keys){
	return crypto.GetIdentityPublicKey(ident, app.GetAccessList(), conf.AppConfig().BasePath + conf.AppConfig().PublicKeys)
}

// Simple check to prevent replay attacks.
func (app *Application) HasSeenTranc(trancHash string) bool{
	if _, ok := app.seenTranc[trancHash]; ok {
		return true;
	}
	return false;
}

func (app *Application) GetSignedTransaction(txtype types.TransactionType, data interface{}) (stranc *crypto.SignedStruct) {
	tx := types.NewTx(data, app.fingerprint, txtype)
	stranc, err := crypto.SignStruct(tx, app.privKey);
	if err != nil {
		panic("Could not sign transaction. Private/Public key pair may not match. Use the --generateKeys to generate a new one. Error: " + err.Error())
	}
	return
}