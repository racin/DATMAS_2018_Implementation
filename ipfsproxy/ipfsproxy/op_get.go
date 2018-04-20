package ipfsproxy

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"github.com/racin/DATMAS_2018_Implementation/types"
	conf "github.com/racin/DATMAS_2018_Implementation/configuration"
)

func (proxy *Proxy) GetFile(w http.ResponseWriter, r *http.Request) {
	// TODO: Options; 1. Send the file directly. 2. Send URL. 3. Should check whats recorded on the ledger, and if the identity has access.
	txString, err := ioutil.ReadAll(r.Body)
	if err != nil {
		writeResponse(&w, types.CodeType_BCFSInvalidInput, "Missing transaction parameter.");
		return
	}

	// Check access to proxy method
	tx, codeType, message := proxy.CheckProxyAccess(string(txString), conf.Client, conf.Consensus)
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

	if err := proxy.client.IPFS().Get(cidStr, conf.IPFSProxyConfig().TempUploadPath); err != nil {
		writeResponse(&w, types.CodeType_BCFSUnknownAddress, "Could not find file with hash. Error: " + err.Error());
		return
	}

	fileBytes, err := ioutil.ReadFile(conf.IPFSProxyConfig().TempUploadPath + cidStr)
	if err != nil {
		writeResponse(&w, types.CodeType_InternalError, "Could not read file. Error: " + err.Error());
		return
	}
	// Add transaction to list of known transactions (message contains hash of tranc)
	proxy.seenTranc[message] = true

	json.NewEncoder(w).Encode(&types.IPFSReponse{Message:fileBytes, Codetype:types.CodeType_OK})
}