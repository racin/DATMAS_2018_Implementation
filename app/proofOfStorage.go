package app

type StorageSample struct {
	cid						string
	sample					map[string]byte
}

type StorageChallenge struct {
	cid						string
	challengesignature		[]byte
	challenge				[]byte
}

type StorageChallengeProof struct {
	StorageChallenge
	proof					[]byte
	proofsignature			[]byte
}


func (app *Application) GenerateStorageSample(fileByte *[]byte) *StorageSample{
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