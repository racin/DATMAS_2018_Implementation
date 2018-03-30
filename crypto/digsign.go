package crypto

import (
	"crypto/rsa"
	"crypto/rand"
	"crypto"
)

type Keys struct {
	private *rsa.PrivateKey
	public  *rsa.PublicKey
}

func (priv *rsa.PrivateKey) Sign(dataHash []byte) ([]byte){
	sign, _ := rsa.SignPSS(rand.Reader, k, crypto.SHA256,
		dataHash, &rsa.PSSOptions{SaltLength:rsa.PSSSaltLengthAuto})

	return sign
}

func (k *PrivateKey) Verify(dataHash []byte, signature [] byte) (bool){
	if err := rsa.VerifyPSS(k.public, crypto.SHA256, dataHash, signature,
		&rsa.PSSOptions{SaltLength:rsa.PSSSaltLengthAuto}); err != nil {
		return false;
	}
	return true
}