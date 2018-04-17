package crypto

import (
	"crypto/rand"
	"math/big"
	"encoding/json"
	"io/ioutil"
	"github.com/pkg/errors"
	conf "github.com/racin/DATMAS_2018_Implementation/configuration"
)

const (
	numSamples = 3000; // Each sample will require approximately 30KB of storage. (Could probably be reduced with another datastructure.)
	challengeSamples = 10; // Probability of guessing a correct proof is about: 1 / (2^(8*10)
)

// By using uint64 as the index it is possible to index files up to 2048 Exa bytes.
type StorageSample struct {
	Cid						string					`json:"cid"`
	Samples					map[uint64]byte		`json:"sample"`
}

type StorageChallenge struct {
	//Challengesignature		[]byte				`json:"challengesignature"`
	Challenge				[]uint64				`json:"challenge"`
	Identity				string					`json:"identity"`
	Cid						string					`json:"cid"`
}

// We can not use map[uint64]byte as the type for the proof, as when iterating the map, the order is not guaranteed to
// be equal between iterations. Therefore the hash of the struct may differ, and signatures will fail to verify.
// Instead we rely on the guaranteed ordering of arrays and use the underlying Challenge []uint64 to specify the index
// of each byte in Proof.
type StorageChallengeProof struct {
	SignedStruct // Of type StorageChallenge
	Proof					[]byte					`json:"proof"`
	Identity				string					`json:"identity"`
	//Proofsignature			[]byte				`json:"proofsignature"`
}

func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

// If the file is smaller than numSamples, we simply store the whole file.
func GenerateStorageSample(fileBytes *[]byte) *StorageSample{
	if fileBytes == nil {
		return nil
	}
	cid, err := IPFSHashData(*fileBytes)
	if err != nil {
		return nil
	}
	nSamples := min(numSamples, len(*fileBytes))
	ret := &StorageSample{Cid: cid, Samples:make(map[uint64]byte, nSamples)}
	max := new(big.Int).SetUint64(uint64(len(*fileBytes)))

	for i := 0; i < nSamples; i++ {
		rnd, err := rand.Int(rand.Reader, max)

		if err != nil {
			return nil // Problems generating a random number.
		}


		rnduint := rnd.Uint64()

		if _, ok := ret.Samples[rnduint]; ok {
			i--
			continue // This byte is already sampled.
		}

		ret.Samples[rnduint] =	(*fileBytes)[rnduint]
	}

	return ret
}

func (sp *StorageSample) StoreSample(basepath string) error{
	bytearr, err := json.Marshal(sp)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(basepath + sp.Cid, bytearr, 0600)
	// Distribute the sample to the other consensus nodes. (Remember that different layers can not act maliciously
	// by colluding).
}

func LoadStorageSample(basepath string, cid string) *StorageSample{
	ret := &StorageSample{}
	bytearr, err := ioutil.ReadFile(basepath + cid)
	if err == nil {
		json.Unmarshal(bytearr,ret);
	}

	return ret
}

func (sp *StorageSample) GenerateChallenge(privkey *Keys, cid string) *SignedStruct{
	chal := &StorageChallenge{Challenge: make([]uint64, challengeSamples), Cid: cid}
	max := new(big.Int).SetUint64(uint64(len((*sp).Samples)))
	for i := 0; i < challengeSamples; i++ {
		rnd, err := rand.Int(rand.Reader, max)

		if err != nil {
			return nil // Problems generating a random number.
		}

		rndUint := rnd.Uint64()
		if _, ok := sp.Samples[rndUint]; ok {
			chal.Challenge[i] = rndUint
		} else {
			i--
			continue;
		}

	}

	ident, err := GetFingerprint(privkey)
	if err != nil {
		return nil // Could not get the keys fingerprint.
	}
	chal.Identity = ident

	// Sign the challenge with our private key
	signed, err := SignStruct(chal, privkey)
	if err != nil {
		// Problem signing the data.
		return nil
	}
	return signed
}

func (signedStruct *SignedStruct) VerifyChallenge(challengerIdent conf.Identity) error {
	challenge, ok := signedStruct.Base.(*StorageChallenge)
	if !ok {
		return errors.New("Could not type assert the StorageChallengeProof.")
	}


	challengerPubkey, err := LoadPublicKey(challengerIdent.PublicKey);
	if err != nil {
		return err
	}

	fp, err := GetFingerprint(challengerPubkey)
	if err != nil {
		return err;
	}

	if (challengerIdent.AccessLevel != conf.Consensus && challengerIdent.AccessLevel != conf.User) ||
		fp != challenge.Identity {
		return errors.New("Challengers identity was unexpected.")
	}

	// Check if the proof is signed by the expected prover.
	if !signedStruct.Verify(challengerPubkey) {
		return errors.New("Could not verify signature of challenger.")
	}

	return nil
}

func (signedStruct *SignedStruct) ProveChallenge(privKey *Keys, fileBytes *[]byte) *SignedStruct {
	if fileBytes == nil {
		return nil
	}

	challenge, ok := signedStruct.Base.(*StorageChallenge)
	if !ok {
		return nil
	}

	fingerprint, err := GetFingerprint(privKey)
	if err != nil {
		return nil
	}

	lenChal := len(challenge.Challenge)
	proof := &StorageChallengeProof{SignedStruct: *signedStruct, Identity: fingerprint, Proof: make([]byte,lenChal)}
	for i := 0; i < lenChal; i++ {
		proof.Proof[i] = (*fileBytes)[challenge.Challenge[i]]
	}

	newSignedStruct, err := SignStruct(proof, privKey)

	if err != nil {
		return nil;
	}

	return newSignedStruct
}

func (signedStruct *SignedStruct) VerifyChallengeProof(sampleBase string, challengerIdent *conf.Identity, proverIdent *conf.Identity) error{
	scp, ok := signedStruct.Base.(*StorageChallengeProof)
	if !ok {
		return errors.New("Could not type assert the StorageChallengeProof.")
	}

	challenge, ok := scp.Base.(*StorageChallenge)
	if !ok {
		return errors.New("Could not type assert the StorageChallenge.")
	}

	// Check if public key exists and if message is signed.
	proverPubkey, err := LoadPublicKey(proverIdent.PublicKey);
	if err != nil {
		return err
	}

	prover_fp, err := GetFingerprint(proverPubkey)
	if err != nil {
		return nil
	}

	challengerPubkey, err := LoadPublicKey(challengerIdent.PublicKey);
	if err != nil {
		return err
	}

	challenger_fp, err := GetFingerprint(challengerPubkey)
	if err != nil {
		return nil
	}

	if proverIdent.AccessLevel != conf.Storage || prover_fp != scp.Identity {
		return errors.New("Provers identity was unexpected.")
	}
	if (challengerIdent.AccessLevel != conf.Consensus && challengerIdent.AccessLevel != conf.User) ||
		challenger_fp != challenge.Identity {
		return errors.New("Challengers identity was unexpected.")
	}

	// Check if the proof is signed by the expected prover.
	if !signedStruct.Verify(proverPubkey) {
		return errors.New("Could not verify signature of prover.")
	}

	// Check if the challenge is signed by the expected challenger
	if  !scp.Verify(challengerPubkey){
		return errors.New("Could not verify signature of challenger.")
	}

	sample := LoadStorageSample(sampleBase, challenge.Cid)
	if sample == nil || sample.Samples == nil {
		return errors.New("Could not find a stored sample for this Cid.")
	}

	lenChal := len(challenge.Challenge)
	lenProof := len(scp.Proof)

	if lenChal != lenProof {
		return errors.New("Length of challenge and proof is not equal.")
	}
	for i := 0; i < lenChal; i++ {
		index := challenge.Challenge[i]
		if scp.Proof[i] != sample.Samples[index] {
			return errors.Errorf("Incorrect value on proof for challenge byte: %v", index)
		}
	}

	return nil
}