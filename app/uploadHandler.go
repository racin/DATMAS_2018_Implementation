package app

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	conf "github.com/racin/DATMAS_2018_Implementation/configuration"
	"github.com/racin/DATMAS_2018_Implementation/crypto"
	"fmt"
	"github.com/racin/DATMAS_2018_Implementation/types"
)
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

	// CHeck identity access
	acl := GetAccessList()
	identity, ok := acl.Identities[tx.Identity];
	if !ok {
		byteArr, _ := json.Marshal(&types.ResponseUpload{Message:"Could not Marshal transaction", Codetype:types.CodeType_BCFSInvalidInput})
		fmt.Fprintf(w, "%s", byteArr)
		return
		return abci.ResponseCheckTx{Code: uint32(types.CodeType_Unauthorized), Log: "Could not get access list"}
	}

	// Check if public key exists and if message is signed.
	pk, err := crypto.LoadPublicKey(conf.AppConfig().BasePath + conf.AppConfig().PublicKeys + identity.PublicKey)
	if err != nil {
		byteArr, _ := json.Marshal(&types.ResponseUpload{Message:"Could not Marshal transaction", Codetype:types.CodeType_BCFSInvalidInput})
		fmt.Fprintf(w, "%s", byteArr)
		return
		return abci.ResponseCheckTx{Code: uint32(types.CodeType_BCFSInvalidPubKey), Log: "Could not locate public key"}
	}

	// Check if transaction is signed.
	if !pk.Verify(tx.Hash(), tx.Signature) {
		byteArr, _ := json.Marshal(&types.ResponseUpload{Message:"Could not Marshal transaction", Codetype:types.CodeType_BCFSInvalidInput})
		fmt.Fprintf(w, "%s", byteArr)
		return
		return abci.ResponseCheckTx{Code: uint32(types.CodeType_BCFSInvalidSignature), Log: "Could not verify signature"}
	}

	// Check if uploader is allowed to upload data.
	if identity.AccessLevel < 1 {
		byteArr, _ := json.Marshal(&types.ResponseUpload{Message:"Could not Marshal transaction", Codetype:types.CodeType_BCFSInvalidInput})
		fmt.Fprintf(w, "%s", byteArr)
		return
		return abci.ResponseCheckTx{Code: uint32(types.CodeType_Unauthorized), Log: "Insufficient access level"}
	}

	// Check if data hash is contained within the transaction.
	fileHash, ok := tx.Data.(string)
	if (!ok) {
		byteArr, _ := json.Marshal(&types.ResponseUpload{Message:"Missing data hash parameter.", Codetype:types.CodeType_BCFSInvalidInput})
		fmt.Fprintf(w, "%s", byteArr)
		return
	}

	// Check if data hash is already in the list of uploads pending
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

	// Check if the hash of the upload file equals the hash contained in the transaction
	fileBytes, err := ioutil.ReadAll(fopen)
	if err != nil {
		byteArr, _ := json.Marshal(&types.ResponseUpload{Message:"Missing data hash parameter.", Codetype:types.CodeType_BCFSInvalidInput})
		fmt.Fprintf(w, "%s", byteArr)
		return
	}

	uplFileHash, err := crypto.IPFSHashData(fileBytes)
	if err != nil {
		byteArr, _ := json.Marshal(&types.ResponseUpload{Message:"Missing data hash parameter.", Codetype:types.CodeType_BCFSInvalidInput})
		fmt.Fprintf(w, "%s", byteArr)
		return
	}

	if uplFileHash != fileHash {
		byteArr, _ := json.Marshal(&types.ResponseUpload{Message:"Missing data hash parameter.", Codetype:types.CodeType_BCFSInvalidInput})
		fmt.Fprintf(w, "%s", byteArr)
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
