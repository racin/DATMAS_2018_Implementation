package ipfsproxy

import (
	"io/ioutil"
	"net/http"
	"github.com/racin/DATMAS_2018_Implementation/types"
	conf "github.com/racin/DATMAS_2018_Implementation/configuration"
	"encoding/json"
	"github.com/racin/DATMAS_2018_Implementation/crypto"
	"strconv"
)

func (proxy *Proxy) GetFile(w http.ResponseWriter, r *http.Request) {
	txString, err := ioutil.ReadAll(r.Body)
	if err != nil {
		writeResponse(&w, types.CodeType_BCFSInvalidInput, "Missing transaction parameter.");
		return
	}

	// Check access to proxy method
	requestTx, codeType, message := proxy.CheckProxyAccess(string(txString), conf.Client)
	if codeType != types.CodeType_OK {
		writeResponse(&w, codeType, message);
		return
	}

	// Check if CID is contained within the transaction.
	cidStr, ok := requestTx.Data.(string)
	if (!ok) {
		writeResponse(&w, types.CodeType_BCFSInvalidInput, "Missing data hash parameter.");
		return
	}

	res, err := proxy.TMClient.ABCIQuery("/prevailingheight", []byte(cidStr))
	if err != nil {
		writeResponse(&w, types.CodeType_BCFSInvalidInput, "Error querying height. Error: " + err.Error());
		return
	}

	responseStx, responseTx, err := types.UnmarshalTransaction([]byte(res.Response.Log))
	if err != nil {
		writeResponse(&w, types.CodeType_InternalError, err.Error())
		return
	}

	prevHeightFl, ok :=responseTx.Data.(float64)
	if !ok {
		writeResponse(&w, types.CodeType_BCFSInvalidInput, "Error querying height. Could not type assert float64");
		return
	}
	prevHeight := int64(prevHeightFl)

	signer, pubKey := proxy.GetIdentityPublicKey(responseTx.Identity)
	if signer == nil {
		writeResponse(&w, types.CodeType_Unauthorized, "Could not get access list")
		return
	}

	if pubKey == nil {
		writeResponse(&w, types.CodeType_BCFSInvalidSignature, "Could not locate public key")
		return
	} else if !responseStx.Verify(pubKey) {
		writeResponse(&w, types.CodeType_BCFSInvalidSignature, "Could not verify signature")
		return
	}

	result, err := proxy.TMClient.Block(&prevHeight)
	if err != nil {
		writeResponse(&w, types.CodeType_InternalError, "Error getting block. Error: " + err.Error())
		return
	}
	if err := result.Block.ValidateBasic(); err != nil {
		writeResponse(&w, types.CodeType_InternalError, "Could not validate block. Error: " + err.Error())
		return
	}
	var access bool = false
	for i := int64(0); i < result.Block.NumTxs; i++ {
		if _, blockTx, err := types.UnmarshalTransaction([]byte(result.Block.Txs[i])); err == nil {
			switch blockTx.Type {
			case types.TransactionType_RemoveData:
				reqRemoval, ok := blockTx.Data.(string);
				if ok && reqRemoval == cidStr {
					writeResponse(&w, types.CodeType_BCFSInvalidBlockHeight, "File was removed at block height: " + strconv.Itoa(int(prevHeight)))
					return
				}
			case types.TransactionType_UploadData:
				if signedStruct, ok :=  blockTx.Data.(*crypto.SignedStruct); ok {
					reqUpload, ok := signedStruct.Base.(*types.RequestUpload);
					if !ok || reqUpload.Cid != cidStr {
						continue
					}
					if blockTx.Identity == requestTx.Identity{
						access = true
					}
					break
				}
			case types.TransactionType_ChangeContentAccess:
				changeAccess, ok := blockTx.Data.(*types.ChangeAccess)
				if !ok || changeAccess.Cid != cidStr {
					continue
				}
				if blockTx.Identity == requestTx.Identity{
					access = true
					break
				}
				for _, reader := range changeAccess.Readers {
					if requestTx.Identity == reader {
						access = true
					}
				}
				break
			}

		}
	}

	if !access {
		writeResponse(&w, types.CodeType_BCFSNoAccess, "Missing access to download requested file.");
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