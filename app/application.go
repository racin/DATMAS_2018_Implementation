package app

import (
	"encoding/json"
	"log"
	"io"
	"github.com/racin/DATMAS_2018_Implementation/crypto"
	"github.com/racin/DATMAS_2018_Implementation/types"
	abci "github.com/tendermint/abci/types"
	//mp "github.com/tendermint/tendermint/mempool"
	//"github.com/tendermint/merkleeyes/iavl"
	"fmt"
	"net/http"
	"os"
	conf "github.com/racin/DATMAS_2018_Implementation/configuration"

)

type Application struct {
	abci.BaseApplication

	info string
	//tree *iavl.IAVLTree
	uploadAddr string

	tempUploads map[string]bool
}

func NewApplication() *Application {
	// tree : iavl.NewIAVLTree(0, nil)
	return &Application{info: conf.AppConfig().Info, uploadAddr: conf.AppConfig().UploadAddr, tempUploads: make(map[string]bool)}
}

func (app *Application) StartUploadHandler(){
	http.HandleFunc(conf.AppConfig().UploadEndpoint, app.UploadHandler)
	if err := http.ListenAndServe(app.uploadAddr, nil); err != nil {
		panic("Error setting up upload handler. Error: " + err.Error())
	}
}

func (app *Application) UploadHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(104857600) // Up to 100MB stored in memory.
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}
	formdata := r.MultipartForm // ok, no problem so far, read the Form data

	if val, ok := formdata.Value["Status"]; ok {
		byteArr, _ := json.Marshal(&types.ResponseUpload{Message:val[0], Codetype:types.CodeType_OK})
		fmt.Fprintf(w, "%s", byteArr)
		return
	}

	txString, ok := formdata.Value["transaction"]
	if !ok {
		byteArr, _ := json.Marshal(&types.ResponseUpload{Message:"Missing transaction parameter.", Codetype:types.CodeType_BCFSInvalidInput})
		fmt.Fprintf(w, "%s", byteArr)
		return
	}

	tx := &SignedTransaction{}
	if err := json.Unmarshal([]byte(txString[0]), tx); err != nil {
		byteArr, _ := json.Marshal(&types.ResponseUpload{Message:"Could not Marshal transaction", Codetype:types.CodeType_BCFSInvalidInput})
		fmt.Fprintf(w, "%s", byteArr)
		return
	}

	fileHash, ok := tx.Data.(string)
	if (!ok) {
		byteArr, _ := json.Marshal(&types.ResponseUpload{Message:"Missing data hash parameter.", Codetype:types.CodeType_BCFSInvalidInput})
		fmt.Fprintf(w, "%s", byteArr)
		return
	}
	if _, ok := app.tempUploads[fileHash]; !ok {
		byteArr, _ := json.Marshal(&types.ResponseUpload{Message:"Data hash not in the list of pending uploads.",
			Codetype:types.CodeType_BCFSInvalidInput})
		fmt.Fprintf(w, "%s", byteArr)
		return // Data hash not in the list of pending uploads
	}

	files, ok := formdata.File["file"]
	if !ok || len(files) > 1 {
		byteArr, _ := json.Marshal(&types.ResponseUpload{Message:"File parameter should contain exactly one file.",
			Codetype:types.CodeType_BCFSInvalidInput})
		fmt.Fprintf(w, "%s", byteArr)
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
		byteArr, _ := json.Marshal(&types.ResponseUpload{Message:"Unable to create the file for writing." +
			" Check your write access privilege", Codetype:types.CodeType_Unauthorized})
		fmt.Fprintf(w, "%s", byteArr)
		return
	}

	_, err = io.Copy(out, fopen) // file not files[i] !

	if err != nil {
		byteArr, _ := json.Marshal(&types.ResponseUpload{Message:err.Error(), Codetype:types.CodeType_InternalError})
		fmt.Fprintf(w, "%s", byteArr)
		return
	}

	byteArr, _ := json.Marshal(&types.ResponseUpload{Message:"Files uploaded successfully : " + file.Filename, Codetype:types.CodeType_OK})
	fmt.Fprintf(w, "%s", byteArr)

	// Replay transaction to CheckTx?
}
func (app *Application) Info(abci.RequestInfo) (resInfo abci.ResponseInfo) {
	fmt.Println("Info trigger");
	return abci.ResponseInfo{Data: app.info}
}
func (app *Application) DeliverTx(txBytes []byte)  abci.ResponseDeliverTx {
	txHash, _ := crypto.IPFSHashData(txBytes)
	fmt.Println("Deliver trigger. Hash of data: " + txHash);
	tx := &Transaction{}
	tx.Data = "abc"
	if err := json.Unmarshal(txBytes, tx); err != nil {
		return abci.ResponseDeliverTx{Info: "Error"}
	}
	fmt.Printf("Hash of transaction: %s\n",tx.Hash())
	switch tx.Type {
	case DownloadData:
		{
			/*if err := deliverAccountAddTransaction(tx, app.state); err != nil {
				return types.Result{Code: types.CodeType_BaseInvalidInput, Log: err.Error()}
			}*/
			fmt.Println("Downloaddata")
			return abci.ResponseDeliverTx{Info: "Error"};
		}

	case UploadData:
		{
			/*if err := deliverAccountDelTransaction(tx, app.state); err != nil {
				return types.Result{Code: types.CodeType_BaseInvalidInput, Log: err.Error()}
			}*/
			return abci.ResponseDeliverTx{Info: "Error"};
		}
	case RemoveData:
		{
			/*if err := deliverReputationGiveTransaction(tx, app.state); err != nil {
				return types.Result{Code: types.CodeType_BaseInvalidInput, Log: err.Error()}
			}*/
			return abci.ResponseDeliverTx{Info: "Error"};
		}
	case VerifyStorage:
		{
			/*if err := deliverSecretAddTransaction(tx, app.state); err != nil {
				return types.Result{Code: types.CodeType_BaseInvalidInput, Log: err.Error()}
			}*/
			return abci.ResponseDeliverTx{Info: "Error"};
		}
	case ChangeContentAccess:
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
	tx := &SignedTransaction{}
	if err := json.Unmarshal(txBytes, tx); err != nil {
		fmt.Println(err.Error())
		return abci.ResponseCheckTx{Info: "Error"}
	}
	fmt.Printf("Hash of transaction: %s\n",tx.Hash())

	acl := GetAccessList()
	identity, ok := acl.Identities[tx.Identity];
	if !ok {
		return abci.ResponseCheckTx{Code: uint32(types.CodeType_Unauthorized), Log: "Could not get access list"}
	}

	// Check if public key exists and if message is signed.
	pk, err := crypto.LoadPublicKey(conf.AppConfig().BasePath + conf.AppConfig().PublicKeys + identity.PublicKey)
	if err != nil {
		return abci.ResponseCheckTx{Code: uint32(types.CodeType_BCFSInvalidPubKey), Log: "Could not locate public key"}
	}

	// Check if transaction is signed.
	if !pk.Verify(tx.Hash(), tx.Signature) {
		return abci.ResponseCheckTx{Code: uint32(types.CodeType_BCFSInvalidSignature), Log: "Could not verify signature"}
	}

	switch tx.Type {
	case DownloadData:
		{
			/*if err := checkAccountAddTransaction(tx, app.state); err != nil {
				return types.Result{Code: types.CodeType_BaseInvalidInput, Log: err.Error()}
			}*/
			fmt.Println("DownloadData")
			return abci.ResponseCheckTx{Info: "Error"}
		}

	case UploadData:
		{
			// Check if uploader is allowed to upload data.
			if identity.AccessLevel < 1 {
				return abci.ResponseCheckTx{Code: uint32(types.CodeType_Unauthorized), Log: "Insufficient access level"}
			}

			dataHash, ok := tx.Data.(string)
			if !ok {
				return abci.ResponseCheckTx{Code: uint32(types.CodeType_BCFSInvalidInput), Log: "Could not type assert Data to string"}
			}

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
	case RemoveData:
		{
			/*if err := checkReputationGiveTransaction(tx, app.state); err != nil {
				return types.Result{Code: types.CodeType_BaseInvalidInput, Log: err.Error()}
			}*/
			fmt.Println("RemoveData")
			return abci.ResponseCheckTx{Info: "Error"}
		}
	case VerifyStorage:
		{
			/*if err := checkSecretAddTransaction(tx, app.state); err != nil {
				return types.Result{Code: types.CodeType_BaseInvalidInput, Log: err.Error()}
			}*/
			fmt.Println("VerifyStorage")
			return abci.ResponseCheckTx{Info: "Error"};
		}
	case ChangeContentAccess:
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