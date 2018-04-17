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
	"os"
	"path/filepath"
	//"github.com/racin/DATMAS_2018_Implementation/types"
	"reflect"
	"bytes"
)

type SignedStruct struct {
	Base		interface{}
	Signature 	[]byte          `json:"signature"`
}

func HashStruct(in interface{}) string {
	buffer := bytes.NewBuffer([]byte{38})
	buffer.WriteString(fmt.Sprintf("%v", reflect.ValueOf(in)))
	hash, _ := HashData(buffer.Bytes())
	return hash
}

func SignStruct(in interface{}, keys *Keys) (*SignedStruct, error) {
	if signature, err := keys.Sign(HashStruct(in)); err != nil {
		return nil, err
	} else {
		return &SignedStruct{Base: in, Signature: signature}, nil
	}
}

func (t *SignedStruct) Verify(keys *Keys) bool {
	return keys.Verify(HashStruct(t.Base), t.Signature)
}

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
	dat, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("Failed to read file. %s", err.Error())
	}

	block, _ := pem.Decode(dat)
	if block == nil {
		return nil, fmt.Errorf("Failed to parse PEM block. %s", err.Error())
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse public key: %s", err.Error())
	}

	if pk, ok := pub.(*rsa.PublicKey); ok {
		return &Keys{public: pk}, nil
	}

	return nil, fmt.Errorf("Could not unmarshal public key.")
}

func LoadPrivateKey(path string) (*Keys, error) {
	dat, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("Failed to read file. %s", err.Error())
	}

	block, _ := pem.Decode(dat)
	if block == nil {
		return nil, fmt.Errorf("Failed to parse PEM block. %s", err.Error())
	}

	priv, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse private key: %s", err.Error())
	}

	if pk, ok := priv.(*rsa.PrivateKey); ok {
		return &Keys{private: pk, public: &pk.PublicKey}, nil
	}

	return nil, fmt.Errorf("Could not unmarshal private key.")
}

func GenerateKeyPair(path, name string, bits int) (*Keys, error){
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return nil, err
	}

	keypair, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, err;
	}

	privatekey_marshal, err := x509.MarshalPKCS8PrivateKey(keypair)
	if err != nil {
		return nil, err
	}

	privatekeypair_pem := pem.EncodeToMemory(
		&pem.Block{
			Type:  "PRIVATE KEY",
			Bytes: privatekey_marshal,
		},
	)

	if err = ioutil.WriteFile(path+"/"+name+".pem", privatekeypair_pem, 0600); err != nil {
		return nil, err
	}

	publickey_marshal, err := x509.MarshalPKIXPublicKey(&keypair.PublicKey)
	if err != nil {
		return nil, err
	}

	publickeypair_pem := pem.EncodeToMemory(
		&pem.Block{
			Type:  "PUBLIC KEY",
			Bytes: publickey_marshal,
		},
	)

	if err = ioutil.WriteFile(path+"/"+name+".pub", publickeypair_pem, 0600); err != nil {
		return nil, err
	}

	return &Keys{private:keypair, public:&keypair.PublicKey}, nil
}