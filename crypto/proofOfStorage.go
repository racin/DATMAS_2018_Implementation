package crypto

type StorageSample struct {
	Sample					map[string]byte		`json:"sample"`
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

func GenerateStorageSample(fileByte *[]byte, privkeyPath string) *StorageSample{
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