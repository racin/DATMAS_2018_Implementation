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

func (proxy *Proxy) StatusAll(w http.ResponseWriter, r *http.Request) {
	txString, err := ioutil.ReadAll(r.Body)
	if err != nil {
		writeResponse(&w, types.CodeType_BCFSInvalidInput, "Missing transaction parameter.");
		return
	}

	if _, codeType, message := proxy.CheckProxyAccess(string(txString), app.User); codeType != types.CodeType_OK {
		writeResponse(&w, codeType, message);
	} else if pininfo, err := proxy.client.StatusAll(false); err != nil {
		writeResponse(&w, types.CodeType_InternalError, err.Error());
	} else {
		writeResponse(&w, types.CodeType_OK, fmt.Sprintf("%+v", pininfo));

		// Add transaction to list of known transactions (message contains hash of tranc)
		proxy.seenTranc[message] = true
	}
}
