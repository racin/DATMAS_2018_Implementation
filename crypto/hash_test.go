package crypto

import (
	"testing"
	"github.com/stretchr/testify/assert"
)
func TestHash(t *testing.T){
	fileHash, err := IPFSHashFile("hash_test.txt")
	if err != nil {
		t.Fatal("IPFSHash failed: " + err.Error())
	}
	assert.Equal(t, "QmRFq5YyyNai59Pvxfd5pGJY6HzpubyzpDp6ceqbJfDUBp", fileHash, "Hash did not match")

	textHash, err := IPFSHashData([]byte("racin\n"))
	if err != nil {
		t.Fatal("IPFSHash failed: " + err.Error())
	}
	assert.Equal(t, "QmdeVbypiSbW24Uhjdvdhczpv1gxDXA9nPYKGiPaAnQs5F", textHash, "Hash did not match")
}