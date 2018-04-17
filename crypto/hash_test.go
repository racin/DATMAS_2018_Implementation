package crypto

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

type TestHashStruct struct {
	Message			string
	Number			int
	Data			[]byte
}

func TestIPFSHash(t *testing.T){
	fileHash, err := IPFSHashFile("hash_test.txt")
	if err != nil {
		t.Fatal("IPFSHash failed: " + err.Error())
	}
	assert.Equal(t, "QmRFq5YyyNai59Pvxfd5pGJY6HzpubyzpDp6ceqbJfDUBp", fileHash, "File hash did not match")

	textHash, err := IPFSHashData([]byte("racin\n"))
	if err != nil {
		t.Fatal("IPFSHash failed: " + err.Error())
	}
	assert.Equal(t, "QmdeVbypiSbW24Uhjdvdhczpv1gxDXA9nPYKGiPaAnQs5F", textHash, "Data hash did not match")
}

func TestHash(t *testing.T){
	fileHash, err := HashFile("hash_test.txt")
	if err != nil {
		t.Fatal("Hash failed: " + err.Error())
	}
	assert.Equal(t, "QmWimMR7o684LGMTHRrCptBDbK3sP3m2ocYwmUUzYVKwfm", fileHash, "File hash did not match")

	textHash, err := HashData([]byte("racin"))
	if err != nil {
		t.Fatal("Hash failed: " + err.Error())
	}
	assert.Equal(t, "QmP4eE9BBqDRHrPbwFN75M9cX84Rm3G8B2fKtxZCtREUyC", textHash, "Data hash did not match")
}

func TestStructHash(t *testing.T) {
	hashstruct := TestHashStruct{Data:[]byte("Test data to sign and verify"), Number: 123, Message:"abc"}
	hash1 := HashStruct(hashstruct);
	hashstruct.Number = 987
	hash2 := HashStruct(hashstruct)

	assert.NotEqual(t, hash1, hash2, "Output of HashStruct is equal with non-equal input.")
}

func TestFingerprint(t *testing.T){
	pubkey, err := LoadPublicKey(clientCertPathTest+".pub")
	if err != nil {
		t.Fatal("Could not load public key. Error: " + err.Error())
	}
	fp, err := GetFingerprint(pubkey)
	if err != nil {
		t.Fatal("Could not get fingerprint. Error: " + err.Error())
	}

	assert.Equal(t, "95c73e8028118d18a961dd1da6b5e7c3", fp, "Fingerprint of Public key not correct")

	privkey, err := LoadPrivateKey(clientCertPathTest+".pem")
	if err != nil {
		t.Fatal("Could not load private key. Error: " + err.Error())
	}
	fp2, err := GetFingerprint(privkey)
	if err != nil {
		t.Fatal("Could not get fingerprint. Error: " + err.Error())
	}

	assert.Equal(t, fp2, fp, "Fingerprint of Public key and Private key not equal")
}
