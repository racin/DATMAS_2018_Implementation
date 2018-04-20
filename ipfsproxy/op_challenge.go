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

func (proxy *Proxy) GetProof(cidStr string) *app.StorageChallengeProof{
	proof := &app.StorageChallengeProof{Proof:[]byte("abc")}
	err := proxy.client.IPFS().Get(cidStr, conf.IPFSProxyConfig().TempUploadPath)
	if err != nil {
		return nil
	}


	// Implement PoS here. Now simply check if hash is correct.
	if ipfsHash, err := crypto.IPFSHashFile(conf.IPFSProxyConfig().TempUploadPath + cidStr); err != nil {
		return nil
	} else if ipfsHash != cidStr {
		return nil
	}

	return proof
}
func (proxy *Proxy) Challenge(w http.ResponseWriter, r *http.Request) {
	// Only Consensus access level can execute this method.
	//

}