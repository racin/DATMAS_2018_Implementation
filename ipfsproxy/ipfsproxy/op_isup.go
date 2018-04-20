package ipfsproxy

import (
	"net/http"
	"github.com/racin/DATMAS_2018_Implementation/types"
)
/**
Simply checks if the IPFS service is up. Does not need to protected.
 */
func (proxy *Proxy) IsUp(w http.ResponseWriter, r *http.Request) {
	if proxy.client.IPFS().IsUp() {
		writeResponse(&w, types.CodeType_OK, "true");
	} else {
		writeResponse(&w, types.CodeType_OK, "false");
	}
}
