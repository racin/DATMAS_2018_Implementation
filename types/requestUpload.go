package types

func (ru *RequestUpload) CompareTo(other *RequestUpload) bool {
	return ru.Cid == other.Cid && ru.Length == other.Length && ru.IpfsNode == other.IpfsNode
}

func GetRequestUploadFromMap(derivedStruct map[string]interface{}) *RequestUpload {
	if cid, ok := derivedStruct["cid"]; ok {
		if ipfsNode, ok := derivedStruct["ipfsNode"]; ok {
			if length, ok := derivedStruct["length"]; ok {
				return &RequestUpload{Cid: cid.(string), IpfsNode: ipfsNode.(string), Length:int64(length.(float64))}
			}
		}
	}
	return nil
}
