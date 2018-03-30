package crypto

import (
	"crypto/rsa"
	"crypto/rand"
	"github.com/stretchr/testify/assert"
	"testing"
	"fmt"
)

func TestSignature(t *testing.T) {
	// Generate a set of random keys
	keypair, _ := rsa.GenerateKey(rand.Reader, 2048)
	key := &Keys{private: keypair, public: &keypair.PublicKey}

	data := []byte("Test data to sign and verify")
	hashData, err := IPFSHashData(data)
	if err != nil {
		t.Fatal("Could not hash data")
	}


	fmt.Printf("Keys: %+v\n", keypair)
	fmt.Printf("Key: %+v\n", key)
	fmt.Printf("Data: %s\n", hashData)
	signature := key.Sign(hashData)
	fmt.Printf("Sign: %s\n", signature)
	verify := key.Verify(hashData, signature)


	assert.True(t, verify, "Signature did not match.")
}