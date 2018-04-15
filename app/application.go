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
)

type Application struct {
	abci.BaseApplication

	info 				string
	uploadAddr 			string

	tempUploads 		map[string]bool
	seenTranc 			map[string]bool

	IpfsHttpClient		*http.Client
	TMRpcClients		map[string]rpcClient.Client

}

func NewApplication() *Application {
	app := &Application{info: conf.AppConfig().Info, uploadAddr: conf.AppConfig().UploadAddr,
		tempUploads: make(map[string]bool), seenTranc: make(map[string]bool),
		IpfsHttpClient: &http.Client{Timeout: time.Duration(conf.AppConfig().IpfsProxyTimeoutSeconds) * time.Second}}
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
	stx := &crypto.SignedStruct{}
	var tx types.Transaction
	var ok bool = false
	if err := json.Unmarshal(txBytes, stx); err != nil {
		return abci.ResponseDeliverTx{Code: uint32(types.CodeType_InternalError), Log: err.Error()}
	} else if tx, ok = stx.Base.(types.Transaction); !ok {
		return abci.ResponseDeliverTx{Code: uint32(types.CodeType_InternalError), Log: "Could not Marshal transaction (Transaction)"}
	}
	fmt.Printf("Hash of transaction: %s\n",crypto.HashStruct(tx))
	switch tx.Type {
	case types.TransactionType_DownloadData:
		{
			/*if err := deliverAccountAddTransaction(tx, app.state); err != nil {
				return types.Result{Code: types.CodeType_BaseInvalidInput, Log: err.Error()}
			}*/
			fmt.Println("Downloaddata")
			return abci.ResponseDeliverTx{Info: "Error"};
		}

	case types.TransactionType_UploadData:
		{
			/*if err := deliverAccountDelTransaction(tx, app.state); err != nil {
				return types.Result{Code: types.CodeType_BaseInvalidInput, Log: err.Error()}
			}*/
			return abci.ResponseDeliverTx{Info: "Error"};
		}
	case types.TransactionType_RemoveData:
		{
			/*if err := deliverReputationGiveTransaction(tx, app.state); err != nil {
				return types.Result{Code: types.CodeType_BaseInvalidInput, Log: err.Error()}
			}*/
			return abci.ResponseDeliverTx{Info: "Error"};
		}
	case types.TransactionType_VerifyStorage:
		{
			/*if err := deliverSecretAddTransaction(tx, app.state); err != nil {
				return types.Result{Code: types.CodeType_BaseInvalidInput, Log: err.Error()}
			}*/
			return abci.ResponseDeliverTx{Info: "Error"};
		}
	case types.TransactionType_ChangeContentAccess:
		{
			/*if err := deliverSecretUpdateTransaction(tx, app.state); err != nil {
				return types.Result{Code: types.CodeType_BaseInvalidInput, Log: err.Error()}
			}*/
			return abci.ResponseDeliverTx{Info: "Error"};
		}
	default:
		{
			//return types.Result{Code: types.CodeType_BaseInvalidInput, Log: "unknown transaction type"}
			return abci.ResponseDeliverTx{Info: "Error"};
		}
	}
	//return types.OK
	return abci.ResponseDeliverTx{Info: "All good"};
}


func (app *Application) CheckTx(txBytes []byte) abci.ResponseCheckTx { //types.Result {
	txHash, _ := crypto.IPFSHashData(txBytes)
	fmt.Println("CheckTx trigger. Hash of data: " + txHash);
	fmt.Println("Data received: " + string(txBytes))
	stx := &crypto.SignedStruct{}
	var tx types.Transaction
	var ok bool = false
	if err := json.Unmarshal(txBytes, stx); err != nil {
		return abci.ResponseCheckTx{Code: uint32(types.CodeType_InternalError), Log: err.Error()}
	} else if tx, ok = stx.Base.(types.Transaction); !ok {
		return abci.ResponseCheckTx{Code: uint32(types.CodeType_InternalError), Log: "Could not Marshal transaction (Transaction)"}
	}
	fmt.Printf("Hash of transaction: %s\n",crypto.HashStruct(tx))

	// Get access list
	identity, ok := GetAccessList().Identities[tx.Identity];
	if !ok {
		return abci.ResponseCheckTx{Code: uint32(types.CodeType_Unauthorized), Log: "Could not get access list"}
	}

	// Check if public key exists and if message is signed.
	if pk, err := crypto.LoadPublicKey(conf.AppConfig().BasePath + conf.AppConfig().PublicKeys + identity.PublicKey); err != nil {
		return abci.ResponseCheckTx{Code: uint32(types.CodeType_BCFSInvalidSignature), Log: "Could not locate public key"}
	} else if !stx.Verify(pk) {
		return abci.ResponseCheckTx{Code: uint32(types.CodeType_BCFSInvalidSignature), Log: "Could not verify signature"}
	}

	switch tx.Type {
	case types.TransactionType_DownloadData:
		{
			/*if err := checkAccountAddTransaction(tx, app.state); err != nil {
				return types.Result{Code: types.CodeType_BaseInvalidInput, Log: err.Error()}
			}*/
			fmt.Println("DownloadData")
			return abci.ResponseCheckTx{Info: "Error"}
		}

	case types.TransactionType_UploadData:
		{
			// Check if uploader is allowed to upload data.
			if identity.AccessLevel < 1 {
				return abci.ResponseCheckTx{Code: uint32(types.CodeType_Unauthorized), Log: "Insufficient access level"}
			}

			// Check if data hash is contained within the transaction.
			reqUpload, ok := tx.Data.(types.RequestUpload)
			if !ok {
				return abci.ResponseCheckTx{Code: uint32(types.CodeType_BCFSInvalidInput), Log: "Could not type assert Data to string"}
			}

			// Check if a file with this hash exists on an IPFS node and is uploaded to our server.
			

			// Check if data hash is already in the list of uploads pending
			if val, ok := app.tempUploads[dataHash]; ok && val {
				// Check if data is stored on disk. Return CodeType_OK
				return abci.ResponseCheckTx{Code: uint32(types.CodeType_BCFSInvalidInput), Log: "Data hash is already in the list of pending uploads"}
			} else {
				// Add data hash to the list of pending uploads
				app.tempUploads[dataHash] = true

				fmt.Println("UploadData")
				return abci.ResponseCheckTx{Code: uint32(types.CodeType_BCFSBeginUploadOK), Log: "Data hash added to list of pending uploads"}
			}
		}
	case types.TransactionType_RemoveData:
		{
			/*if err := checkReputationGiveTransaction(tx, app.state); err != nil {
				return types.Result{Code: types.CodeType_BaseInvalidInput, Log: err.Error()}
			}*/
			fmt.Println("RemoveData")
			return abci.ResponseCheckTx{Info: "Error"}
		}
	case types.TransactionType_VerifyStorage:
		{
			/*if err := checkSecretAddTransaction(tx, app.state); err != nil {
				return types.Result{Code: types.CodeType_BaseInvalidInput, Log: err.Error()}
			}*/
			fmt.Println("VerifyStorage")
			return abci.ResponseCheckTx{Info: "Error"};
		}
	case types.TransactionType_ChangeContentAccess:
		{
			/*if err := checkSecretUpdateTransaction(tx, app.state); err != nil {
				return types.Result{Code: types.CodeType_BaseInvalidInput, Log: err.Error()}
			}*/
			fmt.Println("ChangeContentAccess")
			return abci.ResponseCheckTx{Info: "Error"}
		}
	default:
		{
			//return types.Result{Code: types.CodeType_BaseInvalidInput, Log: "unknown transaction type"}
			return abci.ResponseCheckTx{Info: "Error"}
		}
	}
	//return types.OK
	return abci.ResponseCheckTx{Info: "All good", Code: uint32(types.CodeType_OK)}
}

func (app *Application) Commit() abci.ResponseCommit { //types.Result {
	fmt.Println("Commit trigger");
	return abci.ResponseCommit{}
}

func (app *Application) Query(reqQuery abci.RequestQuery) (resQuery abci.ResponseQuery) {
	fmt.Println("Query trigger");
	log.Print("query")
	switch reqQuery.Path {
	case "/account":
		{
			var (
				result interface{}
				err    error
			)
			if reqQuery.Data == nil {
				log.Printf("got account list: %+v", result)
			} else {
				log.Printf("got account: %+v", result)
			}
			if err != nil {
				resQuery.Code = 1 // types.CodeType_BaseInvalidInput
				resQuery.Log = err.Error()
				return
			}
			bs, _ := json.Marshal(result)
			resQuery.Value = bs
		}
	case "/secret":
		{
			var (
				result interface{}
				err    error
			)
			if reqQuery.Data == nil {
				log.Printf("got secret list: %+v", result)
			} else {
				log.Printf("got secret: %+v", result)
			}
			if err != nil {
				resQuery.Code = 1 //types.CodeType_BaseInvalidInput
				resQuery.Log = err.Error()
				return
			}
			bs, _ := json.Marshal(result)
			resQuery.Value = bs
		}
	default:
		{
			resQuery.Code = 1 //types.CodeType_BaseInvalidInput
			resQuery.Log = "wrong path"
			return
		}
	}

	sumBlock := app.EndBlock(abci.RequestEndBlock{Height:-1})
	fmt.Printf("%+v\n", sumBlock.GetValidatorUpdates())
	return
}