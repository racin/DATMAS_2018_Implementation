package app

import (
	"encoding/json"
	"github.com/racin/DATMAS_2018_Implementation/types"
	abci "github.com/tendermint/abci/types"
	"fmt"
)

func (app *Application) Query_PrevailingHeight(reqQuery abci.RequestQuery) *abci.ResponseQuery{
	fmt.Println("Query prevailingheight")
	if reqQuery.Data == nil {
		return &abci.ResponseQuery{Code: uint32(types.CodeType_BCFSInvalidInput), Log: "Missing data parameter."}
	}

	cid := string(reqQuery.Data)

	prevBlockHeight, ok := app.prevailingBlock[cid]
	if !ok {
		return &abci.ResponseQuery{Code: uint32(types.CodeType_Unauthorized), Log: "File with requested CID is not found in the system."}
	}

	stx := app.GetSignedTransaction(types.TransactionType_PrevailingHeight, prevBlockHeight)
	stxByteArr, err := json.Marshal(stx)
	if err != nil {
		return &abci.ResponseQuery{Code: uint32(types.CodeType_InternalError), Log: "Error marshalling: Error: " + err.Error()}
	}

	return &abci.ResponseQuery{Code: uint32(types.CodeType_OK), Log: string(stxByteArr)}
}