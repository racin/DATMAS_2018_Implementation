package app

import (
	"time"

	"github.com/trusch/passchain/crypto"

	mh "github.com/multiformats/go-multihash"
	"fmt"
)

type Transaction struct {
	Type      TransactionType `json:"type"`
	Timestamp time.Time       `json:"timestamp"`
	Signature string          `json:"signature"`
	Data      interface{}     `json:"data"`
}

type Hashable interface {
	Hash() []byte
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
	ret, _ := mh.Sum(data, mh.SHA2_256, -1)
	return ret.B58String()
}

func (t *Transaction) Sign(key *crypto.Key) error {
	hash := t.Hash()
	signature, err := key.Sign([]byte(hash))
	if err != nil {
		return err
	}
	t.Signature = signature
	return nil
}

func (t *Transaction) Verify(key *crypto.Key) error {
	hash := t.Hash()
	return key.Verify([]byte(hash), t.Signature)
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


func New(t TransactionType, data interface{}) *Transaction {
	return &Transaction{Type: t, Timestamp: time.Now(), Data: data}
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