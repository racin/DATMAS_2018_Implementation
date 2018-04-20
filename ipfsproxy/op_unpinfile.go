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

func (proxy *Proxy) UnPinFile(w http.ResponseWriter, r *http.Request) {
	// For removing stored data.
}
