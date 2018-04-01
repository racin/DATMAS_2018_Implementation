package crypto

import (
	"crypto/rsa"
	"crypto/rand"
	mh "github.com/multiformats/go-multihash"
	"crypto"
	"encoding/pem"
	"crypto/x509"
	"fmt"
	"io/ioutil"
)

type Keys struct {
	private *rsa.PrivateKey
	public  *rsa.PublicKey
}

func convertDataHash(dataHash interface{}) []byte {
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

func (k *Keys) Sign(dh interface{}) ([]byte, error) {
	return rsa.SignPSS(rand.Reader, k.private, crypto.SHA256,
		convertDataHash(dh), &rsa.PSSOptions{SaltLength: rsa.PSSSaltLengthAuto})
}

func (k *Keys) Verify(dh interface{}, signature []byte) (bool) {
	if err := rsa.VerifyPSS(k.public, crypto.SHA256, convertDataHash(dh), signature,
		&rsa.PSSOptions{SaltLength: rsa.PSSSaltLengthAuto}); err != nil {
		return false;
	}
	return true
}

func LoadPublicKey(path string) (*Keys, error) {
	dat, _ := ioutil.ReadFile(path)
	block, _ := pem.Decode(dat)
	if block == nil {
		return nil, fmt.Errorf("failed to parse PEM block containing the public key")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse DER encoded public key: %s", err.Error())
	}

	if pk, ok := pub.(*rsa.PublicKey); ok {
		return &Keys{public: pk}, nil
	}

	return nil, fmt.Errorf("Could not unmarshal public key.")
}
func LoadPrivateKey(path string) (*Keys, error) {
	dat, _ := ioutil.ReadFile(path)
	block, _ := pem.Decode(dat)
	if block == nil {
		return nil, fmt.Errorf("failed to parse PEM block containing the private key")
	}

	priv, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse DER encoded private key: %s", err.Error())
	}

	if pk, ok := priv.(*rsa.PrivateKey); ok {
		return &Keys{private: pk, public: &pk.PublicKey}, nil
	}

	return nil, fmt.Errorf("Could not unmarshal public key.")
}
