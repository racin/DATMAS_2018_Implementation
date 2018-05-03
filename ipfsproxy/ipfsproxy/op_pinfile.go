package ipfsproxy

import (
	cid2 "gx/ipfs/QmcZfnkapfECQGcLZaf9B79NRg7cRa9EnZh4LSbkCzwNvY/go-cid"
	mh "gx/ipfs/QmZyZDi491cCNTLfAhwcaDii2Kg4pwKRkhqQzURGDvY6ua/go-multihash"
	"io/ioutil"
	"net/http"
	"github.com/racin/DATMAS_2018_Implementation/types"
	conf "github.com/racin/DATMAS_2018_Implementation/configuration"
	"fmt"
)

func (proxy *Proxy) PinFile(w http.ResponseWriter, r *http.Request) {
	return // Disabled.

	fmt.Println("IPFS PIN FILE")
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

	if !proxy.isCidInLedger(cidStr) {
		writeResponse(&w, types.CodeType_BCFSUnknownAddress, "Could not find CID in the ledger.");
		return
	}

	pinCodeType, pinMessage := proxy.pinFile(cidStr)
	// Add transaction to list of known transactions (message contains hash of tranc)
	if pinCodeType == types.CodeType_OK {
		proxy.seenTranc[message] = true
	}
	writeResponse(&w, pinCodeType, pinMessage)
}

func (proxy *Proxy) pinFile(cidStr string) (types.CodeType, string){
	// Check if we have stored this file locally.
	// TODO: Test if this is necessary. Might be enough that it has been added to any node in the cluster.
	err := proxy.client.IPFS().Get(cidStr, conf.IPFSProxyConfig().TempUploadPath)
	if err != nil {
		return types.CodeType_BCFSUnknownAddress, "Could not find file with hash. Error: " + err.Error()
	}

	// Pin file.
	b58, err := mh.FromB58String(cidStr)
	if err != nil {
		return types.CodeType_InternalError, err.Error()
	}

	if err := proxy.client.Pin(cid2.NewCidV0(b58), -1, -1, ""); err != nil {
		return types.CodeType_InternalError, err.Error()
	} else {
		return types.CodeType_OK, "File pinned."
	}
}

func (proxy *Proxy) isCidInLedger(cid string) bool {
	// TODO: Implement
	return true
}