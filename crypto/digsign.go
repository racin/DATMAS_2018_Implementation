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
	conf "github.com/racin/DATMAS_2018_Implementation/configuration"
	"sort"
)

type SignedStruct struct {
	Base		interface{}
	Signature 	[]byte          `json:"signature"`
}

// Sorts an array of Map keys in alphabetical order. Required because the order of Map is not guaranteed, and thus
// the output hash is not deterministic and can cause validation of signatures to fail.
func sortMapKeys(keys []reflect.Value) []reflect.Value{
	sortedKeys := make([]reflect.Value,len(keys))
	strKeys := make([]string, len(keys))
	mapKeys := make(map[string]reflect.Value, len(keys))
	for i, key := range keys {
		strKeys[i] = key.String()
		mapKeys[key.String()] = key
	}
	sort.Strings(strKeys)
	for i, key := range strKeys {
		sortedKeys[i] = mapKeys[key]
	}

	return sortedKeys
}

// Handles different kinds of reflect.Value inputs. Not that this function is by no means complete for all possible types,
// and edge cases, just for the ones used in this project.
func internal_handleValue(val reflect.Value, buffer *bytes.Buffer){
	if val.Kind() == reflect.Struct{
		buffer.Write(internal_hashStruct(val))
	} else if val.Kind() == reflect.Ptr {
	} else if val.Kind() == reflect.Interface {
		inf := val.Interface()
		if infVal, ok := inf.(string); ok {
			buffer.WriteString(infVal)
		} else if infVal, ok := inf.(int); ok {
			buffer.WriteByte(byte(infVal))
		} else if infVal, ok := inf.(int64); ok {
			buffer.WriteByte(byte(infVal))
		} else if infVal, ok := inf.(float64); ok {
			buffer.WriteByte(byte(infVal))
		} else {
			buffer.Write(internal_hashStruct(inf))
		}
	} else {
		buffer.WriteString(fmt.Sprintf("%v", val))
	}
}

// Due to the design choice of using a lot of generic/interface types, we need to have logic to generate a deterministic
// hash value of any type. If the input is a struct which has a field which also is a struct, this function will be called
// recursively until it finds a basic type.
func internal_hashStruct(in interface{}) []byte {
	var buffer *bytes.Buffer
	var v reflect.Value
	var ok bool
	if v, ok = in.(reflect.Value); !ok {
		// Start the buffer with an ampersand. Similar to how fmt will print a struct by default.
		buffer = bytes.NewBuffer([]byte{38})
		v = reflect.ValueOf(in);
		if v.Kind() == reflect.Invalid {
			return buffer.Bytes()
		} else if v.Kind() == reflect.Ptr {
			v = v.Elem()
			if v.Kind() == reflect.Invalid {
				return buffer.Bytes()
			}
		}

		if v.Kind() == reflect.Map {
			for _, key := range sortMapKeys(v.MapKeys()) {
				internal_handleValue(v.MapIndex(key), buffer)
			}
			return buffer.Bytes()
		} else if v.Kind() == reflect.Slice {
			for i := 0; i < v.Len(); i++ {
				val := v.Index(i)
				if val.Kind() == reflect.Map {
					for _, key := range sortMapKeys(val.MapKeys()) {
						internal_handleValue(val.MapIndex(key), buffer)
					}
				} else {
					internal_handleValue(v.Index(i), buffer)
				}
			}

			return buffer.Bytes()
		}
	} else {
		buffer = bytes.NewBuffer([]byte{})
	}
	for i := 0; i < v.NumField(); i++ {
		internal_handleValue(v.Field(i), buffer)
	}

	return buffer.Bytes()
}

// Generate a deterministic hash of any type. Strictly required with dealing with digital signatures on various types
// of data.
func HashStruct(in interface{}) string {
	bytes := internal_hashStruct(in)
	hash, _ := HashData(bytes)

	return hash
}

// Digitally signs any type of input.
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
	if k.public == nil {
		return false
	}
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

func GetIdentityPublicKey(ident string, acl *conf.AccessList, pubkeyBase string) (identity *conf.Identity, pubkey *Keys){
	if id, ok := acl.Identities[ident]; ok {
		identity = id
		if pk, err := LoadPublicKey(pubkeyBase + identity.PublicKey); err == nil {
			pubkey = pk
		}
	}
	return
}