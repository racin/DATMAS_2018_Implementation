package crypto

import (
	"crypto/rsa"
	"crypto/rand"
	"github.com/stretchr/testify/assert"
	"testing"
	"os/user"
)

const (
	keyPath = "/.bcfs/"
	certName = "mycert_test"
)
func TestSignature(t *testing.T){
	var privKey *Keys
	var pubKey *Keys
	var err error
	var certPath string
	t.Run("TestGenerateKeys", func(t *testing.T){
		usr, err := user.Current()
		if err != nil {
			t.Fatal("Could not get current user")
			return
		}
		if _, err := GenerateKeyPair(usr.HomeDir + keyPath, certName , 1024); err != nil {
			certPath = "test_certificate/mycert_test"
			t.Fatal("Error: " + err.Error())
		} else {
			certPath = usr.HomeDir + keyPath + certName
		}
	})
	
	t.Run("LoadSignature", func(t *testing.T){
		privKey, err = LoadPrivateKey(certPath + ".pem")
		if err != nil {
			t.Fatal("Could not load private key. Error: " + err.Error())
		}

		pubKey, err = LoadPublicKey(certPath + ".pub")
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