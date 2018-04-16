package crypto

import (
	"testing"
	"github.com/stretchr/testify/assert"
	//conf "github.com/racin/DATMAS_2018_Implementation/configuration"
	"io/ioutil"
	"fmt"
)

const (
	certPathTest 	= "test_certificate/mycert_test"
)
func TestStorageSample(t *testing.T){
	byteArr, err := ioutil.ReadFile("test_pos/RandData")
	if err != nil {
		t.Fatal("Error reading test file: " + err.Error())
	}
	cid, err := IPFSHashData(byteArr)
	if err != nil {
		t.Fatal("Could not get hash of file.")
	}
	var storageSample *StorageSample
	t.Run("TestGenerateStorageSample", func(t *testing.T){
		storageSample = GenerateStorageSample(&byteArr)
		if storageSample == nil {
			t.Fatal("Error generating storage sample.")
		}
		assert.NotEmpty(t, storageSample.Samples, "No samples generated.")
	})
	fmt.Printf("%v, %+v\n", len(storageSample.Samples), storageSample)
	t.Run("StoreStorageSample", func(t *testing.T){
		if err := storageSample.StoreSample("test_pos/", cid); err != nil {
			t.Fatal("Could not store storage sample.")
		}
		storageSample = nil;
	})
	t.Run("LoadStorageSample", func(t *testing.T){
		assert.Nil(t, storageSample, "storageSample is not set to nil.")
		storageSample := LoadStorageSample("test_pos/", cid)
		assert.NotNil(t, storageSample, "Could not load Storage Sample")
		assert.NotEmpty(t, storageSample.Samples, "Could not load samples.")
	})
	var privKey *Keys
	var pubKey *Keys
	var challenge *StorageChallenge
	t.Run("GenerateChallenge", func(t *testing.T){
		privKey, err = LoadPrivateKey(certPathTest + ".pem")
		if err != nil {
			t.Fatal("Could not load private key. Error: " + err.Error())
		}

		pubKey, err = LoadPublicKey(certPathTest + ".pub")
		if err != nil {
			t.Fatal("Could not load public key. Error: " + err.Error())
		}
		challenge := storageSample.GenerateChallenge(privKey)
		assert.Nil(t, storageSample, "storageSample is not set to nil.")
		storageSample := LoadStorageSample("test_pos/", cid)
		assert.NotNil(t, storageSample, "Could not load Storage Sample")
		assert.NotEmpty(t, storageSample.Samples, "Could not load samples.")
	})
}
func TestGenerateStorageSample(t *testing.T) {

}

func TestGenerateChallenge(t *testing.T){

}

func TestVerifyChallengeProof(t *testing.T){

}