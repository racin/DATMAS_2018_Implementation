package client

import (
	"io/ioutil"
	"encoding/json"
	conf "github.com/racin/DATMAS_2018_Implementation/configuration"
	"github.com/racin/DATMAS_2018_Implementation/crypto"
)

/*
type Metadata struct {
	Entries		map[string]MetadataEntry	`json:"entries"`
}*/
type MetadataEntry struct {
	crypto.StorageSample
	Name 						string		`json:"name"`
	Description					string		`json:"description"`
	Blockheight					int64		`json:"blockheight"`
}

func GetMetadata(cid string, mePath ...string) (*MetadataEntry){
	var path string
	if len(mePath) == 0 {
		path = conf.ClientConfig().BasePath + conf.ClientConfig().Metadata
	} else {
		path = mePath[0]
	}

	var me *MetadataEntry = &MetadataEntry{}
	if data, err := ioutil.ReadFile(path + cid); err == nil {
		json.Unmarshal(data, me)
	}

	return me
}

func WriteMetadata(cid string, me *MetadataEntry) error {
	if data, err := json.Marshal(*me); err == nil {
		return ioutil.WriteFile(conf.ClientConfig().BasePath + conf.ClientConfig().Metadata + cid, data, 0600)
	} else {
		return err
	}
}
