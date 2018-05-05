package crypto

import (
	"crypto/rand"
	"math/big"
	"encoding/json"
	"io/ioutil"
	"github.com/pkg/errors"
	conf "github.com/racin/DATMAS_2018_Implementation/configuration"
	"encoding/base64"
)

// A simple strawman implementation of the a proof of storage algorithm. Is reliant on storing the actual file bytes locally.
const (
	numSamples = 3000; // 16400; // Each sample will require approximately 150KB of storage. This can be drastically decreased by
	// using Homomorphic Verifiable Tags instead of the actual file bytes.
	challengeSamples = 10; // 460; // Having 460 challenge samples gives >99% probability that if 1% of the file is missing or corrupted it will be detected.
)

// By using uint64 as the index it is possible to index files up to 16 Exa bytes.
// Identity is the fingerprint of the one that sampled the data. Cid is the identifier for the file.
type StorageSample struct {
	Identity				string					`json:"identity"`
	Cid						string					`json:"cid"`
	Sampleindices			[]uint64				`json:"sampleIndices"`
	Samplevalues			[]byte					`json:"sampleValues"`
	Filesize				int64					`json:"filesize"`
}

type StorageChallenge struct {
	Challenge				[]uint64				`json:"challenge"`
	Identity				string					`json:"identity"`
	Cid						string					`json:"cid"`
	Nonce					float64					`json:"nonce"` // Must be float64 because of interface conversion and overflow issues. (Could possibly be uint64 if unsafe pointers were used.
}

// We can not use map[uint64]byte as the type to represent the samples, as when iterating the map, the order is not guaranteed to
// be equal between iterations. Therefore the hash of the struct may differ, and signatures will fail to verify.
// Instead we rely on the guaranteed ordering of arrays and use the underlying Challenge []uint64 to specify the index
// of each byte in Proof. We include filesize in order to detect a client and a single storage node colluding to report
// erroneous file size to the consensus node.
type StorageChallengeProof struct {
	SignedStruct // Of type StorageChallenge
	Proof					string					`json:"proof"`
	Identity				string					`json:"identity"`
	Filesize				int64					`json:"filesize"`
}
func GetStorageChallengeProofArray(derivedArray []interface{}) []SignedStruct {
	scpSlice := make([]SignedStruct, len(derivedArray))
	for i, val := range derivedArray {
		if mapInf, ok := val.(map[string]interface{}); ok {
			if signedChalProof := GetSignedStorageChallengeProofFromMap(mapInf); signedChalProof == nil {
				return nil
			} else {
				scpSlice[i] = *signedChalProof
			}
		}
	}
	return scpSlice
}
func GetSignedStorageChallengeProofFromMap(derivedStruct map[string]interface{}) *SignedStruct {
	if base, ok := derivedStruct["Base"]; ok {
		if signature, ok := derivedStruct["signature"]; ok {
			if data, err := base64.StdEncoding.DecodeString(signature.(string)); err == nil {
				ss := &SignedStruct{Base: GetStorageChallengeProofFromMap(base.(map[string]interface{})), Signature: data}
				return ss
			}
		}
	}
	return nil
}
func GetStorageChallengeProofFromMap(derivedStruct map[string]interface{}) *StorageChallengeProof {
	if proof, ok := derivedStruct["proof"]; ok {
		if identity, ok := derivedStruct["identity"]; ok {
			if filesize, ok := derivedStruct["filesize"]; ok {
				ss := &StorageChallengeProof{SignedStruct: *GetSignedStorageChallengeFromMap(derivedStruct),
					Proof: proof.(string), Identity: identity.(string), Filesize: int64(filesize.(float64))}
				return ss
			}
		}
	}
	return nil
}
func GetSignedStorageChallengeFromMap(derivedStruct map[string]interface{}) *SignedStruct {
	if base, ok := derivedStruct["Base"]; ok {
		if signature, ok := derivedStruct["signature"]; ok {
			if data, err := base64.StdEncoding.DecodeString(signature.(string)); err == nil {
				ss := &SignedStruct{Base: GetStorageChallengeFromMap(base.(map[string]interface{})), Signature: data}
				return ss
			}
		}
	}
	return nil
}
func GetStorageChallengeFromMap(derivedStruct map[string]interface{}) *StorageChallenge {
	if cid, ok := derivedStruct["cid"]; ok {
		if nonce, ok := derivedStruct["nonce"]; ok {
			if identity, ok := derivedStruct["identity"]; ok {
				if challenge, ok := derivedStruct["challenge"]; ok {
					if clg, ok := challenge.([]interface{}); ok {
						chal := make([]uint64, len(clg))
						for i, val := range clg {
							chal[i] = uint64(val.(float64))
						}

						return &StorageChallenge{Cid: cid.(string), Challenge: chal,
							Identity: identity.(string), Nonce:nonce.(float64)}
					}
				}
			}
		}
	}
	return nil
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
	lenFile := len(*fileBytes)
	nSamples := min(numSamples, lenFile)

	ret := &StorageSample{Cid: cid, Sampleindices: make([]uint64, nSamples),
		Samplevalues:make([]byte, nSamples), Filesize:int64(lenFile)}
	max := new(big.Int).SetUint64(uint64(lenFile))
	addedSamples := make(map[uint64]bool)
	for i := 0; i < nSamples; i++ {
		rnd, err := rand.Int(rand.Reader, max)

		if err != nil {
			return nil // Problems generating a random number.
		}

		rnduint := rnd.Uint64()
		if _, ok := addedSamples[rnduint]; ok {
			i--
			continue // This byte is already sampled.
		}
		ret.Sampleindices[i] = rnduint
		ret.Samplevalues[i] = (*fileBytes)[rnduint]
		addedSamples[rnduint] = true
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

func (sp *StorageSample) CompareTo(other *StorageSample) bool {
	if len(sp.Sampleindices) == len(other.Sampleindices) {
		mySampleMap := sp.getSamplesMap()
		otherSampleMap := other.getSamplesMap()
		for key, val := range *mySampleMap {
			if otherVal, ok := (*otherSampleMap)[key]; !ok || otherVal != val {
				return false
			}
		}
		return true
	}
	return false
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

func (sp *StorageSample) GenerateChallenge(privkey *Keys) (challenge *SignedStruct, challengeHash string, proof string){
	nonce, err := rand.Int(rand.Reader, new(big.Int).SetUint64(2 << 52)) // 9007199254740992
	if err != nil {
		return nil, "", "" // Could not generate nonce.
	}
	chal := &StorageChallenge{Challenge: make([]uint64, challengeSamples), Cid: sp.Cid, Nonce:float64(nonce.Int64())}
	max := new(big.Int).SetUint64(uint64(len(sp.Sampleindices)))
	proofBytes := make([]byte, challengeSamples)
	for i := 0; i < challengeSamples; i++ {
		rnd, err := rand.Int(rand.Reader, max);
		if err != nil {
			return nil, "", ""
		}
		chal.Challenge[i] = sp.Sampleindices[rnd.Uint64()]
		proofBytes[i] = sp.Samplevalues[rnd.Uint64()]
	}
	if proof, err = HashData(proofBytes); err != nil {
		return nil, "", ""
	}

	ident, err := GetFingerprint(privkey)
	if err != nil {
		return nil, "", "" // Could not get the keys fingerprint.
	}
	chal.Identity = ident

	// Sign the challenge with our private key
	challengeHash = HashStruct(chal)
	if signature, err := privkey.Sign(challengeHash); err != nil {
		return nil, "", ""
	} else {
		return &SignedStruct{Base: chal, Signature: signature}, challengeHash, proof
	}
}

func GenerateRandomChallenge(privkey *Keys, cid string, maxIndex int64) (challenge *SignedStruct, challengeHash string){
	nonce, err := rand.Int(rand.Reader, new(big.Int).SetUint64(2 << 52)) // 9007199254740992
	if err != nil {
		return nil, "" // Could not generate nonce.
	}
	chal := &StorageChallenge{Challenge: make([]uint64, challengeSamples), Cid: cid, Nonce:float64(nonce.Int64())}
	max := big.NewInt(maxIndex)

	for i := 0; i < challengeSamples; i++ {
		rndIndex, err := rand.Int(rand.Reader, max)
		if err != nil {
			return nil, ""
		}
		chal.Challenge[i] = rndIndex.Uint64()
	}

	ident, err := GetFingerprint(privkey)
	if err != nil {
		return nil, ""  // Could not get the keys fingerprint.
	}
	chal.Identity = ident

	// Sign the challenge with our private key
	challengeHash = HashStruct(chal)
	if signature, err := privkey.Sign(challengeHash); err != nil {
		return nil, ""
	} else {
		return &SignedStruct{Base: chal, Signature: signature}, challengeHash
	}
}

func (signedStruct *SignedStruct) VerifySample(samplerIdentity *conf.Identity, samplerPubkey *Keys) error {
	if samplerIdentity == nil {
		return errors.New("Samplers identity was nil.")
	} else if samplerPubkey == nil {
		return errors.New("Samplers pubkey was nil.")
	}
	sample, ok := signedStruct.Base.(*StorageSample)
	if !ok {
		return errors.New("Could not type assert the StorageChallengeProof.")
	}

	fp, err := GetFingerprint(samplerPubkey)
	if err != nil {
		return errors.New("Could not get fingerprint of Public key.")
	}
	if fp != sample.Identity || (samplerIdentity.Type != conf.Consensus && samplerIdentity.Type != conf.Client) {
		return errors.New("Challengers identity was unexpected.")
	}

	// Check if the proof is signed by the expected prover.
	if !signedStruct.Verify(samplerPubkey) {
		return errors.New("Could not verify signature of challenger.")
	}

	return nil
}

func (signedStruct *SignedStruct) VerifyChallenge(challengerIdentity *conf.Identity, challengerPubkey *Keys) error {
	if challengerIdentity == nil {
		return errors.New("Challengers identity was nil.")
	} else if challengerPubkey == nil {
		return errors.New("Challengers pubkey was nil.")
	}
	challenge, ok := signedStruct.Base.(*StorageChallenge)
	if !ok {
		return errors.New("Could not type assert the StorageChallengeProof.")
	}

	fp, err := GetFingerprint(challengerPubkey)
	if err != nil {
		return errors.New("Could not get fingerprint of Public key.")
	}
	if fp != challenge.Identity || (challengerIdentity.Type != conf.Consensus && challengerIdentity.Type != conf.Client) {
		return errors.New("Challengers identity was unexpected.")
	}

	// Check if the proof is signed by the expected prover.
	if !signedStruct.Verify(challengerPubkey) {
		return errors.New("Could not verify signature of challenger.")
	}

	return nil
}

// Helper function used with verifying proofs
func (sp *StorageSample) getSamplesMap() *map[uint64]byte {
	ret := make(map[uint64]byte, len(sp.Sampleindices))
	for i := 0; i < len(sp.Sampleindices); i++ {
		ret[sp.Sampleindices[i]] = sp.Samplevalues[i]
	}
	return &ret
}

func (signedStruct *SignedStruct) verifyChallengeProof(sampleBase string, challengerIdentity *conf.Identity, challengerPubkey *Keys,
	proverIdentity *conf.Identity, proverPubkey *Keys, challengeHash string) error{
	if challengerIdentity == nil {
		return errors.New("Challengers identity was nil.")
	} else if challengerPubkey == nil {
		return errors.New("Challengers pubkey was nil.")
	} else if proverIdentity == nil {
		return errors.New("Provers identity was nil.")
	} else if proverPubkey == nil {
		return errors.New("Provers pubkey was nil.")
	}
	scp, ok := signedStruct.Base.(*StorageChallengeProof)
	if !ok {
		return errors.New("Could not type assert the StorageChallengeProof.")
	}

	fpProver, err := GetFingerprint(proverPubkey)
	if err != nil {
		return errors.New("Could not get fingerprint of Public key.")
	}
	if fpProver != scp.Identity || proverIdentity.Type != conf.Storage {
		return errors.New("Provers identity was unexpected.")
	}

	challenge, ok := scp.Base.(*StorageChallenge)
	if !ok {
		return errors.New("Could not type assert the StorageChallenge.")
	}
	sc2 := &SignedStruct{Signature:signedStruct.Signature, Base:&StorageChallengeProof{Identity:scp.Identity, Filesize: scp.Filesize,
		Proof:scp.Proof, SignedStruct:SignedStruct{Signature:scp.Signature, Base:&StorageChallenge{Identity:challenge.Identity, Cid:challenge.Cid,
		Nonce:challenge.Nonce, Challenge:challenge.Challenge}}}}

	// Prevent the Prover of responding with a stored response to a previous issued challenge.
	if challengeHash != "" && HashStruct(scp.Base) != challengeHash {
		return errors.New("The proof contains an unexpected challenge.")
	}

	fpChallenger, err := GetFingerprint(challengerPubkey)
	if err != nil {
		return errors.New("Could not get fingerprint of Public key.")
	}

	if fpChallenger != challenge.Identity || (challengerIdentity.Type != conf.Consensus && challengerIdentity.Type != conf.Client) {
		return errors.New("Challengers identity was unexpected.")
	}

	// Check if the proof is signed by the expected prover.
	if !sc2.Verify(proverPubkey) {
		return errors.New("Could not verify signature of prover.")
	}

	// Check if the challenge is signed by the expected challenger
	if  !scp.Verify(challengerPubkey){
		return errors.New("Could not verify signature of challenger.")
	}

	storageSample := LoadStorageSample(sampleBase, challenge.Cid)
	if storageSample == nil || len(storageSample.Sampleindices) == 0 {
		return errors.New("Could not find a stored sample for this Cid.")
	}

	lenChal := len(challenge.Challenge)
	byteArr := make([]byte, lenChal)
	mapSamples := storageSample.getSamplesMap()
	for i := 0; i < lenChal; i++ {
		index := challenge.Challenge[i]
		if val, ok := (*mapSamples)[index]; !ok {
			return errors.Errorf("Missing sample for byte index: %v", index)
		} else {
			byteArr[i] = val
		}
	}

	if hash, err := HashData(byteArr); err != nil {
		return err
	} else if hash != scp.Proof {
		return errors.Errorf("Could not verify hash of proof.")
	}

	return nil
}

func (signedStruct *SignedStruct) VerifyChallengeProof_Historic(sampleBase string, challengerIdentity *conf.Identity, challengerPubkey *Keys,
		proverIdentity *conf.Identity, proverPubkey *Keys) error{
	return signedStruct.verifyChallengeProof(sampleBase, challengerIdentity, challengerPubkey, proverIdentity, proverPubkey, "")
}

func (signedStruct *SignedStruct) VerifyChallengeProof(sampleBase string, challengerIdentity *conf.Identity, challengerPubkey *Keys,
	proverIdentity *conf.Identity, proverPubkey *Keys, challengeHash string) error{
	return signedStruct.verifyChallengeProof(sampleBase, challengerIdentity, challengerPubkey, proverIdentity, proverPubkey, challengeHash)
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
	file := (*fileBytes)
	proof := &StorageChallengeProof{SignedStruct: *signedStruct, Identity: fingerprint,	Filesize: int64(len(file))}

	byteArr := make([]byte, lenChal)
	for i := 0; i < lenChal; i++ {
		byteArr[i] = file[challenge.Challenge[i]]
	}

	proof.Proof, err = HashData(byteArr)
	newSignedStruct, err := SignStruct(proof, privKey)

	if err != nil {
		return nil;
	}

	return newSignedStruct
}