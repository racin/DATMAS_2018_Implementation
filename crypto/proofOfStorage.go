package crypto

import (
	"crypto/rand"
	"math/big"
	"encoding/binary"
	"math"
	"encoding/json"
	"io/ioutil"
)

const numSamples = 1000;

// By using uint64 as the index it is possible to index files up to 2048 Exa bytes.
type StorageSample struct {
	Samples					map[uint64]byte		`json:"sample"`
}

type StorageChallenge struct {
	//Challengesignature		[]byte				`json:"challengesignature"`
	Challenge				[]uint64				`json:"challenge"`
}

type StorageChallengeProof struct {
	SignedStruct // Of type StorageChallenge
	Proof					map[uint64]byte				`json:"proof"`
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

func (sp *StorageSample) GenerateChallenge() *StorageChallenge{

	// Sign the challenge with our private key
	return &StorageChallenge{}
}

func (scp *StorageChallengeProof) VerifyChallengeProof() error{
	// Verify signatures on both the challenge and the proof.
	return nil
}