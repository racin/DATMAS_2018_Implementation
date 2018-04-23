package ipfsproxy

import (
	cid2 "gx/ipfs/QmcZfnkapfECQGcLZaf9B79NRg7cRa9EnZh4LSbkCzwNvY/go-cid"
	mh "gx/ipfs/QmZyZDi491cCNTLfAhwcaDii2Kg4pwKRkhqQzURGDvY6ua/go-multihash"
	"io/ioutil"
	"net/http"
	"github.com/racin/DATMAS_2018_Implementation/types"
	conf "github.com/racin/DATMAS_2018_Implementation/configuration"
)

// For removing stored data. Should only be called by a consensus node after the ledger is updated to reflect the removal.
func (proxy *Proxy) UnPinFile(w http.ResponseWriter, r *http.Request) {
	txString, err := ioutil.ReadAll(r.Body)
	if err != nil {
		writeResponse(&w, types.CodeType_BCFSInvalidInput, "Missing transaction parameter.");
		return
	}

	// Check access to proxy method
	tx, codeType, message := proxy.CheckProxyAccess(string(txString), conf.Consensus)
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

	b58, err := mh.FromB58String(cidStr)
	if err != nil {
		writeResponse(&w, types.CodeType_InternalError, err.Error());
		return
	}

	cid := cid2.NewCidV0(b58)
	if err := proxy.client.Unpin(cid); err != nil {
		writeResponse(&w, types.CodeType_InternalError, err.Error());
	} else {
		writeResponse(&w, types.CodeType_OK, "File with CID: " + cidStr + " was unpinned.");
		// Add transaction to list of known transactions (message contains hash of tranc)
		proxy.seenTranc[message] = true
	}
}
