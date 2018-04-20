package ipfsproxy

import (
	cid2 "github.com/ipfs/go-cid"
	mh "github.com/multiformats/go-multihash"
	ma "github.com/multiformats/go-multiaddr"
	"os"
	"github.com/racin/DATMAS_2018_Implementation/crypto"
	"io/ioutil"
	"bytes"
	"net/http"
	"github.com/racin/DATMAS_2018_Implementation/types"
)

func (proxy *Proxy) AddFileNoPin(w http.ResponseWriter, r *http.Request) {
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

	// Check access to proxy method
	tx, codeType, message := proxy.CheckProxyAccess(txString[0], app.User)
	if codeType != types.CodeType_OK {
		writeResponse(&w, codeType, message);
		return
	}

	// Check if data hash is contained within the transaction.
	fileHash, ok := tx.Data.(string)
	if (fileHash == "" || !ok) {
		writeResponse(&w, types.CodeType_BCFSInvalidInput, "Missing data hash parameter.");
		return
	}

	files, ok := formdata.File["file"]
	if !ok || len(files) > 1 {
		writeResponse(&w, types.CodeType_BCFSInvalidInput, "File parameter should contain exactly one file.");
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
		writeResponse(&w, types.CodeType_BCFSInvalidInput, "Could not get byte array of input file.");
		return
	} else if uplFileHash, err := crypto.IPFSHashData(fileBytes); err != nil {
		writeResponse(&w, types.CodeType_BCFSInvalidInput, "Could not get hash of input file.");
		return
	} else if uplFileHash != fileHash {
		writeResponse(&w, types.CodeType_BCFSInvalidInput, "Data hash parameter does not equal hash of uploaded file.");
		return
	}

	if resStr, err := proxy.client.IPFS().AddNoPin(bytes.NewReader(fileBytes)); err != nil {
		writeResponse(&w, types.CodeType_BCFSInvalidInput, resStr + ". Error: " + err.Error());
	} else {
		writeResponse(&w, types.CodeType_OK, resStr);
		// Add transaction to list of known transactions (message contains hash of tranc)
		proxy.seenTranc[message] = true
	}

}
