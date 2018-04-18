package crypto

import (
	"testing"
	"github.com/stretchr/testify/assert"

	conf "github.com/racin/DATMAS_2018_Implementation/configuration"
	"io/ioutil"
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
	var signedStorageSample *SignedStruct
	t.Run("SignSample", func(t *testing.T){
		privKey, err := LoadPrivateKey(consensusCertPathTest + ".pem")
		if err != nil {
			t.Fatal("Could not load private key. Error: " + err.Error())
		}
		signedStorageSample, err = storageSample.SignSample(privKey)
		if err != nil {
			t.Fatal("Could not sign storage sample. Error: " + err.Error())
		}
		assert.NotNil(t, signedStorageSample, "Could not sign storage sample.")
	})
	t.Run("StoreStorageSample", func(t *testing.T){
		if err := signedStorageSample.StoreSample(testPosPath); err != nil {
			t.Fatal("Could not store storage sample.")
		}
		storageSample = nil;
	})
	t.Run("LoadStorageSample", func(t *testing.T){
		assert.Nil(t, storageSample, "storageSample is not set to nil.")
		storageSample = LoadStorageSample("test_pos/", cid)
		if storageSample == nil || storageSample.Samples == nil || len(storageSample.Samples) == 0 {
			t.Fatal("Could not load Storage Sample")
		}
	})
	var challenge *SignedStruct
	t.Run("GenerateChallenge", func(t *testing.T){
		privKey, err := LoadPrivateKey(consensusCertPathTest + ".pem")
		if err != nil {
			t.Fatal("Could not load private key. Error: " + err.Error())
		}
		challenge = storageSample.GenerateChallenge(privKey, cid)
		assert.NotNil(t, challenge, "Could not load Storage Sample")
		chalng, ok := challenge.Base.(*StorageChallenge)

		assert.True(t, ok, "Could not type assert SignedStruct to StorageChallenge")
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
		assert.NotNil(t, challengeProof, "Could not generate challenge proof.")
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

		err := challengeProof.VerifyChallengeProof(testPosPath, &challengerIdent, &proverIdent)
		if err != nil {
			t.Fatal("Could not verify challenge proof. Error: " + err.Error())
		}
	})
	t.Run("VerifyChallengeWithWrongIdentity", func(t *testing.T){
		challengerIdent, ok := acl.Identities[clientCertPathFP]
		if !ok {
			t.Fatal("Could not find identity: " + clientCertPathFP)
		}

		assert.NotNil(t, challenge.VerifyChallenge(challengerIdent), "Challenge verified using wrong identity")
	})
	var challengeProofWithWrongIdentity *SignedStruct
	t.Run("GenerateChallengeProofWithWrongIdentity", func(t *testing.T){
		privKey, err := LoadPrivateKey(clientCertPathTest + ".pem")
		if err != nil {
			t.Fatal("Could not load private key. Error: " + err.Error())
		}

		challengeProofWithWrongIdentity = challenge.ProveChallenge(privKey, &byteArr)
		assert.NotNil(t, challengeProofWithWrongIdentity, "Could not generate challenge proof.")
	})
	t.Run("VerifyChallengeProofWithWrongIdentity", func(t *testing.T){
		challengerIdent, ok := acl.Identities[consensusCertPathFP]
		if !ok {
			t.Fatal("Could not find challenger identity in access list.")
		}
		proverIdent, ok := acl.Identities[storageCertPathFP]
		if !ok {
			t.Fatal("Could not find prover identity in access list.")
		}

		err := challengeProofWithWrongIdentity.VerifyChallengeProof(testPosPath, &challengerIdent, &proverIdent)
		assert.NotNil(t, err, "Proof should not be verifiable using a different public key. " +
			"(Client signed the proof.)")
	})
}