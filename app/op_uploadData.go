package app

import (
	"encoding/json"
	"github.com/racin/DATMAS_2018_Implementation/crypto"
	"github.com/racin/DATMAS_2018_Implementation/types"
	abci "github.com/tendermint/abci/types"
	conf "github.com/racin/DATMAS_2018_Implementation/configuration"
)

func (app *Application) UploadData_CheckTx(signer *conf.Identity, tx *types.Transaction) *abci.ResponseCheckTx {
	// Check if uploader is allowed to upload data.
	if signer.Type != 1 {
		return &abci.ResponseCheckTx{Code: uint32(types.CodeType_Unauthorized), Log: "Only registered clients can upload data."}
	}

	// Check if data hash is contained within the transaction.
	reqUpload, ok := tx.Data.(types.RequestUpload)
	if !ok {
		return &abci.ResponseCheckTx{Code: uint32(types.CodeType_BCFSInvalidInput), Log: "Could not type assert Data to string"}
	}

	// Load storage sample for the file.
	storageSample := crypto.LoadStorageSample(conf.AppConfig().StorageSamples, reqUpload.Cid)
	if storageSample == nil {
		return &abci.ResponseCheckTx{Code: uint32(types.CodeType_InternalError), Log: "Could not find associated storage sample."}
	}

	storageChallenge := storageSample.GenerateChallenge(app.privKey)
	if storageChallenge == nil {
		return &abci.ResponseCheckTx{Code: uint32(types.CodeType_InternalError), Log: "Could not generate a StorageChallenge for sample."}
	}


	proverIdent, proverPubKey := app.GetIdentityPublicKey(reqUpload.IpfsNode)
	if proverIdent == nil|| proverPubKey == nil {
		return &abci.ResponseCheckTx{Code: uint32(types.CodeType_InternalError), Log: "Could not find the associated IPFS proxy " + reqUpload.IpfsNode}
	}
	// Check if a file with this hash exists on an IPFS node and is uploaded to our server.
	ipfsResponse := app.queryIPFSproxy(proverIdent.Address, conf.AppConfig().IpfsChallengeEndpoint, storageChallenge)
	if ipfsResponse.Codetype != types.CodeType_OK {
		return &abci.ResponseCheckTx{Code: uint32(types.CodeType_InternalError), Log: "Did not get a proof from IPFS node " +
			proverIdent.Address + ", Error: " + string(ipfsResponse.Message)}
	}

	signedProof := &crypto.SignedStruct{Base: &crypto.StorageChallengeProof{}}
	if err := json.Unmarshal(ipfsResponse.Message, signedProof); err != nil {
		return &abci.ResponseCheckTx{Code: uint32(types.CodeType_BCFSInvalidInput), Log: "Could not unmarshal StorageChallengeProof."}
	}

	challengeProof, ok := signedProof.Base.(*crypto.StorageChallengeProof)
	if !ok {
		return &abci.ResponseCheckTx{Code: uint32(types.CodeType_InternalError), Log: "Could not type assert StorageChallengeProof."}
	}

	// Hash of IPFS public key? WHere?
	if err := challengeProof.VerifyChallengeProof(conf.AppConfig().StorageSamples, app.identity, app.privKey,
		proverIdent, proverPubKey); err != nil {
		return &abci.ResponseCheckTx{Code: uint32(types.CodeType_BCFSInvalidSignature), Log: "Could not verify the StorageChallengeProof."}
	}

	// All checks passed. Return OK.
	return &abci.ResponseCheckTx{Code: uint32(types.CodeType_OK), Log: "All checks passed."}
}