package crypto

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"fmt"
	"crypto/md5"
	"io/ioutil"
)
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
	assert.Equal(t, "QmWimMR7o684LGMTHRrCptBDbK3sP3m2ocYwmUUzYVKwfm", fileHash, "Hash output did not match.")

	textHash, err := HashData([]byte("racin"))
	if err != nil {
		t.Fatal("Hash failed: " + err.Error())
	}
	assert.Equal(t, "QmP4eE9BBqDRHrPbwFN75M9cX84Rm3G8B2fKtxZCtREUyC", textHash, "Hash output did not match.")
}

func TestHash2(t *testing.T){
	pubkey, _ := LoadPublicKey(certPathTest+".pub")
	fmt.Printf("%+v\n", pubkey.public)
	hd, _ := HashData([]byte(fmt.Sprintf("%v", pubkey.public)))
	fmt.Printf("%+v\n", hd)
	hd2, _ := HashFile(certPathTest+".pub")
	fmt.Printf("%+v\n", hd2)

	data, _ := ioutil.ReadFile(certPathTest+".pub")
	fmt.Printf("%x", md5.Sum(data))
}