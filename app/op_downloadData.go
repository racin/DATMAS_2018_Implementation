package app

import (
	"github.com/racin/DATMAS_2018_Implementation/types"
	abci "github.com/tendermint/abci/types"
	conf "github.com/racin/DATMAS_2018_Implementation/configuration"
)

func (app *Application) CheckTx_DownloadData(signer *conf.Identity, tx *types.Transaction) *abci.ResponseCheckTx {
	// All checks passed. Return OK.
	return &abci.ResponseCheckTx{Code: uint32(types.CodeType_OK), Log: "All checks passed."}
}
func (app *Application) DeliverTx_DownloadData(signer *conf.Identity, tx *types.Transaction) *abci.ResponseDeliverTx {
	// All checks passed. Return OK.
	return &abci.ResponseDeliverTx{Code: uint32(types.CodeType_OK), Log: "All checks passed."}
}
