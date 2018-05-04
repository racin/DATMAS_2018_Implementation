package ipfsproxy

import (
	cid2 "gx/ipfs/QmcZfnkapfECQGcLZaf9B79NRg7cRa9EnZh4LSbkCzwNvY/go-cid"
	mh "gx/ipfs/QmZyZDi491cCNTLfAhwcaDii2Kg4pwKRkhqQzURGDvY6ua/go-multihash"
	"github.com/racin/DATMAS_2018_Implementation/types"
)

func (proxy *Proxy) pinFile(cidStr string) (types.CodeType, string) {
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