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
	"time"
)

// Used to prevent replay attacks.
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
	formdata := r.MultipartForm

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
	identity, ok := app.GetAccessList().Identities[tx.Identity];
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
	if fileHash == "" || !ok {
		writeUploadResponse(&w, types.CodeType_BCFSInvalidInput, "Missing data hash parameter.");
		return
	}

	// Check that exactly one file is sent
	files, ok := formdata.File["file"]
	if !ok || len(files) > 1 {
		writeUploadResponse(&w, types.CodeType_BCFSInvalidInput, "File parameter should contain exactly one file.");
		return
	}

	// Try to open the file sent
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

	// Create file on disk in temporary storage.
	// TODO: Confirm the necessity of this step.
	out, err := os.Create(conf.AppConfig().TempUploadPath + fileHash)
	defer out.Close()
	if err != nil {
		writeUploadResponse(&w, types.CodeType_Unauthorized, "Unable to create the file for writing. Check your write access privilege");
		return
	} else if _, err = io.Copy(out, fopen); err != nil {
		writeUploadResponse(&w, types.CodeType_InternalError, err.Error());
		return
	}

	// Generate Sample of data. Distribute it to other TM nodes
	sample := crypto.GenerateStorageSample(&fileBytes)
	signedSample, err := sample.SignSample(app.privKey)
	if err != nil {
		writeUploadResponse(&w, types.CodeType_BCFSInvalidSignature, "Could not sign storage sample.");
		return
	}

	// Broadcast the signed storage sample to all other tendermint consensus nodes.
	// TODO: Add some mechanic to resend the sample to the nodes which are unavailble.
	bytearr, err := json.Marshal(signedSample)
	responseChan := make(chan *QueryBroadcastReponse, len(app.TMRpcClients))
	done := make(chan int)
	app.broadcastQuery("/newsample", &bytearr, responseChan)
	goodResponses := make(map[string]bool)
	for {
		select {
		case v := <-responseChan:
			if v.Err == nil {
				goodResponses[v.Identity] = true
			}
		case <-done:
			return
		case <-time.After(time.Duration(conf.AppConfig().TmQueryTimeoutSeconds) * time.Second):
			return
		}
	}

	if _, ok := goodResponses[app.fingerprint]; !ok {
		// Could not store the sample in my own node? Some serious trouble.
		writeUploadResponse(&w, types.CodeType_OK, "Problems storing Storage sample.");
		return
	}

	// Need 2/3+ precommits to make progress.
	if len(goodResponses) >= ((2*len(conf.AppConfig().TendermintNodes))+1)/3 {
		writeUploadResponse(&w, types.CodeType_OK, "File temporary stored and storage sample distributed. " +
			"After uploading file to IPFS, send a transaction to the mempool.");
	} else {
		writeUploadResponse(&w, types.CodeType_Unauthorized, "Could not distribute storage samples to enough nodes" +
			" within the time period. Try again later.");
	}
}
