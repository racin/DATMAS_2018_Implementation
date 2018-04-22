package app

import (
	"encoding/json"
	"log"
	"github.com/racin/DATMAS_2018_Implementation/crypto"
	"github.com/racin/DATMAS_2018_Implementation/types"
	abci "github.com/tendermint/abci/types"
	"fmt"
	conf "github.com/racin/DATMAS_2018_Implementation/configuration"
	"net/http"
	"time"
	rpcClient "github.com/tendermint/tendermint/rpc/client"
	"github.com/pkg/errors"
)

type Application struct {
	abci.BaseApplication

	info 				string
	uploadAddr 			string

	tempUploads 		map[string]bool
	seenTranc 			map[string]bool

	IpfsHttpClient		*http.Client
	TMRpcClients		map[string]rpcClient.Client

	privKey				*crypto.Keys
	identity			*conf.Identity
	fingerprint			string
}

func NewApplication() *Application {
	app := &Application{info: conf.AppConfig().Info, uploadAddr: conf.AppConfig().UploadAddr,
		tempUploads: make(map[string]bool), seenTranc: make(map[string]bool),
		IpfsHttpClient: &http.Client{Timeout: time.Duration(conf.AppConfig().IpfsProxyTimeoutSeconds) * time.Second}}

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
func (app *Application) DeliverTx(txBytes []byte)  abci.ResponseDeliverTx {
	txHash, _ := crypto.IPFSHashData(txBytes)
	fmt.Println("Deliver trigger. Hash of data: " + txHash);
	stx := &crypto.SignedStruct{Base: &types.Transaction{}}
	var tx *types.Transaction
	var ok bool = false
	if err := json.Unmarshal(txBytes, stx); err != nil {
		return abci.ResponseDeliverTx{Code: uint32(types.CodeType_InternalError), Log: err.Error()}
	} else if tx, ok = stx.Base.(*types.Transaction); !ok {
		return abci.ResponseDeliverTx{Code: uint32(types.CodeType_InternalError), Log: "Could not unmarshal transaction (Transaction)"}
	}
	fmt.Printf("Hash of transaction: %s\n",crypto.HashStruct(tx))

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
	case types.TransactionType_DownloadData:
		{
			fmt.Println("DeliverTx_DownloadData")
			return *app.DeliverTx_DownloadData(signer, tx)
		}

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
			// TODO: Is this really needed?
			return abci.ResponseDeliverTx{Info: "Error"};
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
/*
func (app *Application) verifySignatureTx(txBytes []byte, expectedData interface{}) (ident *conf.Identity, pubKey *crypto.Keys, resp *abci.ResponseCheckTx) {

}*/

func (app *Application) unmarshalTransaction(txBytes []byte) (*crypto.SignedStruct, *types.Transaction, error) {
	stx := &crypto.SignedStruct{Base: &types.Transaction{}}
	if err := json.Unmarshal(txBytes, stx); err != nil {
		return nil, nil, err
	} else if tx, ok := stx.Base.(*types.Transaction); !ok {
		return nil, nil, errors.New("Could not unmarshal transaction (Transaction)")
	} else {
		// Check if the data sent is actually another Struct.
		derivedStruct, ok := stx.Base.(*types.Transaction).Data.(map[string]interface{})

		// If its not, we can simply return and the different transaction types will get the value themselves.
		if !ok {
			return stx, tx, nil
		}

		// types.RequestUpload
		if cid, ok := derivedStruct["cid"]; ok {
			if ipfsNode, ok := derivedStruct["ipfsNode"]; ok {
				reqUpload := &types.RequestUpload{Cid:cid.(string), IpfsNode:ipfsNode.(string)}
				stx.Base.(*types.Transaction).Data = reqUpload
				tx.Data = reqUpload
			}
		}

		return stx, tx, nil
	}
}
func (app *Application) CheckTx(txBytes []byte) abci.ResponseCheckTx { //types.Result {
	//stx := &crypto.SignedStruct{Base: &types.Transaction{}}
	stx, tx, err := app.unmarshalTransaction(txBytes)
	if err != nil {
		return abci.ResponseCheckTx{Code: uint32(types.CodeType_InternalError), Log: err.Error()}
	}
	fmt.Printf("Hash of transaction: %s\n",crypto.HashStruct(tx))

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
	case types.TransactionType_DownloadData:
		{
			fmt.Println("CheckTx_DownloadData")
			return *app.CheckTx_DownloadData(signer, tx)
		}

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
			// TODO: Remove this type?
			return abci.ResponseCheckTx{Info: "Error"};
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

func (app *Application) Commit() abci.ResponseCommit { //types.Result {
	fmt.Println("Commit trigger");
	return abci.ResponseCommit{}
}

func (app *Application) Query(reqQuery abci.RequestQuery) (abci.ResponseQuery) {
	fmt.Println("Query trigger");
	log.Print("query")
	switch reqQuery.Path {
	case "/newsample":
		{
			return *app.Query_Newsample(reqQuery)
		}
	default:
		{
			return abci.ResponseQuery{Code: uint32(types.CodeType_BCFSInvalidInput), Log: "Invalid query path."}
		}
	}
/*
	sumBlock := app.EndBlock(abci.RequestEndBlock{Height:-1})
	fmt.Printf("%+v\n", sumBlock.GetValidatorUpdates())
	return*/
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