package types

import (
	"time"
)

type Transaction struct {
	Data      	interface{}     `json:"data"`
	Identity	string			`json:"identity"`
	Type      	TransactionType `json:"type"`
	Timestamp 	string       	`json:"timestamp"`
}

type TransactionType string
const (
	// Tendermint
	DownloadData		TransactionType = "download"
	UploadData     		TransactionType = "upload"
	RemoveData      	TransactionType = "removedata"
	VerifyStorage		TransactionType = "verifystorage"
	ChangeContentAccess	TransactionType = "changeaccess"

	// IPFS Proxy
	IPFSProxy			TransactionType = "ipfsproxy"
)

/*
func (t *Transaction) Sign(keys *crypto.Keys) (*SignedTransaction, error) {
	if signature, err := keys.Sign(t.Hash()); err != nil {
		return nil, err
	} else {
		return &SignedTransaction{Base: *t, Signature: signature}, nil
	}
}

func (t *SignedTransaction) Verify(keys *crypto.Keys) bool {
	if hashable, ok := t.Base.(Hashable); ok {
		return keys.Verify(hashable.Hash(), t.Signature)
	}
	return false;
}*/

func NewTx(data interface{}, identity string, t TransactionType) *Transaction {
	return &Transaction{Data: data, Identity: identity, Type: t, Timestamp: time.Now().Format(time.RFC3339)}

}