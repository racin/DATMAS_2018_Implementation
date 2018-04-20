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

func (proxy *Proxy) PinFile(w http.ResponseWriter, r *http.Request) {
	txString, err := ioutil.ReadAll(r.Body)
	if err != nil {
		writeResponse(&w, types.CodeType_BCFSInvalidInput, "Missing transaction parameter.");
		return
	}

	// Check access to proxy method
	tx, codeType, message := proxy.CheckProxyAccess(string(txString), app.User)
	if codeType != types.CodeType_OK {
		writeResponse(&w, codeType, message);
		return
	}

	// Check if CID is contained within the transaction.
	cidStr, ok := tx.Data.(string)
	if (!ok) {
		writeResponse(&w, types.CodeType_BCFSInvalidInput, "Missing data hash parameter.");
		return
	}

	err = proxy.client.IPFS().Get(cidStr, conf.IPFSProxyConfig().TempUploadPath)
	if err != nil {
		writeResponse(&w, types.CodeType_BCFSUnknownAddress, "Could not find file with hash. Error: " + err.Error());
		return
	}

	// Pin file.
	b58, err := mh.FromB58String(cidStr)
	if err != nil {
		writeResponse(&w, types.CodeType_InternalError, err.Error());
		return
	}

	if err := proxy.client.Pin(cid2.NewCidV0(b58), -1, -1, ""); err != nil {
		writeResponse(&w, types.CodeType_InternalError, err.Error());
	} else {
		writeResponse(&w, types.CodeType_OK, "File pinned.");
	}
}
