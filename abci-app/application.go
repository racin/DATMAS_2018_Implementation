package app

import (
	"encoding/json"
	"log"
	"io"
	"github.com/racin/DATMAS_2018_Implementation/crypto"
	"github.com/tendermint/abci/types"
	//mp "github.com/tendermint/tendermint/mempool"
	//"github.com/tendermint/merkleeyes/iavl"
	"fmt"
	"net/http"
	"os"
)

type Application struct {
	types.BaseApplication

	info string
	//tree *iavl.IAVLTree
	uploadAddr string

	tempUploads map[string]bool
}



func NewApplication(uploadAddr string) *Application {
	// tree : iavl.NewIAVLTree(0, nil)
	return &Application{info: "____racin", uploadAddr: uploadAddr, tempUploads: make(map[string]bool)}
}

func (app *Application) StartUploadHandler(){
	http.HandleFunc("/", app.UploadHandler)
	http.ListenAndServe(app.uploadAddr, nil)
}
func (app *Application) UploadHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(104857600) // Up to 100MB stored in memory.
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}

	formdata := r.MultipartForm // ok, no problem so far, read the Form data

	datahash, ok := formdata.Value["datahash"]
	if !ok {
		return // Missing data hash
	}
	if _, ok := app.tempUploads[datahash[0]]; !ok {
		return // Data hash not in the list of pending uploads
	}

	files, ok := formdata.File["multiplefiles"]
	if !ok || len(files) > 1 {
		return // Missing files or more than one file
	}


	file := files[0]
	fopen, err := file.Open()
	defer fopen.Close()
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}

	out, err := os.Create("/tmp/" + file.Filename)

	defer out.Close()
	if err != nil {
		fmt.Fprintf(w, "Unable to create the file for writing. Check your write access privilege")
		return
	}

	_, err = io.Copy(out, fopen) // file not files[i] !

	if err != nil {
		fmt.Fprintln(w, err)
		return
	}

	fmt.Fprintf(w, "Files uploaded successfully : ")
	fmt.Fprintf(w, file.Filename+"\n")
}
func (app *Application) Info(types.RequestInfo) (resInfo types.ResponseInfo) {
	fmt.Println("Info trigger");
	return types.ResponseInfo{Data: app.info}
}
func (app *Application) DeliverTx(txBytes []byte)  types.ResponseDeliverTx {
	txHash, _ := crypto.IPFSHashData(txBytes)
	fmt.Println("Deliver trigger. Hash of data: " + txHash);
	tx := &Transaction{}
	if err := json.Unmarshal(txBytes, tx); err != nil {
		return types.ResponseDeliverTx{Info: "Error"}
	}
	fmt.Printf("Hash of transaction: %s\n",tx.Hash())
	switch tx.Type {
	case DownloadData:
		{
			/*if err := deliverAccountAddTransaction(tx, app.state); err != nil {
				return types.Result{Code: types.CodeType_BaseInvalidInput, Log: err.Error()}
			}*/
			fmt.Println("Downloaddata")
			return types.ResponseDeliverTx{Info: "Error"};
		}

	case UploadData:
		{
			/*if err := deliverAccountDelTransaction(tx, app.state); err != nil {
				return types.Result{Code: types.CodeType_BaseInvalidInput, Log: err.Error()}
			}*/
			return types.ResponseDeliverTx{Info: "Error"};
		}
	case RemoveData:
		{
			/*if err := deliverReputationGiveTransaction(tx, app.state); err != nil {
				return types.Result{Code: types.CodeType_BaseInvalidInput, Log: err.Error()}
			}*/
			return types.ResponseDeliverTx{Info: "Error"};
		}
	case VerifyStorage:
		{
			/*if err := deliverSecretAddTransaction(tx, app.state); err != nil {
				return types.Result{Code: types.CodeType_BaseInvalidInput, Log: err.Error()}
			}*/
			return types.ResponseDeliverTx{Info: "Error"};
		}
	case ChangeContentAccess:
		{
			/*if err := deliverSecretUpdateTransaction(tx, app.state); err != nil {
				return types.Result{Code: types.CodeType_BaseInvalidInput, Log: err.Error()}
			}*/
			return types.ResponseDeliverTx{Info: "Error"};
		}
	default:
		{
			//return types.Result{Code: types.CodeType_BaseInvalidInput, Log: "unknown transaction type"}
			return types.ResponseDeliverTx{Info: "Error"};
		}
	}
	//return types.OK
	return types.ResponseDeliverTx{Info: "All good"};
}

func (app *Application) CheckTx(txBytes []byte) types.ResponseCheckTx { //types.Result {
	txHash, _ := crypto.IPFSHashData(txBytes)
	fmt.Println("CheckTx trigger. Hash of data: " + txHash);
	fmt.Println("Data received: " + string(txBytes))
	tx := &Transaction{}
	if err := json.Unmarshal(txBytes, tx); err != nil {
		fmt.Println(err.Error())
		return types.ResponseCheckTx{Info: "Error"}
	}
	fmt.Printf("Hash of transaction: %s\n",tx.Hash())
	switch tx.Type {
	case DownloadData:
		{
			/*if err := checkAccountAddTransaction(tx, app.state); err != nil {
				return types.Result{Code: types.CodeType_BaseInvalidInput, Log: err.Error()}
			}*/
			fmt.Println("DownloadData")
			return types.ResponseCheckTx{Info: "Error"}
		}

	case UploadData:
		{

			/*if err := checkAccountDelTransaction(tx, app.state); err != nil {
				return types.Result{Code: types.CodeType_BaseInvalidInput, Log: err.Error()}
			}*/


			// Check if data hash is already in the list of uploads pending
			dataHash, ok := tx.Data.(string)
			if !ok {
				return types.ResponseCheckTx{Info: "Could not type assert Data to string"}
			} else if _, ok := app.tempUploads[dataHash]; ok {
				return types.ResponseCheckTx{Info: "Data hash is already in the list of pending uploads"}
			}

			// Check if uploader is allowed to upload data.
			acl := GetAccessList()
			identity, ok := acl.Identities[tx.Identity];

			if !ok || identity.AccessLevel < 1{
				return types.ResponseCheckTx{Info: "Insufficient access level"}
			}

			// Check if public key exists and if message is signed.
			pk, err := crypto.LoadPublicKey(identity.KeyPath)
			if err != nil {
				return types.ResponseCheckTx{Info: "Could not locate public key"}
			}

			// Check if transaction is signed.
			if !pk.Verify(tx.Hash(), tx.Signature) {
				return types.ResponseCheckTx{Info: "Could not verify signature"}
			}

			// Add data hash to the list of pending uploads
			app.tempUploads[dataHash] = true

			fmt.Println("UploadData")
			return types.ResponseCheckTx{Code: types.CodeTypeOK, Info: "Data hash added to list of pending uploads"}
		}
	case RemoveData:
		{
			/*if err := checkReputationGiveTransaction(tx, app.state); err != nil {
				return types.Result{Code: types.CodeType_BaseInvalidInput, Log: err.Error()}
			}*/
			fmt.Println("RemoveData")
			return types.ResponseCheckTx{Info: "Error"}
		}
	case VerifyStorage:
		{
			/*if err := checkSecretAddTransaction(tx, app.state); err != nil {
				return types.Result{Code: types.CodeType_BaseInvalidInput, Log: err.Error()}
			}*/
			fmt.Println("VerifyStorage")
			return types.ResponseCheckTx{Info: "Error"};
		}
	case ChangeContentAccess:
		{
			/*if err := checkSecretUpdateTransaction(tx, app.state); err != nil {
				return types.Result{Code: types.CodeType_BaseInvalidInput, Log: err.Error()}
			}*/
			fmt.Println("ChangeContentAccess")
			return types.ResponseCheckTx{Info: "Error"}
		}
	default:
		{
			//return types.Result{Code: types.CodeType_BaseInvalidInput, Log: "unknown transaction type"}
			return types.ResponseCheckTx{Info: "Error"}
		}
	}
	//return types.OK
	return types.ResponseCheckTx{Info: "All good", Code: types.CodeTypeOK}
}

func (app *Application) Commit() types.ResponseCommit { //types.Result {
	fmt.Println("Commit trigger");
	return types.ResponseCommit{}
}

func (app *Application) Query(reqQuery types.RequestQuery) (resQuery types.ResponseQuery) {
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

	sumBlock := app.EndBlock(types.RequestEndBlock{Height:-1})
	fmt.Printf("%+v\n", sumBlock.GetValidatorUpdates())
	return
}
