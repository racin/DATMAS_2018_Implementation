package client

import (
	rpcClient "github.com/tendermint/tendermint/rpc/client"
	"github.com/racin/DATMAS_2018_Implementation/rpc"
	conf "github.com/racin/DATMAS_2018_Implementation/configuration"
	tmtypes "github.com/tendermint/tendermint/types"
	"encoding/json"

	"github.com/racin/DATMAS_2018_Implementation/types"

	"fmt"
	"net/http"
	"time"
	"io"
	"io/ioutil"
	"github.com/racin/DATMAS_2018_Implementation/crypto"
	"os"
	"strings"
	"context"
)

type Client struct {
	TMClient        			rpcClient.Client
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
	}

	TheClient.setupAPI()
}

func (c *Client) GetAccessList() (*conf.AccessList){
	return conf.GetAccessList(conf.ClientConfig().BasePath + conf.ClientConfig().AccessList)
}

func (c *Client) GetIdentityPublicKey(ident string) (identity *conf.Identity, pubkey *crypto.Keys){
	return crypto.GetIdentityPublicKey(ident, c.GetAccessList(), conf.ClientConfig().BasePath + conf.ClientConfig().PublicKeys)
}

func (c *Client) sendMultipartFormDataToIPFS(values *map[string]io.Reader) (*types.IPFSReponse) {
	buffer, boundary := rpc.GetMultipartValues(values)
	result := &types.IPFSReponse{}

	response, err := c.IPFSClient.Post(c.IPFSAddr + conf.ClientConfig().IpfsAddnopinEndpoint, boundary, buffer)
	if err == nil {
		if dat, err := ioutil.ReadAll(response.Body); err == nil {
			if err := json.Unmarshal(dat, &result); err != nil {
				result.AddMessageAndError(err.Error(), types.CodeType_InternalError)
			}
		}
	} else {
		result.AddMessageAndError(err.Error(), types.CodeType_InternalError)
	}

	return result
}

func (c *Client) setupAPI()  {
	tmApiFound, ipfsProxyFound := false, false

	// Get Tendermint blockchain API
	fmt.Printf("%+v\n", conf.ClientConfig().TendermintNodes)
	for _, ident := range conf.ClientConfig().TendermintNodes {
		addr := TheClient.GetAccessList().GetAddress(ident)
		if !tmApiFound {
			apiAddr := strings.Replace(conf.ClientConfig().RemoteAddr, "$TmNode", addr, 1)

			c.TMClient = rpcClient.NewHTTP(apiAddr, conf.ClientConfig().WebsocketEndPoint)
			if _, err := c.TMClient.Status(); err == nil {
				err := c.TMClient.Start()
				if err != nil {
					fmt.Println("Error starting: " + err.Error())
				}
				tmApiFound = true
				c.TMIdent = ident
			}
		}
	}

	if !tmApiFound {
		panic("Fatal: Could not estabilsh connection with Tendermint API.")
	}

	// Get IPFS Proxy API
	for _, ident := range conf.ClientConfig().IpfsNodes {
		addr := TheClient.GetAccessList().GetAddress(ident)
		ipfsAddr := strings.Replace(conf.ClientConfig().IpfsProxyAddr, "$IpfsNode", addr, 1)

		if response, err := c.IPFSClient.Post(ipfsAddr + conf.ClientConfig().IpfsIsupEndpoint, "application/json", nil); err == nil {
			dat, _ := ioutil.ReadAll(response.Body);
			ipfsResp := &types.IPFSReponse{}
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

func (c *Client) VerifyUpload(stx *crypto.SignedStruct) (types.CodeType, error) {
	byteArr, _ := json.Marshal(stx)
	return types.CheckBroadcastResult(c.TMClient.BroadcastTxSync(tmtypes.Tx(byteArr)))
}

func (c *Client) SubToNewBlock(newBlock chan interface{}) error {
	return c.TMClient.Subscribe(context.Background(), "bcfs-client", tmtypes.EventQueryNewBlock, newBlock)
}

func (c *Client) GetSignedTransaction(txtype types.TransactionType, data interface{}) (stranc *crypto.SignedStruct) {
	tx := types.NewTx(data, TheClient.fingerprint, txtype)
	stranc, err := crypto.SignStruct(tx, TheClient.privKey);
	if err != nil {
		panic("Could not sign transaction. Private/Public key pair may not match. Use the --generateKeys to generate a new one. Error: " + err.Error())
	}
	return
}