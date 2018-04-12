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
	"net/http"
	"time"
	"io"
	"bytes"
	"mime/multipart"
	"os"
	"io/ioutil"
)

type BaseClient struct {
	TMClient        			rpcClient.Client

	TMUploadClient				*http.Client
	TMUploadAPI					string

	IPFSClient					*http.Client
	IPFSAPI						string
}

func NewTMHTTPClient(endpoint string) *BaseClient {
	tm := rpcClient.NewHTTP(endpoint, conf.ClientConfig().WebsocketEndPoint)
	TMhttpClient := &http.Client{Timeout: time.Duration(conf.ClientConfig().UploadTimeoutSeconds) * time.Second}
	IpfshttpClient := &http.Client{Timeout: time.Duration(conf.ClientConfig().UploadTimeoutSeconds) * time.Second}
	return &BaseClient{TMClient: tm, TMUploadClient: TMhttpClient, IPFSClient: IpfshttpClient}
}

func (c *BaseClient) SendMultipartFormData(endpoint string, values *map[string]io.Reader) (bt.ResponseUpload) {
	buffer, boundary := getMultipartValues(values)
	var result bt.ResponseUpload

	response, err := c.TMUploadClient.Post(endpoint, boundary, buffer)
	if err == nil {
		if dat, err := ioutil.ReadAll(response.Body); err == nil {
			fmt.Printf("Got response: %#v\n", string(dat))
			if err := json.Unmarshal(dat, &result); err != nil {
				result = bt.ResponseUpload{Codetype:bt.CodeType_InternalError, Message:err.Error()}
			}
		}
	} else {
		result = bt.ResponseUpload{Codetype:bt.CodeType_InternalError, Message:err.Error()}
	}
	fmt.Printf("The result: %#v\n", result)
	return result
}

func getMultipartValues(values *map[string]io.Reader) (buffer *bytes.Buffer, boundary string){
	var b bytes.Buffer
	var err error
	w := multipart.NewWriter(&b)
	defer w.Close()

	for index, element := range *values {
		var writer io.Writer
		// If file has a close method.
		if file, ok := element.(io.Closer); ok {
			defer file.Close()
		}

		// Check if a file is added. Else add it as a regular data element.
		if file, ok := element.(*os.File); ok {
			writer, err = w.CreateFormFile(index, file.Name());
		} else {
			writer, err = w.CreateFormField(index);
		}

		// If there are problems with adding an element, continue to the next one.
		if err != nil {
			continue
		}

		if _, err = io.Copy(writer, element); err != nil {
			continue
		}

	}

	return &b, w.FormDataContentType()
}

func checkBroadcastResult(commit interface{}, err error) (bt.CodeType, error) {
	fmt.Printf("Data: %+v\n", commit)
	if c, ok := commit.(*ctypes.ResultBroadcastTxCommit); ok && c != nil {
		if err != nil {
			return bt.CodeType_InternalError, err
		} else if c.CheckTx.IsErr() {
			return bt.CodeType_InternalError, errors.New(c.CheckTx.String())
		} else if c.DeliverTx.IsErr() {
			return bt.CodeType_InternalError, errors.New(c.DeliverTx.String())
		} else {
			fmt.Printf("Data: %+v\n", c)
			return bt.CodeType_OK, nil;
		}
	} else if c, ok := commit.(*ctypes.ResultBroadcastTx); ok && c != nil {
		fmt.Printf("Data: %+v\n", c)
		code := bt.CodeType(c.Code)
		if code == bt.CodeType_OK {
			fmt.Printf("Data: %+v\n", c)
			return code, nil
		} else {
			return code, errors.New("CheckTx. Log: " + c.Log + ", Code: " + bt.CodeType_name[int32(c.Code)])
		}
	}
	/**/
	return bt.CodeType_InternalError, errors.New("Could not type assert result.")
}

func (c *BaseClient) BeginUploadData(stx *app.SignedTransaction) (bt.CodeType, error) {
	byteArr, _ := json.Marshal(stx)
	return checkBroadcastResult(c.TM.BroadcastTxSync(tmtypes.Tx(byteArr)))
}
func (c *BaseClient) EndUploadData(values *map[string]io.Reader) (bt.ResponseUpload) {
	//data := map[string]io.Reader
	fmt.Println("Uploadendpoint: " + conf.ClientConfig().UploadAddr)
	return c.SendMultipartFormData(conf.ClientConfig().UploadAddr, values)
	//return checkBroadcastResult(c.TM.BroadcastTxSync(tmtypes.Tx(byteArr)))
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