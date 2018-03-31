package crypto

import (
	"crypto/rsa"
	"crypto/rand"
	mh "github.com/multiformats/go-multihash"
	"crypto"
	"encoding/pem"
	"crypto/x509"
	"fmt"
	"crypto/dsa"
	"crypto/ecdsa"
	"io/ioutil"
)

type Keys struct {
	private *rsa.PrivateKey
	public  *rsa.PublicKey
}

func convertDataHash(dataHash interface {}) []byte {
	switch val := dataHash.(type) {
	case string:
		// First check if its a Base58 encoded multihash string
		if x, err := mh.FromB58String(val); err == nil {
			// Two first bytes represent <hash function code> & <digest size>.
			// The remaining 32 is the actual hash output.
			return x[2:]
		} else {
			return []byte(val)
		}
	case []byte:
		return val;
	default:
		return nil
	}
}

func (k *Keys) Sign(dh interface {}) ([]byte, error){
	return rsa.SignPSS(rand.Reader, k.private, crypto.SHA256,
		convertDataHash(dh), &rsa.PSSOptions{SaltLength:rsa.PSSSaltLengthAuto})
}

func (k *Keys) Verify(dh interface{}, signature []byte) (bool){
	if err := rsa.VerifyPSS(k.public, crypto.SHA256, convertDataHash(dh), signature,
		&rsa.PSSOptions{SaltLength:rsa.PSSSaltLengthAuto}); err != nil {
			return false;
	}
	return true
}

func LoadPublicKey(path string) (*Keys, error){
	dat, _ := ioutil.ReadFile("test_certificate/mycert_test.pub")
	fmt.Println(dat)
	//block, _ := pem.Decode([]byte(pubPEM))
	block, _ := pem.Decode(dat)
	if block == nil {
		panic("failed to parse PEM block containing the public key")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		panic("failed to parse DER encoded public key: " + err.Error())
	}

	switch pub := pub.(type) {
	case *rsa.PublicKey:
		fmt.Println("pub is of type RSA:", pub)
	case *dsa.PublicKey:
		fmt.Println("pub is of type DSA:", pub)
	case *ecdsa.PublicKey:
		fmt.Println("pub is of type ECDSA:", pub)
	default:
		panic("unknown type of public key")
	}
	return &Keys{}
}
func LoadPrivateKey(path string) (*Keys, error){
	dat, _ := ioutil.ReadFile("test_certificate/mycert_test.pub")
	fmt.Println(dat)
	//block, _ := pem.Decode([]byte(pubPEM))
	block, _ := pem.Decode(dat)
	if block == nil {
		panic("failed to parse PEM block containing the public key")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		panic("failed to parse DER encoded public key: " + err.Error())
	}

	switch pub := pub.(type) {
	case *rsa.PublicKey:
		fmt.Println("pub is of type RSA:", pub)
	case *dsa.PublicKey:
		fmt.Println("pub is of type DSA:", pub)
	case *ecdsa.PublicKey:
		fmt.Println("pub is of type ECDSA:", pub)
	default:
		panic("unknown type of public key")
	}
	return &Keys{}
}



