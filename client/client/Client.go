package client

import (
	rpcClient "github.com/tendermint/tendermint/rpc/client"
	"github.com/racin/DATMAS_2018_Implementation/app"
	conf "github.com/racin/DATMAS_2018_Implementation/configuration"
	bt "github.com/racin/DATMAS_2018_Implementation/types"
	tmtypes "github.com/tendermint/tendermint/types"
	"encoding/json"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	"github.com/racin/DATMAS_2018_Implementation/types"
	"github.com/pkg/errors"
	"fmt"
	"net/http"
	"time"
	"io"
	"io/ioutil"
	"github.com/racin/DATMAS_2018_Implementation/crypto"
	"os"
	"math/rand"
	"strconv"
	"strings"
)

type Client struct {
	TMClient        			rpcClient.Client

	TMUploadClient				*http.Client
	TMUploadAPI					string
	TMIdent						string

	IPFSClient					*http.Client
	IPFSAddr					string
	IPFSIdent					string

	privKey						*crypto.Keys
	identity					*conf.Identity
	fingerprint					string
	cfgfile						string
}

var TheClient *Client
func NewClient() {
	var err error
	if cfgFile != "" {
		_, err = conf.LoadClientConfig(cfgFile);
	} else {
		_, err = conf.LoadClientConfig();
	}
	if err != nil {
		fmt.Println("Could not load configuration:", err)
		os.Exit(1)
	}

	TheClient = &Client {
		cfgfile: cfgFile,
		TMUploadClient: &http.Client{Timeout: time.Duration(conf.ClientConfig().UploadTimeoutSeconds) * time.Second},
		IPFSClient: &http.Client{Timeout: time.Duration(conf.ClientConfig().IpfsProxyTimeoutSeconds) * time.Second},
	}

	// Load private keys in order to later digitally sign transactions
	if myPrivKey, err := crypto.LoadPrivateKey(conf.ClientConfig().BasePath + conf.ClientConfig().PrivateKey); err != nil {
		panic("Could not load private key. Error: " + err.Error())
	} else if fp, err := crypto.GetFingerprint(myPrivKey); err != nil{
		panic("Could not get fingerprint of private key.")
	} else {
		TheClient.fingerprint = fp;
		TheClient.privKey = myPrivKey
		TheClient.identity = TheClient.GetAccessList().Identities[fp]
		fmt.Printf("%+v\n", TheClient)
	}

	TheClient.setupAPI()
}

func (c *Client) GetAccessList() (*conf.AccessList){
	return conf.GetAccessList(conf.ClientConfig().BasePath + conf.ClientConfig().AccessList)
}

func (c *Client) GetIdentityPublicKey(ident string) (identity *conf.Identity, pubkey *crypto.Keys){
	return crypto.GetIdentityPublicKey(ident, c.GetAccessList(), conf.ClientConfig().BasePath + conf.ClientConfig().PublicKeys)
}

func (c *Client) SendMultipartFormData(endpoint string, values *map[string]io.Reader) (bt.ResponseUpload) {
	buffer, boundary := app.GetMultipartValues(values)
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

func (c *Client) setupAPI()  {
	tmApiFound, tmUplApiFound, ipfsProxyFound := false, false, false

	// Get Tendermint blockchain API
	s1 := rand.NewSource(time.Now().UnixNano())
	reqNum := strconv.Itoa(rand.New(s1).Int())
	fmt.Printf("%+v\n", conf.ClientConfig().TendermintNodes)
	for _, ident := range conf.ClientConfig().TendermintNodes {
		addr := TheClient.GetAccessList().GetAddress(ident)
		if !tmApiFound {
			apiAddr := strings.Replace(conf.ClientConfig().RemoteAddr, "$TmNode", addr, 1)


			fmt.Println("Trying to connect to (TM_api: " + apiAddr)
			c.TMClient = rpcClient.NewHTTP(apiAddr, conf.ClientConfig().WebsocketEndPoint)
			if _, err := c.TMClient.Status(); err == nil {
				//conf.ClientConfig().RemoteAddr = apiAddr
				tmApiFound = true
				c.TMIdent = ident
			}
		}

		if !tmUplApiFound {
			uploadAddr := strings.Replace(conf.ClientConfig().UploadAddr, "$TmNode", addr, 1)
			fmt.Println("Trying to connect to (TM_uplApi): " + uploadAddr)

			values := map[string]io.Reader{
				"Status":    strings.NewReader(reqNum),
			}

			c.TMUploadAPI = uploadAddr + conf.ClientConfig().UploadEndPoint
			response := c.SendMultipartFormData(c.TMUploadAPI, &values);
			if response.Codetype == types.CodeType_OK && response.Message == reqNum{
				//conf.ClientConfig().UploadAddr = uploadAddr
				tmUplApiFound = true
			}
		}
	}

	if !tmApiFound || !tmUplApiFound {
		panic("Fatal: Could not estabilsh connection with Tendermint API.")
	}

	// Get IPFS Proxy API
	for _, ident := range conf.ClientConfig().IpfsNodes {
		addr := TheClient.GetAccessList().GetAddress(ident)
		ipfsAddr := strings.Replace(conf.ClientConfig().IpfsProxyAddr, "$IpfsNode", addr, 1)
		fmt.Println("Trying to connect to (IPFS addr): " + ipfsAddr)

		if response, err := c.IPFSClient.Post(ipfsAddr + conf.ClientConfig().IpfsIsupEndpoint, "application/json", nil); err == nil {
			dat, _ := ioutil.ReadAll(response.Body);
			ipfsResp := &types.IPFSReponse{}
			fmt.Printf("Isup: %s\n",dat)
			if json.Unmarshal(dat, ipfsResp) == nil && ipfsResp.Codetype == types.CodeType_OK {
				ipfsProxyFound = true
				c.IPFSAddr = ipfsAddr
				c.IPFSIdent= ident
				break
			}
		}
	}

	if !ipfsProxyFound {
		panic("Fatal: Could not estabilsh connection with IPFS Proxy API.")
	}
}


func (c *Client) VerifyUpload(stx *crypto.SignedStruct) (bt.CodeType, error) {
	byteArr, _ := json.Marshal(stx)
	return checkBroadcastResult(c.TMClient.BroadcastTxSync(tmtypes.Tx(byteArr)))
}
func (c *Client) UploadData(values *map[string]io.Reader) (bt.ResponseUpload) {
	//data := map[string]io.Reader
	fmt.Println("Uploadendpoint: " + c.TMUploadAPI)
	return c.SendMultipartFormData(c.TMUploadAPI, values)
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