package types

import (
	"time"
	"crypto/rand"
	"math/big"
	"math"
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