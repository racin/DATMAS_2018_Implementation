package client

import (
	rpcClient "github.com/tendermint/tendermint/rpc/client"
	"github.com/racin/DATMAS_2018_Implementation/app"
	conf "github.com/racin/DATMAS_2018_Implementation/configuration"
	bt "github.com/racin/DATMAS_2018_Implementation/types"
	tmtypes "github.com/tendermint/tendermint/types"
	"encoding/json"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	"github.com/pkg/errors"
	"fmt"
)

type BaseClient struct {
	TM        rpcClient.Client
}

func NewHTTPClient(endpoint string) *BaseClient {
	tm := rpcClient.NewHTTP(endpoint, conf.ClientConfig().WebsocketEndPoint)
	return &BaseClient{tm}
}


func checkBroadcastResult(commit interface{}, err error) error {
	fmt.Printf("Data: %+v\n", commit)
	if c, ok := commit.(*ctypes.ResultBroadcastTxCommit); ok && c != nil {
		if err != nil {
			return err
		} else if c.CheckTx.IsErr() {
			return errors.New(c.CheckTx.String())
		} else if c.DeliverTx.IsErr() {
			return errors.New(c.DeliverTx.String())
		} else {
			fmt.Printf("Data: %+v\n", c)
			return nil;
		}
	} else if c, ok := commit.(*ctypes.ResultBroadcastTx); ok && c != nil {
		fmt.Printf("Data: %+v\n", c)
		if bt.CodeType(c.Code) == bt.CodeType_OK {
			fmt.Printf("Data: %+v\n", c)
			return nil
		} else {
			return errors.New("CheckTx: " + bt.CodeType_name[int32(c.Code)])
		}
	}
	/**/
	return errors.New("Could not type assert result.")
}

func (c *BaseClient) BeginUploadData(stx *app.SignedTransaction) error {
	byteArr, _ := json.Marshal(stx)
	return checkBroadcastResult(c.TM.BroadcastTxSync(tmtypes.Tx(byteArr)))
}
/*
func (c *BaseClient) AddAccount(acc *state.Account) error {
	tx := transaction.New(transaction.AccountAdd, &transaction.AccountAddData{Account: acc})
	if err := tx.ProofOfWork(transaction.DefaultProofOfWorkCost); err != nil {
		return err
	}
	bs, _ := tx.ToBytes()
	res, err := c.tm.BroadcastTxCommit(types.Tx(bs))
	if err != nil {
		return err
	}
	if res.CheckTx.IsErr() {
		return errors.New(res.CheckTx.Error())
	}
	if res.DeliverTx.IsErr() {
		return errors.New(res.DeliverTx.Error())
	}
	return nil
}

func (c *BaseClient) DelAccount(id string) error {
	tx := transaction.New(transaction.AccountDel, &transaction.AccountDelData{ID: id})
	if err := tx.ProofOfWork(transaction.DefaultProofOfWorkCost); err != nil {
		return err
	}
	if err := tx.Sign(c.Key); err != nil {
		return err
	}
	bs, _ := tx.ToBytes()
	res, err := c.tm.BroadcastTxCommit(types.Tx(bs))
	if err != nil {
		return err
	}
	if res.CheckTx.IsErr() {
		return errors.New(res.CheckTx.Error())
	}
	if res.DeliverTx.IsErr() {
		return errors.New(res.DeliverTx.Error())
	}
	return nil
}

func (c *BaseClient) GiveReputation(from, to string, value int) error {
	tx := transaction.New(transaction.ReputationGive, &transaction.ReputationGiveData{
		From:  from,
		To:    to,
		Value: value,
	})
	if err := tx.ProofOfWork(transaction.DefaultProofOfWorkCost); err != nil {
		return err
	}
	if err := tx.Sign(c.Key); err != nil {
		return err
	}
	bs, _ := tx.ToBytes()
	res, err := c.tm.BroadcastTxCommit(types.Tx(bs))
	if err != nil {
		return err
	}
	if res.CheckTx.IsErr() {
		return errors.New(res.CheckTx.Error())
	}
	if res.DeliverTx.IsErr() {
		return errors.New(res.DeliverTx.Error())
	}
	return nil
}

func (c *BaseClient) GetAccount(id string) (*state.Account, error) {
	resp, err := c.tm.ABCIQuery("/account", []byte(id), false)
	if err != nil {
		return nil, err
	}
	if len(resp.Value) == 0 {
		return nil, errors.New("account not found")
	}
	acc := &state.Account{}
	if err = json.Unmarshal(resp.Value, acc); err != nil {
		log.Printf("request account but got rubbish: %v", string(resp.Value))
		return nil, err
	}
	return acc, nil
}

func (c *BaseClient) ListAccounts() ([]*state.Account, error) {
	resp, err := c.tm.ABCIQuery("/account", nil, false)
	if err != nil {
		log.Print(err)
		return nil, err
	}
	if len(resp.Value) == 0 {
		return nil, errors.New("account not found")
	}
	acc := []*state.Account{}
	if err = json.Unmarshal(resp.Value, &acc); err != nil {
		return nil, err
	}
	return acc, nil
}

func (c *BaseClient) ListSecrets() ([]*state.Secret, error) {
	resp, err := c.tm.ABCIQuery("/secret", nil, false)
	if err != nil {
		log.Print(err)
		return nil, err
	}
	if len(resp.Value) == 0 {
		return nil, errors.New("secret not found")
	}
	acc := []*state.Secret{}
	if err = json.Unmarshal(resp.Value, &acc); err != nil {
		return nil, err
	}
	return acc, nil
}

func (c *BaseClient) GetSecret(id string) (*state.Secret, error) {
	resp, err := c.tm.ABCIQuery("/secret", []byte(id), false)
	if err != nil {
		return nil, err
	}
	if len(resp.Value) == 0 {
		return nil, errors.New("secret not found")
	}
	acc := &state.Secret{}
	if err = json.Unmarshal(resp.Value, acc); err != nil {
		return nil, err
	}
	return acc, nil
}

func (c *BaseClient) AddSecret(acc *state.Secret) error {
	tx := transaction.New(transaction.SecretAdd, &transaction.SecretAddData{Secret: acc})
	if err := tx.ProofOfWork(transaction.DefaultProofOfWorkCost); err != nil {
		return err
	}
	bs, _ := tx.ToBytes()
	res, err := c.tm.BroadcastTxCommit(types.Tx(bs))
	if err != nil {
		return err
	}
	if res.CheckTx.IsErr() {
		return errors.New(res.CheckTx.Error())
	}
	if res.DeliverTx.IsErr() {
		return errors.New(res.DeliverTx.Error())
	}
	return nil
}

func (c *BaseClient) DelSecret(id string) error {
	tx := transaction.New(transaction.SecretDel, &transaction.SecretDelData{
		ID:       id,
		SenderID: c.AccountID,
	})
	if err := tx.ProofOfWork(transaction.DefaultProofOfWorkCost); err != nil {
		return err
	}
	if err := tx.Sign(c.Key); err != nil {
		return err
	}
	bs, _ := tx.ToBytes()
	res, err := c.tm.BroadcastTxCommit(types.Tx(bs))
	if err != nil {
		return err
	}
	if res.CheckTx.IsErr() {
		return errors.New(res.CheckTx.Error())
	}
	if res.DeliverTx.IsErr() {
		return errors.New(res.DeliverTx.Error())
	}
	return nil
}

func (c *BaseClient) UpdateSecret(acc *state.Secret) error {
	tx := transaction.New(transaction.SecretUpdate, &transaction.SecretUpdateData{
		Secret:   acc,
		SenderID: c.AccountID,
	})
	if err := tx.ProofOfWork(transaction.DefaultProofOfWorkCost); err != nil {
		return err
	}
	if err := tx.Sign(c.Key); err != nil {
		return err
	}
	bs, _ := tx.ToBytes()
	res, err := c.tm.BroadcastTxCommit(types.Tx(bs))
	if err != nil {
		return err
	}
	if res.CheckTx.IsErr() {
		return errors.New(res.CheckTx.Error())
	}
	if res.DeliverTx.IsErr() {
		return errors.New(res.DeliverTx.Error())
	}
	return nil
}
*/