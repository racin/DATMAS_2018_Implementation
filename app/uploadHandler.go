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

func (app *Application) HasSeenTranc(trancHash string) bool{
	if _, ok := app.seenTranc[trancHash]; ok {
		return true;
	}
	return false;
}

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

	stx := &crypto.SignedStruct{}
	var tx types.Transaction
	if err := json.Unmarshal([]byte(txString[0]), tx); err != nil {
		writeUploadResponse(&w, types.CodeType_BCFSInvalidInput, "Could not Marshal transaction (SignedTransaction)");
		return
	} else if tx, ok = stx.Base.(types.Transaction); !ok {
		writeUploadResponse(&w, types.CodeType_BCFSInvalidInput, "Could not Marshal transaction (Transaction)");
		return
	}

	// Check for replay attack
	txHash := crypto.HashStruct(tx)
	if app.HasSeenTranc(txHash) {
		writeUploadResponse(&w, types.CodeType_BadNonce, "Could not process transaction. Possible replay attack.");
		return
	}

	// Check identity access
	identity, ok := GetAccessList().Identities[tx.Identity];
	if !ok {
		writeUploadResponse(&w, types.CodeType_Unauthorized, "Could not get access list");
		return
	}

	// Check if public key exists and if message is signed.
	if pk, err := crypto.LoadPublicKey(conf.AppConfig().BasePath + conf.AppConfig().PublicKeys + identity.PublicKey); err != nil {
		writeUploadResponse(&w, types.CodeType_BCFSInvalidSignature, "Could not locate public key");
		return
	} else if pk.Verify(txHash, stx.Signature) {
		writeUploadResponse(&w, types.CodeType_BCFSInvalidSignature, "Could not verify signature");
		return
	}

	// Check if uploader is allowed to upload data.
	if identity.AccessLevel < 1 {
		writeUploadResponse(&w, types.CodeType_Unauthorized, "Insufficient access level");
		return
	}

	// Check if data hash is contained within the transaction.
	fileHash, ok := tx.Data.(string)
	if (fileHash == "" || !ok) {
		writeUploadResponse(&w, types.CodeType_BCFSInvalidInput, "Missing data hash parameter.");
		return
	}

	/*// Check if data hash is already in the list of uploads pending
	if _, ok := app.tempUploads[fileHash]; !ok {
		writeUploadResponse(&w, types.CodeType_BCFSInvalidInput, "Data hash not in the list of pending uploads.");
		return // Data hash not in the list of pending uploads
	}*/

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
	fileBytes, err := ioutil.ReadAll(fopen);
	if err != nil {
		writeUploadResponse(&w, types.CodeType_BCFSInvalidInput, "Could not get byte array of input file.");
		return
	} else if uplFileHash, err := crypto.IPFSHashData(fileBytes); err != nil {
		writeUploadResponse(&w, types.CodeType_BCFSInvalidInput, "Could not get hash of input file.");
		return
	} else if uplFileHash != fileHash {
		writeUploadResponse(&w, types.CodeType_BCFSInvalidInput, "Hash of input file not present in transaction.");
		return
	}

	out, err := os.Create(conf.AppConfig().TempUploadPath + fileHash)
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

	// Generate Sample of data. Distribute it to other TM nodes
	sample := crypto.GenerateStorageSample(&fileBytes)

	// Store the sample in Query.
	/*if err := sample.StoreSample(conf.AppConfig().BasePath + conf.AppConfig().StorageSamples); err != nil {
		writeUploadResponse(&w, types.CodeType_InternalError, err.Error());
		return
	}*/
	myPrivKey, err := crypto.LoadPrivateKey(conf.AppConfig().BasePath + conf.AppConfig().PublicKeys +
		conf.AppConfig().PrivateKey);
	if err != nil {
		writeUploadResponse(&w, types.CodeType_BCFSInvalidSignature, "Could not locate public key");
		return
	}

	signedSample, err := crypto.SignStruct(sample, myPrivKey)
	if err != nil {
		writeUploadResponse(&w, types.CodeType_BCFSInvalidSignature, "Could not locate public key");
		return
	}
	app.broadcastQuery("/newsample", signedSample)


	writeUploadResponse(&w, types.CodeType_OK, "Files uploaded successfully : " + file.Filename);

	//app.BaseApplication.CheckTx(tx)
	// TODO: Replay transaction to CheckTx?
	// CheckTx issues a challenge to verify that file is stored. Then Pins and Delivers and Commits.
}
