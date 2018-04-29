package app

import (
	"encoding/json"
	"github.com/racin/DATMAS_2018_Implementation/crypto"
	"github.com/racin/DATMAS_2018_Implementation/types"
	abci "github.com/tendermint/abci/types"
	conf "github.com/racin/DATMAS_2018_Implementation/configuration"
	"os"
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
	// Check if this sample is already stored. Should use a different path if we want to remove it (future work...)
	// Return OK if the actual sample equals the current stored one.
	if _, err := os.Lstat(conf.AppConfig().StorageSamples + storageChallenge.Cid); err == nil {
		currStoredSample := crypto.LoadStorageSample(conf.AppConfig().StorageSamples, storageSample.Cid)
		if storageSample.CompareTo(currStoredSample) {
			return &abci.ResponseQuery{Code: uint32(types.CodeType_OK), Log: "The same sample was already stored."}
		} else {
			return &abci.ResponseQuery{Code: uint32(types.CodeType_InternalError), Log: "A different sample for this file is already stored."}
		}
	}

	// Store the sample.
	if err := signedStruct.StoreSample(conf.AppConfig().BasePath + conf.AppConfig().StorageSamples); err != nil {
		return &abci.ResponseQuery{Code: uint32(types.CodeType_InternalError), Log: "Could not store sample. Error: " + err.Error()}
	}

	return &abci.ResponseQuery{Code: uint32(types.CodeType_OK), Log: "Sample stored."}
}
