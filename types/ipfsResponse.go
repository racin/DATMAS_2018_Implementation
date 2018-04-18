package types

func (ipfsResponse *IPFSReponse) AddMessage(msg string) {
	ipfsResponse.Message = []byte(msg)
}
