package app

import (
	"encoding/json"
	"github.com/racin/DATMAS_2018_Implementation/crypto"
	"github.com/racin/DATMAS_2018_Implementation/types"
	abci "github.com/tendermint/abci/types"
	conf "github.com/racin/DATMAS_2018_Implementation/configuration"
	tmtypes "github.com/tendermint/tendermint/types"
	"fmt"
)

func (app *Application) Query_Challenge(reqQuery abci.RequestQuery) *abci.ResponseQuery{
	fmt.Println("Query challenge")
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

	//lenStorNodes := len(conf.AppConfig().IpfsNodes)
	proofs := make([]crypto.SignedStruct, 0)

	// Issue the challenge from the Client first
	for _, ident := range conf.AppConfig().IpfsNodes {
		addr := app.GetAccessList().GetAddress(ident)
		ipfsResp := app.queryIPFSproxy(addr, conf.AppConfig().IpfsChallengeEndpoint, signedStruct)
		fmt.Printf("IpfsResp: %v\n", ipfsResp)

		if (ipfsResp.Codetype != types.CodeType_OK) {
			continue // Not a valid proof. Do not care about why
		}
		scp := &crypto.SignedStruct{Base:&crypto.StorageChallengeProof{SignedStruct: crypto.SignedStruct{Base:&crypto.StorageChallenge{}}}}
		if err := json.Unmarshal(ipfsResp.Message, scp); err == nil {
			proofs = append(proofs, *scp)
		}
	}
	fmt.Printf("Proofs: %v\n", proofs)
	// Then the randomly generated ones
	for _, ident := range conf.AppConfig().IpfsNodes {
		addr := app.GetAccessList().GetAddress(ident)
		ipfsResp := app.queryIPFSproxy(addr, conf.AppConfig().IpfsChallengeEndpoint, signRndChal)
		fmt.Printf("IpfsResp: %v\n", ipfsResp)
		if (ipfsResp.Codetype != types.CodeType_OK) {
			continue // Not a valid proof. Do not care about why
		}
		scp := &crypto.SignedStruct{Base:&crypto.StorageChallengeProof{SignedStruct: crypto.SignedStruct{Base:&crypto.StorageChallenge{}}}}
		if err := json.Unmarshal(ipfsResp.Message, scp); err == nil {
			proofs = append(proofs, *scp)
		}
	}
	fmt.Printf("Proofs: %v\n", proofs)
	// Now lets send the proofs to the mempool
	tx := types.NewTx(proofs, app.fingerprint, types.TransactionType_VerifyStorage)
	stx, err := crypto.SignStruct(tx, app.privKey)
	if err != nil {
		return &abci.ResponseQuery{Code: uint32(types.CodeType_BCFSInvalidSignature), Log: "Could not sign StorageChallengeProofs"}
	}

	stxByteArr, err := json.Marshal(stx)
	fmt.Printf("TX: %+v\n", tx.Data)
	fmt.Printf("Stx: %+v\n", stx)
	fmt.Printf("Stx Base: %+v\n", stx.Base)
	fmt.Printf("Hash of STX base: %v\n", crypto.HashStruct(stx.Base))
	if err != nil {
		return &abci.ResponseQuery{Code: uint32(types.CodeType_InternalError), Log: "Error marshalling: Error: " + err.Error()}
	}

	// TODO: Setup Mempool connection.
	//res := app.CheckTx(stxByteArr)
	//fmt.Printf("CheckTx result: %+v\n", res)
	// Sends the transaction to itself though the RPC client
	fmt.Println("-----------------------------------------------------------")
	fmt.Println("-----------------------------------------------------------")
	fmt.Println("-----------------------------------------------------------")
	fmt.Println("-----------------------------------------------------------")

	stx2, tx2, err := types.UnmarshalTransaction(stxByteArr)
	tx2.Data = proofs
	fmt.Printf("TX2: %+v\n", tx2.Data)
	fmt.Printf("Stx2: %+v\n", stx2)
	fmt.Printf("Hash of STX2 base: %v\n", crypto.HashStruct(stx2.Base))
	fmt.Println("-----------------------------------------------------------")
	fmt.Println("-----------------------------------------------------------")
	fmt.Println("-----------------------------------------------------------")
	fmt.Println("-----------------------------------------------------------")

	codetype, err := types.CheckBroadcastResult(app.TMRpcClients[app.fingerprint].BroadcastTxSync(tmtypes.Tx(stxByteArr)))
	fmt.Printf("CodeType: %v, Error: %v", codetype, err)

	return &abci.ResponseQuery{Code: uint32(types.CodeType_OK), Log: "Transaction with proofs sent to mempool. Wait for commit."}
}