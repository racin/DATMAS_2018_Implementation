package ipfsproxy

import (
	"github.com/racin/DATMAS_2018_Implementation/crypto"
	"io/ioutil"
	"bytes"
	"net/http"
	"github.com/racin/DATMAS_2018_Implementation/types"
	conf "github.com/racin/DATMAS_2018_Implementation/configuration"
	"fmt"
	"encoding/json"
	"strconv"
)

// Adds the file to a single IPFS node. Only a client should be able to do this. (Consensus node can distribute an
// already uploaded file by pinning it.)
func (proxy *Proxy) AddFileNoPin(w http.ResponseWriter, r *http.Request) {
	fmt.Println("IPFS ADDNOPIN")
	err := r.ParseMultipartForm(104857600) // Up to 100MB stored in memory.
	if err != nil {
		writeResponse(&w, types.CodeType_InternalError, err.Error());
		return
	}
	formdata := r.MultipartForm
	txString, ok := formdata.Value["transaction"]
	if !ok {
		writeResponse(&w, types.CodeType_BCFSInvalidInput, "Missing transaction parameter.");
		return
	}

	fmt.Printf("Got tranc: %+v\n", txString)
	// Check access to proxy method
	tx, codeType, message := proxy.CheckProxyAccess(txString[0], conf.Client)
	if codeType != types.CodeType_OK {
		writeResponse(&w, codeType, message);
		return
	}

	// Check if data hash is contained within the transaction.
	reqUpload, ok := tx.Data.(*types.RequestUpload)
	if (reqUpload.Cid == "" || !ok) {
		writeResponse(&w, types.CodeType_BCFSInvalidInput, "Missing data hash parameter.");
		return
	}

	// Check that exactly one file is sent
	files, ok := formdata.File["file"]
	if !ok || len(files) > 1 {
		writeResponse(&w, types.CodeType_BCFSInvalidInput, "File parameter should contain exactly one file.");
		return
	}

	file := files[0]
	fopen, err := file.Open()
	defer fopen.Close()
	if err != nil {
		writeResponse(&w, types.CodeType_InternalError, "Could not access attached file.");
		return
	}

	// Check if the hash of the upload file equals the hash contained in the transaction
	fileBytes, err := ioutil.ReadAll(fopen)
	fmt.Printf("Filebytes: %v\n", fileBytes)
	if err != nil {
		writeResponse(&w, types.CodeType_BCFSInvalidInput, "Could not get byte array of input file.");
		return
	} else if uplFileHash, err := crypto.IPFSHashData(fileBytes); err != nil {
		writeResponse(&w, types.CodeType_BCFSInvalidInput, "Could not get hash of input file.");
		return
	} else if uplFileHash != reqUpload.Cid {
		writeResponse(&w, types.CodeType_BCFSInvalidInput, "Data hash parameter does not equal hash of uploaded file.");
		return
	} else if file.Size != reqUpload.Length {
		writeResponse(&w, types.CodeType_BCFSInvalidInput, "File size did not match the uploaded file. LS: " +
			strconv.Itoa(int(file.Size)) + ", RS: " + strconv.Itoa(int(reqUpload.Length)));
		return
	}

	if resStr, err := proxy.client.IPFS().AddNoPin(bytes.NewReader(fileBytes)); err != nil {
		writeResponse(&w, types.CodeType_BCFSInvalidInput, resStr + ". Error: " + err.Error());
	} else {
		signedStruct, err := crypto.SignStruct(reqUpload, proxy.privKey)
		if err != nil {
			writeResponse(&w, types.CodeType_InternalError, "Could not sign response: " + err.Error());
			return
		}
		byteArr, err := json.Marshal(signedStruct)
		if err != nil {
			writeResponse(&w, types.CodeType_InternalError, "Could not sign response: " + err.Error());
			return
		}
		writeResponse(&w, types.CodeType_OK, string(byteArr));

		// Add transaction to list of known transactions (message contains hash of tranc)
		proxy.seenTranc[message] = true
	}

}
