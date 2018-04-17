package crypto

import (
	"testing"
	"github.com/stretchr/testify/assert"

	conf "github.com/racin/DATMAS_2018_Implementation/configuration"
	"io/ioutil"
	"fmt"
)

const (
	testPosPath = "test_pos/"
	aclTestPath = "../configuration/" + conf.ListPathTest
)
func TestStorageSample(t *testing.T){
	acl := conf.GetAccessList(aclTestPath)
	byteArr, err := ioutil.ReadFile(testPosPath + "RandData")
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
		if err := storageSample.StoreSample(testPosPath, cid); err != nil {
			t.Fatal("Could not store storage sample.")
		}
		storageSample = nil;
	})
	t.Run("LoadStorageSample", func(t *testing.T){
		assert.Nil(t, storageSample, "storageSample is not set to nil.")
		storageSample = LoadStorageSample("test_pos/", cid)
		assert.NotNil(t, storageSample, "Could not load Storage Sample")
		assert.NotEmpty(t, storageSample.Samples, "Could not load samples.")
	})

	var challenge *SignedStruct
	t.Run("GenerateChallenge", func(t *testing.T){
		privKey, err := LoadPrivateKey(clientCertPathTest + ".pem")
		if err != nil {
			t.Fatal("Could not load private key. Error: " + err.Error())
		}
		challenge = storageSample.GenerateChallenge(privKey, cid)
		fmt.Printf("Challenge: %v, %+v\n", len(storageSample.Samples), challenge)
		assert.NotNil(t, challenge, "Could not load Storage Sample")
		chalng, ok := challenge.Base.(*StorageChallenge)
		fmt.Printf("Chill challenge: %v, %+v\n", len(storageSample.Samples), chalng)
		assert.False(t, ok, "Could not type assert SignedStruct to StorageChallenge")
		assert.NotEmpty(t, chalng.Challenge, "Challenge was empty")
	})
	t.Run("VerifyChallenge", func(t *testing.T){
		challengerIdent, ok := acl.Identities[consensusCertPathFP]
		if !ok {
			t.Fatal("Could not find identity: " + consensusCertPathFP)
		}
		err = challenge.VerifyChallenge(challengerIdent)
		if err != nil {
			t.Fatal("Could not verify the challenge. Error: " + err.Error())
		}
	})
	var challengeProof *SignedStruct
	t.Run("GenerateChallengeProof", func(t *testing.T){
		privKey, err := LoadPrivateKey(storageCertPathTest + ".pem")
		if err != nil {
			t.Fatal("Could not load private key. Error: " + err.Error())
		}

		challengeProof = challenge.ProveChallenge(privKey, &byteArr)
		if challengeProof == nil {
			t.Fatal("Could not generate challenge proof.")
		}
	})
	t.Run("VerifyChallengeProof", func(t *testing.T){
		challengerIdent, ok := acl.Identities[consensusCertPathFP]
		if !ok {
			t.Fatal("Could not find challenger identity in access list.")
		}
		proverIdent, ok := acl.Identities[storageCertPathFP]
		if !ok {
			t.Fatal("Could not find prover identity in access list.")
		}

		challengeProof := challenge.VerifyChallengeProof("", &challengerIdent, &proverIdent)
		if challengeProof == nil {
			t.Fatal("Could not generate challenge proof.")
		}
	})
}
func TestGenerateStorageSample(t *testing.T) {

}

func TestGenerateChallenge(t *testing.T){

}

func TestVerifyChallengeProof(t *testing.T){

}