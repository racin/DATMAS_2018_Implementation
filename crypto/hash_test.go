package crypto

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"fmt"
	"crypto/md5"
	"io/ioutil"
	"strings"
	"encoding/base64"
	/*"github.com/yosida95/golang-sshkey"
	"crypto"
	"crypto/rsa"*/
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
	assert.Equal(t, "QmWimMR7o684LGMTHRrCptBDbK3sP3m2ocYwmUUzYVKwfm", fileHash, "File hash did not match")

	textHash, err := HashData([]byte("racin"))
	if err != nil {
		t.Fatal("Hash failed: " + err.Error())
	}
	assert.Equal(t, "QmP4eE9BBqDRHrPbwFN75M9cX84Rm3G8B2fKtxZCtREUyC", textHash, "Data hash did not match")
}

func TestHash2(t *testing.T){
	pubkey, _ := LoadPublicKey(certPathTest+".pub")
	fmt.Printf("%+v\n", pubkey.public)
	hd, _ := HashData([]byte(fmt.Sprintf("%v", pubkey.public)))
	fmt.Printf("%+v\n", hd)
	hd2, _ := HashFile(certPathTest+".pub")
	fmt.Printf("%+v\n", hd2)

	data, _ := ioutil.ReadFile(certPathTest+".pub")
	fmt.Printf("%x\n", md5.Sum(data))
	pubkey2, _ := LoadPublicKey(certPathTest+".pub")
	hd3 :=  md5.Sum([]byte(fmt.Sprintf("%v", pubkey2.public)))
	fmt.Printf("%x\n", hd3)

	key, _ := ioutil.ReadFile(certPathTest+".pub")


	parts := strings.Fields(string(key))


	k, _ := base64.StdEncoding.DecodeString(parts[1])


	fp := md5.Sum([]byte(k))
	fmt.Print("MD5:")
	for i, b := range fp {
		fmt.Printf("%02x", b)
		if i < len(fp)-1 {
			fmt.Print(":")
		}
	}
	fmt.Println()
/*
	privkey, _ := LoadPrivateKey(certPathTest+".pem")
	a, _ := sshkey.Fingerprint(privkey.private.Public().(sshkey.PublicKey), crypto.MD5)
	fmt.Printf("%x\n", a)*/

	fp2, _ := GetFingerPrint(pubkey)
	fmt.Printf("%s\n", fp2)

	fps := GetFingerPrintShort(pubkey)
	fmt.Printf("%s\n", fps)
}
