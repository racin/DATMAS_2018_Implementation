package app

import (
	"github.com/racin/DATMAS_2018_Implementation/crypto"
	"github.com/racin/DATMAS_2018_Implementation/types"
	abci "github.com/tendermint/abci/types"
	conf "github.com/racin/DATMAS_2018_Implementation/configuration"
	"github.com/racin/DATMAS_2018_Implementation/rpc"
)

func (app *Application) CheckTx_UploadData(signer *conf.Identity, tx *types.Transaction) *abci.ResponseCheckTx {
	// Check if uploader is allowed to upload data.
	if signer.Type != conf.Client {
		return &abci.ResponseCheckTx{Code: uint32(types.CodeType_Unauthorized), Log: "Only registered clients can upload data."}
	}

	stxReq, ok := tx.Data.(*crypto.SignedStruct)
	if !ok {
		return &abci.ResponseCheckTx{Code: uint32(types.CodeType_BCFSInvalidInput), Log: "Could not type assert Data."}
	}

	// Check contents of transaction.
	reqUpload, ok := stxReq.Base.(*types.RequestUpload)
	if !ok {
		return &abci.ResponseCheckTx{Code: uint32(types.CodeType_BCFSInvalidInput), Log: "Could not type assert StxReq."}
	}

	// Check if we have registered the public key of the IPFS node which holds the uploaded file in temporary storage.
	proverIdent, proverPubKey := app.GetIdentityPublicKey(reqUpload.IpfsNode)
	if proverIdent == nil|| proverPubKey == nil {
		return &abci.ResponseCheckTx{Code: uint32(types.CodeType_InternalError), Log: "Could not find the associated IPFS proxy " + reqUpload.IpfsNode}
	}

	if !stxReq.Verify(proverPubKey) {
		return &abci.ResponseCheckTx{Code: uint32(types.CodeType_BCFSInvalidSignature), Log: "Could not verify signature of upload request"}
	}

	// Issue a simple status check to see if the storage node still claims to have the file. Could use the timestamp
	// of the transaction instead.
	cidStx := app.GetSignedTransaction(types.TransactionType_IPFSProxyPin, reqUpload.Cid)
	ipfsResponse := rpc.QueryIPFSproxy(app.IpfsHttpClient, conf.AppConfig().IpfsProxyAddr, proverIdent.Address, conf.AppConfig().IpfsStatusEndpoint, cidStx)
	if ipfsResponse.Codetype != types.CodeType_OK {
		return &abci.ResponseCheckTx{Code: uint32(ipfsResponse.Codetype), Log: "Storage node does not claim to still hold the file. Addr: " +
			proverIdent.Address + ", Error: " + string(ipfsResponse.Message)}
	}

	return &abci.ResponseCheckTx{Code: uint32(types.CodeType_OK), Log: "All checks passed. CID: " + reqUpload.Cid}
}

func (app *Application) DeliverTx_UploadData(signer *conf.Identity, tx *types.Transaction) *abci.ResponseDeliverTx {
	// Check contents of transaction.
	stxReq, ok := tx.Data.(*crypto.SignedStruct)
	if !ok {
		return &abci.ResponseDeliverTx{Code: uint32(types.CodeType_BCFSInvalidInput), Log: "Could not type assert Data."}
	}

	// Check contents of transaction.
	reqUpload, ok := stxReq.Base.(*types.RequestUpload)
	if !ok {
		return &abci.ResponseDeliverTx{Code: uint32(types.CodeType_BCFSInvalidInput), Log: "Could not type assert StxReq."}
	}

	// Write Simple Metadata.
	types.WriteSimpleMetadata(conf.AppConfig().BasePath + conf.AppConfig().SimpleMetadata, reqUpload.Cid,
		&types.SimpleMetadataEntry{CID:reqUpload.Cid, FileSize:reqUpload.Length})

	// Update Prevailing block height
	app.prevailingBlock[reqUpload.Cid] = app.nextBlockHeight

	// All checks passed. Return OK.
	return &abci.ResponseDeliverTx{Code: uint32(types.CodeType_OK), Log: "File uploaded and recorded on the ledger. CID: " + reqUpload.Cid}
}