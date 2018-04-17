package crypto

import (
	"crypto/rsa"
	"crypto/rand"
	"github.com/stretchr/testify/assert"
	"testing"
	conf "github.com/racin/DATMAS_2018_Implementation/configuration"
)

const (
	certName 				= "mycert_test"
	testCertPath			= "test_certificate/"
	clientCertPathTest 		= "test_certificate/client_test"
	clientCertPathFP		= "95c73e8028118d18a961dd1da6b5e7c3"
	storageCertPathTest 	= "test_certificate/storage_test"
	storageCertPathFP		= "64168bb2f7a0a4d67d83471470ce757c"
	consensusCertPathTest 	= "test_certificate/consensus_test"
	consensusCertPathFP		= "cc418e456ae72df5bdb39d65bb8945e8"
	testKeysBits	= 1024
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
		if _, err := GenerateKeyPair(conf.AppConfig().BasePath, certName , testKeysBits); err != nil {
			certPath = clientCertPathTest
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
		keypair, _ := rsa.GenerateKey(rand.Reader, testKeysBits)
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

		signature, err := privKey.Sign(hashData);
		if err == nil {
			verify := pubKey.Verify(hashData, signature)
			assert.True(t, verify, "Signature did not match.")
		} else {
			t.Fatal("Could not sign data. Error: " + err.Error())
		}
		signature2, err := privKey.Sign(hashData);
		assert.NotEqual(t, signature, signature2, "Signatures is not properly salted.")
	})
	t.Run("SignHashStruct", func(t *testing.T){
		trans := TestHashStruct{Data:[]byte("Test data to sign and verify"), Number: 123, Message:"abc"}
		if err != nil {
			t.Fatal("Could not hash data. Error: " + err.Error())
		}

		signedTrans, err := SignStruct(trans, privKey);
		if err == nil {
			verify := signedTrans.Verify(pubKey)
			assert.True(t, verify, "Signature did not match.")
		} else {
			t.Fatal("Could not sign data. Error: " + err.Error())
		}
		signedTrans2, err := SignStruct(trans, privKey);
		assert.NotEqual(t, signedTrans, signedTrans2, "Signatures is not properly salted.")
	})
}