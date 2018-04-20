package ipfsproxy

import (
	"io/ioutil"
	"net/http"
	"github.com/racin/DATMAS_2018_Implementation/types"
	conf "github.com/racin/DATMAS_2018_Implementation/configuration"
	"fmt"
)

func (proxy *Proxy) StatusAll(w http.ResponseWriter, r *http.Request) {
	txString, err := ioutil.ReadAll(r.Body)
	if err != nil {
		writeResponse(&w, types.CodeType_BCFSInvalidInput, "Missing transaction parameter.");
		return
	}

	if _, codeType, message := proxy.CheckProxyAccess(string(txString), conf.Client, conf.Consensus); codeType != types.CodeType_OK {
		writeResponse(&w, codeType, message);
	} else if pininfo, err := proxy.client.StatusAll(false); err != nil {
		writeResponse(&w, types.CodeType_InternalError, err.Error());
	} else {
		writeResponse(&w, types.CodeType_OK, fmt.Sprintf("%+v", pininfo));

		// Add transaction to list of known transactions (message contains hash of tranc)
		proxy.seenTranc[message] = true
	}
}
