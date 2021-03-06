package app

import (
	"encoding/json"
	"github.com/racin/DATMAS_2018_Implementation/crypto"
	"github.com/racin/DATMAS_2018_Implementation/types"
	abci "github.com/tendermint/abci/types"
	conf "github.com/racin/DATMAS_2018_Implementation/configuration"
	tmtypes "github.com/tendermint/tendermint/types"
	"github.com/racin/DATMAS_2018_Implementation/rpc"
	"sync"
)

func (app *Application) Query_Challenge(reqQuery abci.RequestQuery) *abci.ResponseQuery{
	if reqQuery.Data == nil {
		return &abci.ResponseQuery{Code: uint32(types.CodeType_BCFSInvalidInput), Log: "Missing data parameter."}
	}
	signedStruct := &crypto.SignedStruct{Base: &crypto.StorageChallenge{}}
	if err := json.Unmarshal(reqQuery.Data, signedStruct); err != nil {
		return &abci.ResponseQuery{Code: uint32(types.CodeType_BCFSInvalidInput), Log: "Could not unmarshal SignedStruct. Error: " + err.Error()}
	}

	storageChallenge, ok := signedStruct.Base.(*crypto.StorageChallenge)
	if !ok {
		return &abci.ResponseQuery{Code: uint32(types.CodeType_BCFSInvalidInput), Log: "Could not unmarshal StorageChallenge."}
	}

	// Verify the signature and identity of the challenge.
	challengerIdent, challengerPubKey := app.GetIdentityPublicKey(storageChallenge.Identity)
	if !signedStruct.Verify(challengerPubKey) {
		return &abci.ResponseQuery{Code: uint32(types.CodeType_Unauthorized), Log: "Could not verify the signature of the challenge."}
	}

	// Challenge must have been issued by a client.
	if challengerIdent.Type != conf.Client {
		return &abci.ResponseQuery{Code: uint32(types.CodeType_Unauthorized), Log: "Identity unauthorized"}
	}

	// Load Simple metadata
	simpleMetaData := types.GetSimpleMetadata(conf.AppConfig().BasePath + conf.AppConfig().SimpleMetadata, storageChallenge.Cid)
	if simpleMetaData == nil || simpleMetaData.FileSize == 0 {
		return &abci.ResponseQuery{Code: uint32(types.CodeType_Unauthorized), Log: "Could not find a stored SimpleMetaData for this CID. Fatal error." +
			" Re-upload file to resolve this."}
	}

	// Generate a random challenge which we don't know the answer to.
	signRndChal, _ := crypto.GenerateRandomChallenge(app.privKey, storageChallenge.Cid, simpleMetaData.FileSize)
	if signRndChal == nil {
		return &abci.ResponseQuery{Code: uint32(types.CodeType_Unauthorized), Log: "Could not generate random challenge."}
	}

	proofs := make([]crypto.SignedStruct, 0)
	var wg sync.WaitGroup
	wg.Add(len(conf.AppConfig().IpfsNodes) * 2)

	// Issue the challenge from the Client first
	for _, ident := range conf.AppConfig().IpfsNodes {
		go func(id string) {
			addr := app.GetAccessList().GetAddress(id)
			ipfsResp := rpc.QueryIPFSproxy(app.IpfsHttpClient, conf.AppConfig().IpfsProxyAddr, addr, conf.AppConfig().IpfsChallengeEndpoint, signedStruct)

			if (ipfsResp.Codetype != types.CodeType_OK) {
				wg.Done()
				return // Not a valid proof. Do not care about why
			}
			scp := &crypto.SignedStruct{Base: &crypto.StorageChallengeProof{SignedStruct: crypto.SignedStruct{Base: &crypto.StorageChallenge{}}}}
			if err := json.Unmarshal(ipfsResp.Message, scp); err == nil {
				proofs = append(proofs, *scp)
			}
			wg.Done()
		}(ident)
	}

	// Then the randomly generated ones
	for _, ident := range conf.AppConfig().IpfsNodes {
		go func(id string) {
			addr := app.GetAccessList().GetAddress(id)
			ipfsResp := rpc.QueryIPFSproxy(app.IpfsHttpClient, conf.AppConfig().IpfsProxyAddr, addr, conf.AppConfig().IpfsChallengeEndpoint, signRndChal)

			if (ipfsResp.Codetype != types.CodeType_OK) {
				wg.Done()
				return // Not a valid proof. Do not care about why
			}
			scp := &crypto.SignedStruct{Base: &crypto.StorageChallengeProof{SignedStruct: crypto.SignedStruct{Base: &crypto.StorageChallenge{}}}}
			if err := json.Unmarshal(ipfsResp.Message, scp); err == nil {
				proofs = append(proofs, *scp)
			}
			wg.Done()
		}(ident)
	}

	// Wait for responses.
	wg.Wait()

	// Now lets send the proofs to the mempool
	stx := app.GetSignedTransaction(types.TransactionType_VerifyStorage, proofs)
	stxByteArr, err := json.Marshal(stx)
	if err != nil {
		return &abci.ResponseQuery{Code: uint32(types.CodeType_InternalError), Log: "Error marshalling: Error: " + err.Error()}
	}

	if codetype, err := types.CheckBroadcastResult(app.TMRpcClients[app.fingerprint].BroadcastTxSync(tmtypes.Tx(stxByteArr))); err != nil {
		return &abci.ResponseQuery{Code: uint32(codetype), Log: "Error broadcasting challenge. Error: " + err.Error()}
	} else{
		return &abci.ResponseQuery{Code: uint32(codetype), Log: "Transaction with proofs sent to mempool. Wait for commit."}
	}
}