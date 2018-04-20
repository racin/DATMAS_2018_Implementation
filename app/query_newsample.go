package app

import (
	"encoding/json"
	"github.com/racin/DATMAS_2018_Implementation/crypto"
	"github.com/racin/DATMAS_2018_Implementation/types"
	abci "github.com/tendermint/abci/types"
	conf "github.com/racin/DATMAS_2018_Implementation/configuration"
	"os"
)

func (app *Application) Query_Newsample(reqQuery abci.RequestQuery) *abci.ResponseQuery{
	if reqQuery.Data == nil {
		return &abci.ResponseQuery{Code: uint32(types.CodeType_BCFSInvalidInput), Log: "Missing data parameter."}
	}
	signedStruct := &crypto.SignedStruct{Base: &crypto.StorageSample{}}
	if err := json.Unmarshal(reqQuery.Data, signedStruct); err != nil {
		return &abci.ResponseQuery{Code: uint32(types.CodeType_BCFSInvalidInput), Log: "Could not unmarshal SignedStruct. Error: " + err.Error()}
	}

	storageSample, ok := signedStruct.Base.(*crypto.StorageSample)
	if !ok {
		return &abci.ResponseQuery{Code: uint32(types.CodeType_BCFSInvalidInput), Log: "Could not unmarshal StorageSample."}
	}

	// Verify the signature and identity of the sample.
	samplerIdent, samplerPubKey := app.GetIdentityPublicKey(storageSample.Identity)
	if err := signedStruct.VerifySample(samplerIdent, samplerPubKey); err != nil {
		return &abci.ResponseQuery{Code: uint32(types.CodeType_Unauthorized), Log: "Could not verify sample. Error: " + err.Error()}
	}

	// Sample must have been generated by a consensus node.
	if samplerIdent.Type != conf.Consensus {
		return &abci.ResponseQuery{Code: uint32(types.CodeType_Unauthorized), Log: "Identity unauthorized"}
	}

	// Check if this sample is already stored. Should use a different path if we want to remove it (future work...)
	// Return OK if the actual sample equals the current stored one.
	if _, err := os.Lstat(conf.AppConfig().StorageSamples + storageSample.Cid); err == nil {
		currStoredSample := crypto.LoadStorageSample(conf.AppConfig().StorageSamples, storageSample.Cid)
		if storageSample.CompareTo(currStoredSample) {
			return &abci.ResponseQuery{Code: uint32(types.CodeType_OK), Log: "The same sample was already stored."}
		} else {
			return &abci.ResponseQuery{Code: uint32(types.CodeType_InternalError), Log: "A different sample for this file is already stored."}
		}
	}

	// Store the sample.
	if err := signedStruct.StoreSample(conf.AppConfig().StorageSamples); err != nil {
		return &abci.ResponseQuery{Code: uint32(types.CodeType_InternalError), Log: "Could not store sample. Error: " + err.Error()}
	}

	return &abci.ResponseQuery{Code: uint32(types.CodeType_OK), Log: "Sample stored."}
}