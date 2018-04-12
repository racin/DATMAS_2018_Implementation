package main

import (
	"github.com/racin/DATMAS_2018_Implementation/cmd/cmd"
)

/*
type Metadata struct {
	Entries		map[string]MetadataEntry	`json:"entries"`
}*/
type MetadataEntry struct {
	Name 						string		`json:"name"`
	Description					string		`json:"description"`
}

func main() {
	cmd.Execute()
}
