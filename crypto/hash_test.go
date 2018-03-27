package crypto

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestHash(t *testing.T){
	fileHash, err := IPFSHashFile("hash.go")
	if err != nil {
		t.Fatal("IPFSHash failed: " + err.Error())
	}
	assert.Equal(t, "QmagcdMKqqbvR9gUH9SFQNXuxSo7ErB2TbJgQfY5rxtFXG", fileHash, "Hash did not match")

	textHash, err := IPFSHashData([]byte("racin"))
	assert.Equal(t, "QmagcdMKqqbvR9gUH9SFQNXuxSo7ErB2TbJgQfY5rxtFXG", textHash, "Hash did not match")

}