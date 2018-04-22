package types

func (ipfsResponse *IPFSReponse) AddMessage(msg string) {
	ipfsResponse.Message = []byte(msg)
}
func (ipfsResponse *IPFSReponse) AddMessageAndError(msg string, ct CodeType) {
	ipfsResponse.Message = []byte(msg)
	ipfsResponse.Codetype = ct
}