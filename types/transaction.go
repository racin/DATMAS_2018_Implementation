package types

import (
	"time"
	"crypto/rand"
	"math/big"
	"math"
	"github.com/racin/DATMAS_2018_Implementation/crypto"
	"encoding/json"
	"errors"
	"fmt"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
)

type Transaction struct {
	Data      	interface{}     `json:"data"`
	Identity	string			`json:"identity"`
	Type      	TransactionType `json:"type"`
	Timestamp 	string       	`json:"timestamp"`
	Nonce		uint64			`json:"nonce"`
}

func NewTx(data interface{}, identity string, t TransactionType) *Transaction {
	nonce, err := rand.Int(rand.Reader, new(big.Int).SetUint64(math.MaxUint64)) // 1 << 64 - 1
	if err != nil {
		nonce = big.NewInt(0)
	}
	return &Transaction{Data: data, Identity: identity, Type: t,
		Timestamp: time.Now().Format(time.RFC3339), Nonce:nonce.Uint64()}

}

func UnmarshalTransaction(txBytes []byte) (*crypto.SignedStruct, *Transaction, error) {
	stx := &crypto.SignedStruct{Base: &Transaction{}}
	if err := json.Unmarshal(txBytes, stx); err != nil {
		return nil, nil, err
	} else if tx, ok := stx.Base.(*Transaction); !ok {
		return nil, nil, errors.New("Could not unmarshal transaction (Transaction)")
	} else {
		// Check if the data sent is actually another Struct.


		// If its not, we can simply return and the different transaction types will get the value themselves.
		if derivedStruct, ok := stx.Base.(*Transaction).Data.(map[string]interface{}); ok {
			fmt.Printf("DerivedStruct: %+v\n", derivedStruct)

			if signedReqUpload := GetSignedRequestUploadFromMap(derivedStruct); signedReqUpload != nil { // Signed types.RequestUpload
				fmt.Printf("Signed ReqUpload: %+v\n", signedReqUpload)
				stx.Base.(*Transaction).Data = signedReqUpload
				tx.Data = signedReqUpload
			} else if reqUpload := GetRequestUploadFromMap(derivedStruct); reqUpload != nil { // types.RequestUpload
				fmt.Printf("ReqUpload: %+v\n", reqUpload)
				stx.Base.(*Transaction).Data = reqUpload
				tx.Data = reqUpload
			}
		} else if derivedArray, ok := stx.Base.(*Transaction).Data.([]interface{}); ok {
			fmt.Printf("DerivedArray: %+v\n", derivedArray)
			if storProofArr := crypto.GetStorageChallengeProofArray(derivedArray); storProofArr != nil { // Signed types.RequestUpload
				fmt.Printf("Storage Proof array: %+v\n", storProofArr)
				stx.Base.(*Transaction).Data = storProofArr
				tx.Data = storProofArr
			}
		}



		return stx, tx, nil
	}
}

func CheckBroadcastResult(commit interface{}, err error) (CodeType, error) {
	fmt.Printf("Data 1: %+v\n", commit)
	if c, ok := commit.(*ctypes.ResultBroadcastTxCommit); ok && c != nil {
		if err != nil {
			return CodeType_InternalError, err
		} else if c.CheckTx.IsErr() {
			return CodeType_InternalError, errors.New(c.CheckTx.String())
		} else if c.DeliverTx.IsErr() {
			return CodeType_InternalError, errors.New(c.DeliverTx.String())
		} else {
			fmt.Printf("Data: %+v\n", c)
			return CodeType_OK, nil;
		}
	} else if c, ok := commit.(*ctypes.ResultBroadcastTx); ok && c != nil {
		fmt.Printf("Data 2: %+v\n", c)
		code := CodeType(c.Code)
		if code == CodeType_OK {
			fmt.Printf("Data 3: %+v\n", c)
			return code, nil
		} else {
			return code, errors.New("CheckTx. Log: " + c.Log + ", Code: " + CodeType_name[int32(c.Code)])
		}
	}
	/**/
	return CodeType_InternalError, errors.New("Could not type assert result.")
}