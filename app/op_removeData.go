package app

import (
	"github.com/racin/DATMAS_2018_Implementation/types"
	abci "github.com/tendermint/abci/types"
	conf "github.com/racin/DATMAS_2018_Implementation/configuration"
	"github.com/racin/DATMAS_2018_Implementation/crypto"
	"strconv"
)

func (app *Application) CheckTx_RemoveData(signer *conf.Identity, requestTx *types.Transaction) *abci.ResponseCheckTx {
	cidStr, ok := requestTx.Data.(string)
	if !ok {
		return &abci.ResponseCheckTx{Code: uint32(types.CodeType_BCFSInvalidInput), Log: "Could not type assert Data."}
	}

	// To remove data, the identity of the client must be the same as the one that originally uploaded it.
	prevailingHeight, ok := app.prevailingBlock[cidStr]
	if !ok {
		return &abci.ResponseCheckTx{Code: uint32(types.CodeType_Unauthorized), Log: "File with requested CID is not found in the system."}
	}

	result, err  := app.TMRpcClients[app.fingerprint].Block(&prevailingHeight)
	if err != nil {
		return &abci.ResponseCheckTx{Code: uint32(types.CodeType_InternalError), Log: "Error getting block. Error: " + err.Error()}
	}

	if err := result.Block.ValidateBasic(); err != nil {
		return &abci.ResponseCheckTx{Code: uint32(types.CodeType_InternalError), Log: "Could not validate block. Error: " + err.Error()}
	}

	for i := int64(0); i < result.Block.NumTxs; i++ {
		if _, blockTx, err := types.UnmarshalTransaction([]byte(result.Block.Txs[i])); err == nil {
			switch blockTx.Type {
			case types.TransactionType_RemoveData:
				reqRemoval, ok := blockTx.Data.(string)
				if ok && reqRemoval == cidStr {
					return &abci.ResponseCheckTx{Code: uint32(types.CodeType_InternalError), Log: "File was already removed at block height: " + strconv.Itoa(int(prevailingHeight))}
				}
			case types.TransactionType_UploadData:
				if signedStruct, ok :=  blockTx.Data.(*crypto.SignedStruct); ok {
					reqUpload, ok := signedStruct.Base.(*types.RequestUpload);
					if ok && reqUpload.Cid == cidStr && blockTx.Identity == requestTx.Identity{
						return &abci.ResponseCheckTx{Code: uint32(types.CodeType_OK), Log: "All checks passed."}
					}
				}
			case types.TransactionType_ChangeContentAccess:
				changeAccess, ok := blockTx.Data.(*types.ChangeAccess)
				if ok && changeAccess.Cid == cidStr && blockTx.Identity == requestTx.Identity{
					return &abci.ResponseCheckTx{Code: uint32(types.CodeType_OK), Log: "All checks passed."}
				}
			}
		}
	}
	
	// All checks passed. Return OK.
	return &abci.ResponseCheckTx{Code: uint32(types.CodeType_BCFSInvalidBlockHeight), Log: "Could not verify ownership of data."}
}
func (app *Application) DeliverTx_RemoveData(signer *conf.Identity, tx *types.Transaction) *abci.ResponseDeliverTx {
	cidStr, ok := tx.Data.(string)
	if !ok {
		return &abci.ResponseDeliverTx{Code: uint32(types.CodeType_BCFSInvalidInput), Log: "Could not type assert Data."}
	}

	// Update Prevailing block height
	app.prevailingBlock[cidStr] = app.nextBlockHeight

	// All checks passed. Return OK.
	return &abci.ResponseDeliverTx{Code: uint32(types.CodeType_OK), Log: "All checks passed."}
}