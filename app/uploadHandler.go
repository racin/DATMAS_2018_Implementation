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

func writeUploadResponse(w *http.ResponseWriter, codeType types.CodeType, message string){
	//json.NewEncoder(*w).Encode(&types.ResponseUpload{Message:message, Codetype:codeType})
	byteArr, _ := json.Marshal(&types.ResponseUpload{Message:message, Codetype:codeType})
	fmt.Fprintf(*w, "%s", byteArr)
}

func (app *Application) UploadHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(104857600) // Up to 100MB stored in memory.
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}
	formdata := r.MultipartForm // ok, no problem so far, read the Form data

	if val, ok := formdata.Value["Status"]; ok {
		writeUploadResponse(&w, types.CodeType_OK, val[0]);
		return
	}

	txString, ok := formdata.Value["transaction"]
	if !ok {
		writeUploadResponse(&w, types.CodeType_BCFSInvalidInput, "Missing transaction parameter.");
		return
	}

	tx := &SignedTransaction{}
	if err := json.Unmarshal([]byte(txString[0]), tx); err != nil {
		writeUploadResponse(&w, types.CodeType_BCFSInvalidInput, "Could not Marshal transaction");
		return
	}

	// Check identity access
	identity, ok := GetAccessList().Identities[tx.Identity];
	if !ok {
		writeUploadResponse(&w, types.CodeType_Unauthorized, "Could not get access list");
		return
	}

	// Verify signature in transaction.
	if ok, msg := verifySignature(&identity, tx); !ok {
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

	// TODO: Replay transaction to CheckTx?
}
