package crypto

import (
	"crypto/rsa"
	"crypto/rand"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSignature(t *testing.T) {
	// Generate a set of random keys
	keypair, _ := rsa.GenerateKey(rand.Reader, 2048)
	key := &Keys{private: keypair, public: &keypair.PublicKey}

	data := []byte("Test data to sign and verify")
	hashData, err := IPFSHashData(data)
	if err != nil {
		t.Fatal("Could not hash data. Error: " + err.Error())
	}

	if signature, err := key.Sign(hashData); err == nil {
		verify := key.Verify(hashData, signature)
		assert.True(t, verify, "Signature did not match.")
	} else {
		t.Fatal("Could not sign data. Error: " + err.Error())
	}
}