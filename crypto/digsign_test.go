package crypto

import (
	"crypto/rsa"
	"crypto/rand"
	"github.com/stretchr/testify/assert"
	"testing"
	"os/user"
)

func TestSignature(t *testing.T){
	var privKey *Keys
	var pubKey *Keys
	var err error
	t.Run("LoadSignature", func(t *testing.T){
		privKey, err = LoadPrivateKey("test_certificate/mycert_test.pem")
		if err != nil {
			t.Fatal("Could not load private key. Error: " + err.Error())
		}

		pubKey, err = LoadPublicKey("test_certificate/mycert_test.pub")
		if err != nil {
			t.Fatal("Could not load public key. Error: " + err.Error())
		}
	})
	// If keys could not be loaded, generate a random one for testing.
	if privKey == nil || privKey.private == nil || pubKey == nil || pubKey.public == nil {
		keypair, _ := rsa.GenerateKey(rand.Reader, 1024)
		privKey = &Keys{private: keypair, public: &keypair.PublicKey}
		pubKey = &Keys{public: &keypair.PublicKey}
	}
	t.Run("SignData", func(t *testing.T){
		// Generate a set of random keys
		data := []byte("Test data to sign and verify")
		hashData, err := IPFSHashData(data)
		if err != nil {
			t.Fatal("Could not hash data. Error: " + err.Error())
		}

		if signature, err := privKey.Sign(hashData); err == nil {
			verify := pubKey.Verify(hashData, signature)
			assert.True(t, verify, "Signature did not match.")
		} else {
			t.Fatal("Could not sign data. Error: " + err.Error())
		}
	})
}

func TestGenerateKeys(t *testing.T){
	if usr, err := user.Current(); err != nil {
		t.Fatal("Could not get current user")
	} else if err := GenerateKeyPair(usr.HomeDir + "/.bcfs/mycert.pem", 1024); err != nil {
		t.Fatal("Error: " + err.Error())
	}
}