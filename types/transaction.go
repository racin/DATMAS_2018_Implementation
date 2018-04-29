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
		derivedStruct, ok := stx.Base.(*Transaction).Data.(map[string]interface{})

		// If its not, we can simply return and the different transaction types will get the value themselves.
		if !ok {
			return stx, tx, nil
		}

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

		return stx, tx, nil
	}
}