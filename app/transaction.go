package app

import (
	"time"
	"github.com/racin/DATMAS_2018_Implementation/crypto"
	"fmt"
)

type Transaction struct {
	Data      	interface{}     `json:"data"`
	Identity	string			`json:"identity"`
	Type      	TransactionType `json:"type"`
	Timestamp 	string       	`json:"timestamp"`
}

type SignedTransaction struct {
	Transaction
	Signature 	[]byte          `json:"signature"`
}


type TransactionType string

const (
	DownloadData		TransactionType = "data-download"
	UploadData     		TransactionType = "data-upload"
	RemoveData      	TransactionType = "data-remove"
	VerifyStorage		TransactionType = "data-verify"
	ChangeContentAccess	TransactionType = "data-access"
)


/*
func (t *Transaction) FromBytes(bs []byte) error {
	return json.Unmarshal(bs, t)
}

func (t *Transaction) ToBytes() ([]byte, error) {
	return json.Marshal(t)
}*/

func (t *Transaction) Hash() string {
	data := []byte(fmt.Sprintf("%v", t))
	hash, _ := crypto.HashData(data)
	fmt.Println(hash)
	return hash
}

func (t *Transaction) Sign(keys *crypto.Keys) (*SignedTransaction, error) {
	if signature, err := keys.Sign(t.Hash()); err != nil {
		return nil, err
	} else {
		return &SignedTransaction{Transaction: *t, Signature: signature}, nil
	}
}

func (t *SignedTransaction) Verify(keys *crypto.Keys) bool {
	return keys.Verify(t.Hash(), t.Signature)
}
/*
func (t *Transaction) ProofOfWork(cost byte) error {
	for round := 0; round < (1 << 32); round++ {
		t.Nonce = uint32(round)
		if err := t.VerifyProofOfWork(cost); err == nil {
			return nil
		}
	}
	return errors.New("can not find pow")
}*/

func NewTx(data interface{}, identity string, t TransactionType) *Transaction {
	return &Transaction{Data: data, Identity: identity, Type: t, Timestamp: time.Now().Format(time.RFC3339)}

}
/*
func hashStringMap(m map[string]interface{}) []byte {
	hash := sha3.New512()
	encoder := json.NewEncoder(hash)
	keys := make([]string, len(m))
	i := 0
	for id := range m {
		keys[i] = id
		i++
	}
	sort.Strings(keys)
	for _, key := range keys {
		encoder.Encode(key)
		encoder.Encode(m[key])
	}
	return hash.Sum(nil)
}
*/