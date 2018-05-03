package ipfsproxy

import (
	"encoding/json"
	"github.com/racin/DATMAS_2018_Implementation/crypto"
	"io/ioutil"
	"net/http"
	"github.com/racin/DATMAS_2018_Implementation/types"
	conf "github.com/racin/DATMAS_2018_Implementation/configuration"
	"fmt"
)

func (proxy *Proxy) Challenge(w http.ResponseWriter, r *http.Request) {
	fmt.Println("IPFS CHALLENGE")
	// Both Clients and Consensus can issue challenges.
	txString, err := ioutil.ReadAll(r.Body)
	if err != nil {
		writeResponse(&w, types.CodeType_BCFSInvalidInput, "Missing transaction parameter.");
		return
	}

	stx := &crypto.SignedStruct{Base: &crypto.StorageChallenge{}}
	var storageChallenge *crypto.StorageChallenge
	var ok bool = false
	if err := json.Unmarshal([]byte(txString), stx); err != nil {
		writeResponse(&w, types.CodeType_BCFSInvalidInput, "Could not Marshal transaction. Error: " + err.Error())
		return
	} else if storageChallenge, ok = stx.Base.(*crypto.StorageChallenge); !ok {
		writeResponse(&w, types.CodeType_BCFSInvalidInput, "Could not Marshal transaction.")
		return
	}

	// Check for replay attack
	txHash := crypto.HashStruct(storageChallenge)
	if proxy.HasSeenTranc(txHash) {
		writeResponse(&w, types.CodeType_BadNonce, "Could not process transaction. Possible replay attack.")
		return
	}


	// Get signers identity and public key
	signer, pubKey := proxy.GetIdentityPublicKey(storageChallenge.Identity)
	if signer == nil {
		writeResponse(&w, types.CodeType_Unauthorized, "Could not get access list")
		return
	}

	// Check if public key exists and if message is signed.
	if pubKey == nil {
		writeResponse(&w, types.CodeType_BCFSInvalidSignature, "Could not locate public key")
		return
	} else if err := stx.VerifyChallenge(signer, pubKey); err != nil {
		writeResponse(&w, types.CodeType_BCFSInvalidSignature, "Could not verify signature b")
		return
	}

	if err := proxy.client.IPFS().Get(storageChallenge.Cid, conf.IPFSProxyConfig().TempUploadPath); err != nil {
		writeResponse(&w, types.CodeType_BCFSUnknownAddress, "Could not find file with hash. Error: " + err.Error());
		return
	}

	fileBytes, err := ioutil.ReadFile(conf.IPFSProxyConfig().TempUploadPath + storageChallenge.Cid)
	if err != nil {
		writeResponse(&w, types.CodeType_InternalError, "Could not read file. Error: " + err.Error());
		return
	}

	if proof := stx.ProveChallenge(proxy.privKey, &fileBytes); proof == nil {
		// TODO: Fatal error.
		writeResponse(&w, types.CodeType_InternalError, "Unable to prove challenge.");
	} else {
		byteArr, _ := json.Marshal(proof)
		json.NewEncoder(w).Encode(&types.IPFSReponse{Message:byteArr, Codetype:0})
	}
}