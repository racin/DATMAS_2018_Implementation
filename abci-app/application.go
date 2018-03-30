package app

import (
	"encoding/json"
	"log"
	"io"
	"github.com/racin/DATMAS_2018_Implementation/crypto"
	"github.com/tendermint/abci/types"
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
}

func NewApplication(uploadAddr string) *Application {
	// tree : iavl.NewIAVLTree(0, nil)
	return &Application{info: "____racin", uploadAddr: uploadAddr}
}

func (app *Application) StartUploadHandler(){
	http.HandleFunc("/", app.UploadHandler)
	http.ListenAndServe(app.uploadAddr, nil)
}
func (app *Application) UploadHandler(w http.ResponseWriter, r *http.Request){
	err := r.ParseMultipartForm(200000) // grab the multipart form
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}

	formdata := r.MultipartForm // ok, no problem so far, read the Form data

	//get the *fileheaders
	files := formdata.File["multiplefiles"] // grab the filenames

	for i, _ := range files { // loop through the files one by one
		file, err := files[i].Open()
		defer file.Close()
		if err != nil {
			fmt.Fprintln(w, err)
			return
		}

		out, err := os.Create("/tmp/" + files[i].Filename)

		defer out.Close()
		if err != nil {
			fmt.Fprintf(w, "Unable to create the file for writing. Check your write access privilege")
			return
		}

		_, err = io.Copy(out, file) // file not files[i] !

		if err != nil {
			fmt.Fprintln(w, err)
			return
		}

		fmt.Fprintf(w, "Files uploaded successfully : ")
		fmt.Fprintf(w, files[i].Filename+"\n")

	}
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
			fmt.Println("UploadData")
			return types.ResponseCheckTx{Info: "Error"}
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
	return
}
