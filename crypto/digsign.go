package crypto

import (
	"crypto/rsa"
	"crypto/rand"
	mh "github.com/multiformats/go-multihash"
	"crypto"
	"fmt"
)

type Keys struct {
	private *rsa.PrivateKey
	public  *rsa.PublicKey
}

func (k *Keys) Sign(dh interface {}) ([]byte){
	fmt.Printf("Privkey: %+v\n", k.private)
	fmt.Printf("Formatted: %v\n", convertDataHash(dh))
	sign, err := rsa.SignPSS(rand.Reader, k.private, crypto.SHA256,
		convertDataHash(dh), &rsa.PSSOptions{SaltLength:rsa.PSSSaltLengthAuto})

	if err != nil {
		fmt.Println(err)
	}
	return sign
}

func (k *Keys) Verify(dh interface{}, signature []byte) (bool){
	fmt.Printf("Pubkey: %+v\n", k.public)
	fmt.Printf("Signature: %+v\n", signature)
	if err := rsa.VerifyPSS(k.public, crypto.SHA256, convertDataHash(dh), signature,
		&rsa.PSSOptions{SaltLength:rsa.PSSSaltLengthAuto}); err != nil {
			fmt.Println(err.Error())
			return false;
	}
	return true
}

func convertDataHash(dataHash interface {}) []byte {
	switch val := dataHash.(type) {
	case string:
		fmt.Println("racin")
		// First check if its a Base58 encoded multihash string
		_, yerr := mh.FromB58String(val);
		fmt.Println(yerr)
		if x, err := mh.FromB58String(val); err == nil {
			// Two first bytes represent <hash function code> & <digest size>.
			// The remaining 32 is the actual hash output.
			fmt.Println(val)
			fmt.Println(x[2:])
			return x[2:]
		} else {
			return []byte(val)
		}
	case []byte:
		fmt.Println("wilhelm")
		return val;
	default:
		return nil
	}
}