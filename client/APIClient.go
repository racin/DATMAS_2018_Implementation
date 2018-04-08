package client

import (
	"github.com/racin/DATMAS_2018_Implementation/app"
	bt "github.com/racin/DATMAS_2018_Implementation/types"
)

type API interface {
	GetBase() *BaseClient
	//ProofAPI
	DataAPI
	//AccessAPI
}
/*
type ProofAPI interface {
	GenerateProof(tx app.BasicTransaction) ([]byte, error)
	VerifyProof(tx app.BasicTransaction) (bool, error)
}*/

type DataAPI interface {
	/*DownloadData(tx app.BasicTransaction) error
	RemoveData(tx app.BasicTransaction) error*/
	BeginUploadData(tx *app.SignedTransaction) (bt.CodeType, error)
	EndUploadData(tx *app.SignedTransaction) (bt.CodeType, error)
	//EndUploadData(values map[string]io.Reader) error
}
/*
type AccessAPI interface {
	ChangeContentAccess(tx app.BasicTransaction) error
}*/


func NewTM_API(endpoint string) API {
	base := NewTMHTTPClient(endpoint)
	return &apiClient{endpoint, base}
}

type apiClient struct {
	endpoint string
	base     *BaseClient
}

func (api *apiClient) GetBase() (*BaseClient) {
	return api.base
}
/** Proof API **/
/*
func (api *apiClient) GenerateProof(tx app.BasicTransaction) ([]byte, error) {
	return api.base.GetAccount(id)
}

func (api *apiClient) VerifyProof(tx app.BasicTransaction) ([]byte, error) {
	return api.base.DelAccount(id)
}*/

/** Data API **/
/*
func (api *apiClient) DownloadData(tx app.BasicTransaction) (error) {
	key, err := crypto.CreateKeyPair()
	if err != nil {
		return "", "", err
	}
	api.base.Key = key
	if err := api.base.AddAccount(&state.Account{ID: id, PubKey: key.GetPubString()}); err != nil {
		return "", "", err
	}
	return key.GetPubString(), key.GetPrivString(), nil
}

func (api *apiClient) RemoveData(tx app.BasicTransaction) (error) {
	return api.base.GetAccount(id)
}
*/
func (api *apiClient) BeginUploadData(tx *app.SignedTransaction) (bt.CodeType, error) {
	return api.base.BeginUploadData(tx)
}

func (api *apiClient) EndUploadData(tx *app.SignedTransaction) (bt.CodeType, error) {
	return api.base.EndUploadData(tx)
}

/** Access API **/
/*
func (api *apiClient) ChangeContentAccess(tx app.BasicTransaction) (error) {

}
*/
/*
func (api *apiClient) EndUploadData(sid string, value string) error {
	s := &state.Secret{
		ID:     sid,
		Value:  value,
		Shares: make(map[string]string),
		Owners: map[string]bool{
			api.base.AccountID: true,
		},
	}
	aesKey, err := s.Encrypt()
	if err != nil {
		return err
	}
	k := api.base.Key
	encryptedAESKey, err := k.EncryptToString(aesKey)
	if err != nil {
		return err
	}
	s.Shares[api.base.AccountID] = encryptedAESKey
	if err := api.base.AddSecret(s); err != nil {
		return err
	}
	return nil
}

func (api *apiClient) GetSecret(sid string) (*state.Secret, error) {
	secret, err := api.base.GetSecret(sid)
	if err != nil {
		return nil, err
	}
	encryptedAESKey, ok := secret.Shares[api.base.AccountID]
	if ok {
		k := api.base.Key
		aesKey, err := k.DecryptString(encryptedAESKey)
		if err != nil {
			return nil, err
		}
		err = secret.Decrypt(aesKey)
		if err != nil {
			return nil, err
		}
	}
	return secret, nil
}

func (api *apiClient) DeleteSecret(sid string) error {
	return api.base.DelSecret(sid)
}

func (api *apiClient) ListSecrets(sidPrefix string) ([]*state.Secret, error) {
	secrets, err := api.base.ListSecrets()
	if err != nil {
		return nil, err
	}
	for _, s := range secrets {
		if encryptedKey, ok := s.Shares[api.base.AccountID]; ok {
			key, e := api.base.Key.DecryptString(encryptedKey)
			if e != nil {
				log.Fatal(e)
			}
			if e = s.Decrypt(key); e != nil {
				log.Fatal(e)
			}
		}
	}
	return secrets, nil
}

func (api *apiClient) ShareSecret(sid, accountID string, ownerRights bool) error {
	secret, err := api.base.GetSecret(sid)
	if err != nil {
		return err
	}
	encryptedAESKey, ok := secret.Shares[api.base.AccountID]
	if !ok {
		return errors.New("no share for us on this secret")
	}
	k := api.base.Key
	aesKey, err := k.DecryptString(encryptedAESKey)
	if err != nil {
		return err
	}
	acc, err := api.GetAccount(accountID)
	if err != nil {
		log.Print("can not find account " + accountID)
		return err
	}
	otherKey, err := crypto.NewFromStrings(acc.PubKey, "")
	if err != nil {
		return err
	}
	otherEncrptedAESKey, err := otherKey.EncryptToString(aesKey)
	if err != nil {
		return err
	}
	secret.Shares[accountID] = otherEncrptedAESKey
	if ownerRights {
		secret.Owners[accountID] = true
	}
	return api.base.UpdateSecret(secret)
}

func (api *apiClient) UpdateSecret(sid, value string) error {
	sec, err := api.base.GetSecret(sid)
	if err != nil {
		return fmt.Errorf("failed to get secret: %v", err)
	}
	aesKey, err := api.base.Key.DecryptString(sec.Shares[api.base.AccountID])
	if err != nil {
		return fmt.Errorf("failed to decrypt secret: %v", err)
	}
	sec.Value = value
	err = sec.EncryptWithKey(aesKey)
	if err != nil {
		return err
	}
	if err := api.base.UpdateSecret(sec); err != nil {
		return err
	}
	return nil
}

func (api *apiClient) UnshareSecret(sid, accountID string) error {
	secret, err := api.base.GetSecret(sid)
	if err != nil {
		log.Fatal(err)
	}
	delete(secret.Shares, accountID)
	delete(secret.Owners, accountID)
	return api.base.UpdateSecret(secret)
}

func (api *apiClient) GiveReputation(receiver string, value int) error {
	return api.base.GiveReputation(api.base.AccountID, receiver, value)
}
*/