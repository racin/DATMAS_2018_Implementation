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
	Identity				string					`json:"identity"`
	Cid						string					`json:"cid"`
	Samples					map[uint64]byte			`json:"sample"`
}

type StorageChallenge struct {
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
}

func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

func GenerateStorageSample(fileBytes *[]byte) *StorageSample{
	if fileBytes == nil {
		return nil
	}
	cid, err := IPFSHashData(*fileBytes)
	if err != nil {
		return nil
	}
	// If the file is smaller than numSamples, we simply store the whole file.
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

func (sp *StorageSample) SignSample(privKey *Keys) (*SignedStruct, error) {
	fp, err := GetFingerprint(privKey)
	if err != nil {
		return nil, err
	}
	(*sp).Identity = fp
	return SignStruct(sp, privKey)
}

func (sp *SignedStruct) StoreSample(basepath string) error{
	bytearr, err := json.Marshal(sp)
	if err != nil {
		return err
	}

	storageSample, ok := sp.Base.(*StorageSample)
	if !ok {
		return errors.New("SignedStruct must have StorageSample as underlying type.")
	}

	return ioutil.WriteFile(basepath + storageSample.Cid, bytearr, 0600)
	// Distribute the sample to the other consensus nodes. (Remember that different layers can not act maliciously
	// by colluding).
}

// We store the signature of the sample so that the authenticity of the sample can be proven later. (non-repudiation).
func LoadStorageSample(basepath string, cid string) *StorageSample{
	if bytearr, err := ioutil.ReadFile(basepath + cid); err == nil {
		signedStruct := &SignedStruct{Base: &StorageSample{}}
		if json.Unmarshal(bytearr,signedStruct) == nil {
			if ret, ok := signedStruct.Base.(*StorageSample); ok {
				return ret
			}
		}
	}

	return nil
}

func (sp *StorageSample) getSampleIndices() []uint64 {
	ret := make([]uint64, len(sp.Samples))
	i := 0
	for key, _:= range sp.Samples {
		ret[i] = key
		i++
	}
	return ret
}

func (sp *StorageSample) GenerateChallenge(privkey *Keys, cid string) *SignedStruct{
	chal := &StorageChallenge{Challenge: make([]uint64, challengeSamples), Cid: cid}
	sampleIndices := sp.getSampleIndices()
	max := new(big.Int).SetUint64(uint64(len(sampleIndices)))

	for i := 0; i < challengeSamples; i++ {
		if rnd, err := rand.Int(rand.Reader, max); err == nil {
			chal.Challenge[i] = sampleIndices[rnd.Uint64()]
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

func (signedStruct *SignedStruct) VerifySample(acl *conf.AccessList) error {
	/*challenge, ok := signedStruct.Base.(*StorageSample)
	if !ok {
		return errors.New("Could not type assert the StorageChallengeProof.")
	}

	challengerPubkey, err := LoadPublicKey(samplerIdent.PublicKey);
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
*/
	return nil
}

func (signedStruct *SignedStruct) VerifyChallenge(acl *conf.AccessList) error {
	challenge, ok := signedStruct.Base.(*StorageChallenge)
	if !ok {
		return errors.New("Could not type assert the StorageChallengeProof.")
	}

	identity, ok := acl.Identities[challenge.Identity]
	if !ok {
		return errors.New("Could not find identity attached to StorageChallenge.")
	}

	challengerPubkey, err := LoadPublicKey(identity.PublicKey);
	if err != nil {
		return err
	}

	if (identity.AccessLevel != conf.Consensus && identity.AccessLevel != conf.User) {
		return errors.New("Challengers identity was unexpected.")
	}

	// Check if the proof is signed by the expected prover.
	if !signedStruct.Verify(challengerPubkey) {
		return errors.New("Could not verify signature of challenger.")
	}

	return nil
}

// Challenger can simply keep a list of challenges sent, and to whom.
func (signedStruct *SignedStruct) VerifyChallengeProof(sampleBase string, acl *conf.AccessList, requiredProverIdent string) error{
	scp, ok := signedStruct.Base.(*StorageChallengeProof)
	if !ok {
		return errors.New("Could not type assert the StorageChallengeProof.")
	}

	proverIdent, ok := acl.Identities[requiredProverIdent]
	if !ok {
		return errors.New("Could not find proverIdent attached to StorageChallengeProof.")
	}

	if proverIdent.AccessLevel != conf.Storage || requiredProverIdent != scp.Identity {
		return errors.New("Provers identity was unexpected.")
	}

	challenge, ok := scp.Base.(*StorageChallenge)
	if !ok {
		return errors.New("Could not type assert the StorageChallenge.")
	}

	challengerIdent, ok := acl.Identities[challenge.Identity]
	if !ok {
		return errors.New("Could not find challengerIdent attached to StorageChallenge.")
	}

	if (challengerIdent.AccessLevel != conf.Consensus && challengerIdent.AccessLevel != conf.User) {
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