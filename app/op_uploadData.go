package app

import (
	"encoding/json"
	"github.com/racin/DATMAS_2018_Implementation/crypto"
	"github.com/racin/DATMAS_2018_Implementation/types"
	abci "github.com/tendermint/abci/types"
	conf "github.com/racin/DATMAS_2018_Implementation/configuration"
	"os"
	"fmt"
)

func (app *Application) CheckTx_UploadData(signer *conf.Identity, tx *types.Transaction) *abci.ResponseCheckTx {
	// Check if uploader is allowed to upload data.
	if signer.Type != 1 {
		return &abci.ResponseCheckTx{Code: uint32(types.CodeType_Unauthorized), Log: "Only registered clients can upload data."}
	}

	// Check contents of transaction.
	reqUpload, ok := tx.Data.(*types.RequestUpload)
	if !ok {
		return &abci.ResponseCheckTx{Code: uint32(types.CodeType_BCFSInvalidInput), Log: "Could not type assert Data."}
	}

	// Load storage sample for the file.
	storageSample := crypto.LoadStorageSample(conf.AppConfig().BasePath + conf.AppConfig().StorageSamples, reqUpload.Cid)
	if storageSample == nil {
		return &abci.ResponseCheckTx{Code: uint32(types.CodeType_InternalError), Log: "Could not find associated storage sample."}
	}

	// Generate a storage challenge and digitally sign it.
	storageChallenge, challengeHash := storageSample.GenerateChallenge(app.privKey)
	if storageChallenge == nil {
		return &abci.ResponseCheckTx{Code: uint32(types.CodeType_InternalError), Log: "Could not generate a StorageChallenge for sample."}
	}

	// Check if we have registered the public key of the IPFS node which holds the uploaded file in temporary storage.
	proverIdent, proverPubKey := app.GetIdentityPublicKey(reqUpload.IpfsNode)
	if proverIdent == nil|| proverPubKey == nil {
		return &abci.ResponseCheckTx{Code: uint32(types.CodeType_InternalError), Log: "Could not find the associated IPFS proxy " + reqUpload.IpfsNode}
	}

	// Issue a StorageChallenge to the IPFS node and check that the response is OK.
	ipfsResponse := app.queryIPFSproxy(proverIdent.Address, conf.AppConfig().IpfsChallengeEndpoint, storageChallenge)
	if ipfsResponse.Codetype != types.CodeType_OK {
		return &abci.ResponseCheckTx{Code: uint32(types.CodeType_InternalError), Log: "Did not get a proof from IPFS node " +
			proverIdent.Address + ", Error: " + string(ipfsResponse.Message)}
	}

	// Type assert the StorageProof.
	// signedProof := &crypto.SignedStruct{Base: &crypto.StorageChallengeProof{SignedStruct:crypto.SignedStruct{Base:&crypto.StorageChallenge{}}}}
	signedProof := &crypto.SignedStruct{Base: &crypto.StorageChallengeProof{SignedStruct:crypto.SignedStruct{Base:&crypto.StorageChallenge{}}}}
	fmt.Printf("IPFS Response: %+v\n", ipfsResponse.Message)
	if err := json.Unmarshal(ipfsResponse.Message, signedProof); err != nil {
		return &abci.ResponseCheckTx{Code: uint32(types.CodeType_BCFSInvalidInput), Log: err.Error()}
	}
	fmt.Printf("IPFS Response Signed: %+v\n", signedProof)
	challengeProof, ok := signedProof.Base.(*crypto.StorageChallengeProof)
	if !ok {
		return &abci.ResponseCheckTx{Code: uint32(types.CodeType_InternalError), Log: "Could not type assert StorageChallengeProof."}
	}
	fmt.Printf("IPFS Response Proof: %+v\n", challengeProof)
	signedProof.Base.(*crypto.StorageChallengeProof).Base, ok = challengeProof.Base.(*crypto.StorageChallenge)
	if !ok {
		return &abci.ResponseCheckTx{Code: uint32(types.CodeType_InternalError), Log: "Could not type assert StorageChallengeProof."}
	}
	fmt.Printf("IPFS Response Challenge: %+v\n", signedProof.Base.(*crypto.StorageChallengeProof).Base)

	// Verify the signatures of the StorageChallenge.
	if err := signedProof.VerifyChallengeProof(conf.AppConfig().BasePath + conf.AppConfig().StorageSamples, app.identity, app.privKey,
		proverIdent, proverPubKey, challengeHash); err != nil {
		return &abci.ResponseCheckTx{Code: uint32(types.CodeType_BCFSInvalidSignature), Log: err.Error()}
	}

	// All checks passed. Return OK.
	return &abci.ResponseCheckTx{Code: uint32(types.CodeType_OK), Log: "All checks passed."}
}

func (app *Application) DeliverTx_UploadData(signer *conf.Identity, tx *types.Transaction) *abci.ResponseDeliverTx {
	// Check contents of transaction.
	reqUpload, ok := tx.Data.(types.RequestUpload)
	if !ok {
		return &abci.ResponseDeliverTx{Code: uint32(types.CodeType_BCFSInvalidInput), Log: "Could not type assert Data."}
	}

	// Check if we have registered the public key of the IPFS node which holds the uploaded file in temporary storage.
	proverIdent, proverPubKey := app.GetIdentityPublicKey(reqUpload.IpfsNode)
	if proverIdent == nil|| proverPubKey == nil {
		return &abci.ResponseDeliverTx{Code: uint32(types.CodeType_InternalError), Log: "Could not find the associated IPFS proxy " + reqUpload.IpfsNode}
	}

	// TODO: Check how IPFS handles many Pin requests. If this is a problem we need make sure just one does it.
	ipfsResponse := app.queryIPFSproxy(proverIdent.Address, conf.AppConfig().IpfsPinfileEndpoint, reqUpload.Cid)
	if ipfsResponse.Codetype != types.CodeType_OK {
		// Couldnt pin the file.. Not good. Attempt send the same request to a different proxy.
		if proxyAddr, err := app.getIPFSProxyAddr(); err == nil {
			ipfsResponse = app.queryIPFSproxy(proxyAddr, conf.AppConfig().IpfsPinfileEndpoint, reqUpload.Cid)
		}
		if ipfsResponse.Codetype != types.CodeType_OK {
			return &abci.ResponseDeliverTx{Code: uint32(types.CodeType_InternalError), Log: "Unable to pin the file. Error: " + string(ipfsResponse.Message)}
		}
	}

	// Remove temporary stored file if its stored.
	filePath := conf.AppConfig().BasePath + conf.AppConfig().StorageSamples + reqUpload.Cid
	if _, err := os.Lstat(filePath); err == nil {
		os.Remove(conf.AppConfig().BasePath + conf.AppConfig().StorageSamples + reqUpload.Cid)
	}

	// All checks passed. Return OK.
	return &abci.ResponseDeliverTx{Code: uint32(types.CodeType_OK), Log: "File uploaded and recorded on the ledger. CID: " + string(ipfsResponse.Message)}
}