package crypto

import (
	"crypto/rand"
	"math/big"
	"encoding/binary"
	"math"
)

const numSamples = 1000;

// By using uint64 as the index it is possible to index files up to 2048 Exa bytes.
type StorageSample struct {
	Samples					map[uint64]byte		`json:"sample"`
}

type StorageChallenge struct {
	//Challengesignature		[]byte				`json:"challengesignature"`
	Challenge				[]byte				`json:"challenge"`
}

type StorageChallengeProof struct {
	SignedStruct // Of type StorageChallenge
	Proof					[]byte				`json:"proof"`
	//Proofsignature			[]byte				`json:"proofsignature"`
}

func GenerateStorageSample(fileByte *[]byte) *StorageSample{
	ret := &StorageSample{Samples:make(map[uint64]byte)}

	a, b := rand.Int(rand.Reader, big.NewInt(int64(len(*fileByte))))
	for i := 0; i < numSamples; i++ {
		buf := make([]byte, 8)
		_, err := rand.Read(buf)

		if err != nil {
			return nil // Problems generating a random number.
		}

		smplByte := binary.LittleEndian.Uint64(buf)

		ret.Samples[smplByte] =
	}



	z.SetUint64(64)
	num := rand.Int(rand.Reader, big.NewInt(math.Exp2(64)))

	return &StorageSample{}
}

func (sp *StorageSample) StoreSample() error{
	// Distribute the sample to the other consensus nodes. (Remember that different layers can not act maliciously
	// by colluding).
	return nil
}

func (sp *StorageSample) GenerateChallenge(fileByte *[]byte) *StorageChallenge{
	// Sign the challenge with our private key
	return &StorageChallenge{}
}

func (scp *StorageChallengeProof) VerifyChallengeProof() error{
	// Verify signatures on both the challenge and the proof.
	return nil
}