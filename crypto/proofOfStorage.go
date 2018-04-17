package crypto

import (
	"crypto/rand"
	"math/big"
	"encoding/json"
	"io/ioutil"
	"github.com/pkg/errors"
	conf "github.com/racin/DATMAS_2018_Implementation/configuration"
	"fmt"
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
	Cid						string					`json:"cid"`
}

type StorageChallengeProof struct {
	SignedStruct // Of type StorageChallenge
	Proof					map[uint64]byte				`json:"proof"`
	Identity				string						`json:"identity"`
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
	nSamples := min(numSamples, len(*fileBytes))
	ret := &StorageSample{Samples:make(map[uint64]byte, nSamples)}
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

func (sp *StorageSample) StoreSample(basepath string, cid string) error{
	bytearr, err := json.Marshal(sp)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(basepath + cid, bytearr, 0600)
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
	fmt.Printf("%+v\n", sp)
	chal := &StorageChallenge{Challenge: make([]uint64, challengeSamples), Cid: cid}
	max := new(big.Int).SetUint64(uint64(len((*sp).Samples)))
	for i := 0; i < challengeSamples; i++ {
		rnd, err := rand.Int(rand.Reader, max)

		if err != nil {
			return nil // Problems generating a random number.
		}

		chal.Challenge = append(chal.Challenge, rnd.Uint64())
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

	if (challengerIdent.AccessLevel != conf.Consensus && challengerIdent.AccessLevel != conf.User) ||
		challengerIdent.PublicKey != challenge.Identity {
		return errors.New("Challengers identity was unexpected.")
	}

	challengerPubkey, err := LoadPublicKey(challengerIdent.PublicKey);
	if err != nil {
		return err
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

	proof := &StorageChallengeProof{SignedStruct: *signedStruct, Identity: fingerprint, Proof: make(map[uint64]byte)}
	for _, value := range challenge.Challenge {
		proof.Proof[value] = (*fileBytes)[value]
	}

	newSignedStruct, err := SignStruct(proof, privKey)
	if err != nil {
		return nil;
	}

	return newSignedStruct
}

func (signedStruct *SignedStruct) VerifyChallengeProof(sampleBase string, proverIdent *conf.Identity, challengerIdent *conf.Identity) error{
	scp, ok := signedStruct.Base.(*StorageChallengeProof)
	if !ok {
		return errors.New("Could not type assert the StorageChallengeProof.")
	}

	challenge, ok := scp.Base.(*StorageChallenge)
	if !ok {
		return errors.New("Could not type assert the StorageChallenge.")
	}

	if proverIdent.AccessLevel != conf.Storage || proverIdent.PublicKey != scp.Identity {
		return errors.New("Provers identity was unexpected.")
	}
	if (challengerIdent.AccessLevel != conf.Consensus && challengerIdent.AccessLevel != conf.User) ||
		challengerIdent.PublicKey != challenge.Identity {
		return errors.New("Challengers identity was unexpected.")
	}

	// Check if public key exists and if message is signed.
	proverPubkey, err := LoadPublicKey(proverIdent.PublicKey);
	if err != nil {
		return err
	}

	challengerPubkey, err := LoadPublicKey(challengerIdent.PublicKey);
	if err != nil {
		return err
	}

	// Check if the proof is signed by the expected prover.
	if !signedStruct.Verify(proverPubkey) {
		return errors.New("Could not verify signature of prover.")
	}

	// Check if the challenge is signed by the expected challenger
	if  !scp.Verify(challengerPubkey){
		return errors.New("Could not verify signature of challenge.")
	}

	sample := LoadStorageSample(sampleBase, challenge.Cid)
	if sample == nil || sample.Samples == nil {
		return errors.New("Could not find a stored sample for this Cid.")
	}

	for _, value := range challenge.Challenge {
		if val, ok := scp.Proof[value]; !ok {
			return errors.New("Proof is missing for challenge byte: " + string(value))
		} else if val != sample.Samples[value] {
			return errors.New("Incorrect value on proof for challenge byte: " + string(value))
		}
	}

	return nil
}