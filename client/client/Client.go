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
	"math/rand"
	"strconv"
	"strings"
	"context"
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

func (c *Client) sendMultipartFormDataToTM(values *map[string]io.Reader) (*types.ResponseUpload) {
	buffer, boundary := rpc.GetMultipartValues(values)
	var result *types.ResponseUpload

	response, err := c.TMUploadClient.Post(c.TMUploadAPI, boundary, buffer)
	if err == nil {
		if dat, err := ioutil.ReadAll(response.Body); err == nil {
			fmt.Printf("Got response: %#v\n", string(dat))
			if err := json.Unmarshal(dat, &result); err != nil {
				result = &types.ResponseUpload{Codetype:types.CodeType_InternalError, Message:err.Error()}
			}
		}
	} else {
		result = &types.ResponseUpload{Codetype:types.CodeType_InternalError, Message:err.Error()}
	}
	fmt.Printf("The result: %#v\n", result)
	return result
}

func (c *Client) sendMultipartFormDataToIPFS(values *map[string]io.Reader) (*types.IPFSReponse) {
	buffer, boundary := rpc.GetMultipartValues(values)
	result := &types.IPFSReponse{}

	response, err := c.IPFSClient.Post(c.IPFSAddr + conf.ClientConfig().IpfsAddnopinEndpoint, boundary, buffer)
	if err == nil {
		if dat, err := ioutil.ReadAll(response.Body); err == nil {
			fmt.Printf("Got response: %#v\n", string(dat))
			if err := json.Unmarshal(dat, &result); err != nil {
				result.AddMessageAndError(err.Error(), types.CodeType_InternalError)
			}
		}
	} else {
		result.AddMessageAndError(err.Error(), types.CodeType_InternalError)
	}
	fmt.Printf("The result: %+v\n", result)
	return result
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
				err := c.TMClient.Start()
				if err != nil {
					fmt.Println("Error starting: " + err.Error())
				}
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
			response := c.sendMultipartFormDataToTM(&values);
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

func (c *Client) VerifyUpload(stx *crypto.SignedStruct) (types.CodeType, error) {
	byteArr, _ := json.Marshal(stx)
	return types.CheckBroadcastResult(c.TMClient.BroadcastTxSync(tmtypes.Tx(byteArr)))
}
func (c *Client) UploadDataToTM(values *map[string]io.Reader) (*types.ResponseUpload) {
	//data := map[string]io.Reader
	fmt.Println("Uploadendpoint: " + c.TMUploadAPI)
	return c.sendMultipartFormDataToTM(values)
	//return checkBroadcastResult(c.TM.BroadcastTxSync(tmtypes.Tx(byteArr)))
}
func (c *Client) UploadDataToIPFS(values *map[string]io.Reader) (*types.IPFSReponse) {
	//data := map[string]io.Reader
	fmt.Println("IPFS Upload: " + c.IPFSAddr + conf.ClientConfig().IpfsAddnopinEndpoint)
	return c.sendMultipartFormDataToIPFS(values)
	//return checkBroadcastResult(c.TM.BroadcastTxSync(tmtypes.Tx(byteArr)))
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