package crypto

import (
	"crypto/rsa"
	"crypto/rand"
	mh "github.com/multiformats/go-multihash"
	"crypto"
)

type Keys struct {
	private *rsa.PrivateKey
	public  *rsa.PublicKey
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