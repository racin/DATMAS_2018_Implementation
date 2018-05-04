package types

import (
	"time"
	"crypto/rand"
	"math/big"
	"math"
	"github.com/racin/DATMAS_2018_Implementation/crypto"
	"encoding/json"
	"errors"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
)

type Transaction struct {
	Data      	interface{}     `json:"data"`
	Identity	string			`json:"identity"`
	Type      	TransactionType `json:"type"`
	Timestamp 	string       	`json:"timestamp"`
	Nonce		uint64			`json:"nonce"`
}

type ChangeAccess struct {
	Cid			string		`json:"cid"`
	Readers		[]string	`json:"readers"`
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
		// Check if the data sent is actually another Struct or Slice. Have to do some heavy conversion if thats the case.
		// If its not, we can simply return and the different transaction types will get the value themselves.

		if derivedStruct, ok := stx.Base.(*Transaction).Data.(map[string]interface{}); ok {
			if signedReqUpload := GetSignedRequestUploadFromMap(derivedStruct); signedReqUpload != nil { // Signed types.RequestUpload
				stx.Base.(*Transaction).Data = signedReqUpload
				tx.Data = signedReqUpload
			} else if reqUpload := GetRequestUploadFromMap(derivedStruct); reqUpload != nil { // types.RequestUpload
				stx.Base.(*Transaction).Data = reqUpload
				tx.Data = reqUpload
			} else if changeAccess := GetChangeAccessFromMap(derivedStruct); changeAccess != nil {
				stx.Base.(*Transaction).Data = changeAccess
				tx.Data = changeAccess
			}
		} else if derivedArray, ok := stx.Base.(*Transaction).Data.([]interface{}); ok {
			if storProofArr := crypto.GetStorageChallengeProofArray(derivedArray); storProofArr != nil { // Signed types.RequestUpload
				stx.Base.(*Transaction).Data = storProofArr
				tx.Data = storProofArr
			}
		}

		return stx, tx, nil
	}
}

func CheckBroadcastResult(commit interface{}, err error) (CodeType, error) {
	if c, ok := commit.(*ctypes.ResultBroadcastTxCommit); ok && c != nil {
		if err != nil {
			return CodeType_InternalError, err
		} else if c.CheckTx.IsErr() {
			return CodeType_InternalError, errors.New(c.CheckTx.String())
		} else if c.DeliverTx.IsErr() {
			return CodeType_InternalError, errors.New(c.DeliverTx.String())
		} else {
			return CodeType_OK, nil;
		}
	} else if c, ok := commit.(*ctypes.ResultBroadcastTx); ok && c != nil {
		code := CodeType(c.Code)
		if code == CodeType_OK {
			return code, nil
		} else {
			return code, errors.New("CheckTx. Log: " + c.Log + ", Code: " + CodeType_name[int32(c.Code)])
		}
	}
	return CodeType_InternalError, errors.New("Could not type assert result.")
}

func GetChangeAccessFromMap(derivedStruct map[string]interface{}) *ChangeAccess {
	if cid, ok := derivedStruct["cid"]; ok {
		if readers, ok := derivedStruct["readers"]; ok {
			if readersArr, ok := readers.([]interface{}); ok {
				rdrs := make([]string, len(readersArr))
				for i, val := range readersArr {
					rdrs[i] = val.(string)
				}
				return &ChangeAccess{Cid: cid.(string), Readers:rdrs}
			}
		}
	}
	return nil
}