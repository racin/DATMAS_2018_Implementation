package crypto

import (
	"crypto/rand"
	"math/big"
	"encoding/binary"
	"math"
	"encoding/json"
	"io/ioutil"
	"github.com/pkg/errors"
)

const (
	numSamples = 10000; // Each sample will require approximately 10KB of storage.
	challengeSamples = 10; // Probability of guessing a correct proof is about: 1 / (2^(8*10)
)

// By using uint64 as the index it is possible to index files up to 2048 Exa bytes.
type StorageSample struct {
	Samples					map[uint64]byte		`json:"sample"`
}

type StorageChallenge struct {
	//Challengesignature		[]byte				`json:"challengesignature"`
	Challenge				[]uint64				`json:"challenge"`
	Identity				string					`json:"identity"`
}

type StorageChallengeProof struct {
	SignedStruct // Of type StorageChallenge
	Proof					map[uint64]byte				`json:"proof"`
	Identity				string						`json:"identity"`
	//Proofsignature			[]byte				`json:"proofsignature"`
}

func GenerateStorageSample(fileByte *[]byte) *StorageSample{
	ret := &StorageSample{Samples:make(map[uint64]byte, numSamples)}
	max := new(big.Int).SetUint64(uint64(len(*fileByte)))

	for i := 0; i < numSamples; i++ {
		rnd, err := rand.Int(rand.Reader, max)

		if err != nil {
			return nil // Problems generating a random number.
		}


		rnduint := rnd.Uint64()

		if _, ok := ret.Samples[rnduint]; ok {
			continue // This byte is already sampled.
		}

		ret.Samples[rnduint] =	(*fileByte)[rnduint]
	}

	return ret
}

func (sp *StorageSample) StoreSample(basepath string, cid string) error{
	bytearr, err := json.Marshal(sp)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(basepath + cid, bytearr, 0600)
	// Distribute the sample to the other consensus nodes. (Remember that different layers can not act maliciously
	// by colluding).
}

func (sp *StorageSample) GenerateChallenge(keys *Keys) *SignedStruct{
	chal := &StorageChallenge{Challenge: make([]uint64, challengeSamples)}
	max := new(big.Int).SetUint64(uint64(len((*sp).Samples)))
	for i := 0; i < challengeSamples; i++ {
		rnd, err := rand.Int(rand.Reader, max)

		if err != nil {
			return nil // Problems generating a random number.
		}

		chal.Challenge = append(chal.Challenge, rnd.Uint64())
	}

	ident, err := GetFingerPrint(keys)
	if err != nil {
		return nil // Could not get the keys fingerprint.
	}
	chal.Identity = ident

	// Sign the challenge with our private key
	signed, err := SignStruct(chal, keys)
	if err != nil {
		// Problem signing the data.
		return nil
	}
	return signed
}

func (scp *StorageChallengeProof) VerifyChallengeProof(keys *Keys) error{
	
	challenge, ok := scp.Base.(*StorageChallenge)
	if !ok {
		return errors.New("Could not type assert the StorageChalleng.")
	}


	// Verify signatures on both the challenge and the proof.
	return nil
}