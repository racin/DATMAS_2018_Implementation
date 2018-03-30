package crypto

import (
	"crypto/rsa"
	"crypto/rand"
	"testing"
)

func TestSignature(t *testing.T){
	// Generate a set of random keys
	a, _ := rsa.GenerateKey(rand.Reader, 2048)
	key := &Keys{private: a, public: &a.PublicKey}

	
}