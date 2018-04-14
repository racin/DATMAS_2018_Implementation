package crypto

import (
	"crypto/rsa"
	"crypto/rand"
	"github.com/stretchr/testify/assert"
	"testing"
	conf "github.com/racin/DATMAS_2018_Implementation/configuration"
	"fmt"
)

const (
	certName 		= "mycert_test"
	certPathTest 	= "test_certificate/mycert_test"
)

func TestSignature(t *testing.T){
	if _, err := conf.LoadAppConfig("../configuration/test/appConfig"); err != nil {
		t.Fatal("Error loading app config: " + err.Error())
	}
	var privKey *Keys
	var pubKey *Keys
	var err error
	var certPath string
	t.Run("TestGenerateKeys", func(t *testing.T){
		if _, err := GenerateKeyPair(conf.AppConfig().BasePath, certName , 1024); err != nil {
			certPath = certPathTest
			t.Fatal("Error: " + err.Error())
		} else {
			certPath = conf.AppConfig().BasePath + certName
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
	t.Run("SignHashStruct", func(t *testing.T){
		// Generate a set of random keys
		trans := &TestHashStruct{Signable: Signable{}, Data:[]byte("Test data to sign and verify"), Number: 123, Message:"abc"}
		if err != nil {
			t.Fatal("Could not hash data. Error: " + err.Error())
		}
		fmt.Printf("%v\n", trans)
		_ = trans.Hash()
		//HashStruct(trans)

		if signedTrans, err := trans.Sign(privKey); err == nil {
			verify := signedTrans.Verify(pubKey)
			assert.True(t, verify, "Signature did not match.")
		} else {
			t.Fatal("Could not sign data. Error: " + err.Error())
		}
	})
}