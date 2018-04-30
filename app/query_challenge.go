package app

import (
	"encoding/json"
	"github.com/racin/DATMAS_2018_Implementation/crypto"
	"github.com/racin/DATMAS_2018_Implementation/types"
	abci "github.com/tendermint/abci/types"
	conf "github.com/racin/DATMAS_2018_Implementation/configuration"
	tmtypes "github.com/tendermint/tendermint/types"
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
	if err := signedStruct.VerifySample(challengerIdent, challengerPubKey); err != nil {
		return &abci.ResponseQuery{Code: uint32(types.CodeType_Unauthorized), Log: "Could not verify sample. Error: " + err.Error()}
	}

	// Challenge must have been issued by a client.
	if challengerIdent.Type != conf.Client {
		return &abci.ResponseQuery{Code: uint32(types.CodeType_Unauthorized), Log: "Identity unauthorized"}
	}

	// Load Simple metadata
	simpleMetaData := types.GetSimpleMetadata(conf.AppConfig().BasePath + conf.AppConfig().SimpleMetadata, storageChallenge.Cid)
	if simpleMetaData == nil {
		return &abci.ResponseQuery{Code: uint32(types.CodeType_Unauthorized), Log: "Could not find a stored SimpleMetaData for this CID. Fatal error." +
			" Re-upload file to resolve this."}
	}

	// Generate a random challenge which we don't know the answer to.
	rndChal, err := crypto.GenerateRandomChallenge(simpleMetaData.FileSize)
	if err != nil {
		return &abci.ResponseQuery{Code: uint32(types.CodeType_Unauthorized), Log: "Could not generate random challenge."}
	}
	signRndChal, err := crypto.SignStruct(rndChal, app.privKey)

	lenStorNodes := len(conf.AppConfig().IpfsNodes)
	proofs := make([]crypto.SignedStruct, lenStorNodes*2)

	// Issue the challenge from the Client first
	for i, ipfsNode := range conf.AppConfig().IpfsNodes {
		ipfsResp := app.queryIPFSproxy(ipfsNode, conf.AppConfig().IpfsChallengeEndpoint, storageChallenge)
		if (ipfsResp.Codetype != types.CodeType_OK) {
			continue // Not a valid proof. Do not care about why
		}
		scp := &crypto.SignedStruct{Base:&crypto.StorageChallengeProof{SignedStruct: crypto.SignedStruct{Base:&crypto.StorageChallenge{}}}}
		if err := json.Unmarshal(ipfsResp.Message, scp); err == nil {
			proofs[i] = *scp
		}
	}

	// Then the randomly generated ones
	for i, ipfsNode := range conf.AppConfig().IpfsNodes {
		ipfsResp := app.queryIPFSproxy(ipfsNode, conf.AppConfig().IpfsChallengeEndpoint, signRndChal)
		if (ipfsResp.Codetype != types.CodeType_OK) {
			continue // Not a valid proof. Do not care about why
		}
		scp := &crypto.SignedStruct{Base:&crypto.StorageChallengeProof{SignedStruct: crypto.SignedStruct{Base:&crypto.StorageChallenge{}}}}
		if err := json.Unmarshal(ipfsResp.Message, scp); err == nil {
			proofs[i + lenStorNodes] = *scp
		}
	}

	// Now lets send the proofs to the mempool
	tx := types.NewTx(proofs, app.fingerprint, types.TransactionType_VerifyStorage)
	stx,err := crypto.SignStruct(tx, app.privKey)
	if err != nil {
		return &abci.ResponseQuery{Code: uint32(types.CodeType_BCFSInvalidSignature), Log: "Could not sign StorageChallengeProofs"}
	}

	stxByteArr, _ := json.Marshal(stx)
	app.TMRpcClients[app.fingerprint].BroadcastTxAsync(tmtypes.Tx(stxByteArr))
	return &abci.ResponseQuery{Code: uint32(types.CodeType_OK), Log: "Transaction with proofs sent to mempool. Wait for commit."}
}