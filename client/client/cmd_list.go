package client

import (
	"github.com/spf13/cobra"
	"log"
	conf "github.com/racin/DATMAS_2018_Implementation/configuration"
	"fmt"
	"github.com/racin/DATMAS_2018_Implementation/types"
	"io/ioutil"
)

var listCmd = &cobra.Command{
	Use:     "list [CID]",
	Aliases: []string{"list"},
	Short:   "List available metadata",
	Long:    `Lists all the metadata which the client possesses.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			// List detailed data for a single entry
			cid := args[0]
			me := types.GetMetadata(cid)
			fmt.Printf("Listing detailed metadata. %+v\n", me)
		} else {
			// List simple data for every entry
			readdir, err := ioutil.ReadDir(conf.ClientConfig().BasePath + conf.ClientConfig().Metadata)
			if err != nil {
				log.Fatal(err.Error())
			}
			for i, file := range readdir {
				me := types.GetMetadata(file.Name())
				fmt.Printf("Entry %v. CID: %v Name: %v\n", i, me.Cid, me.Name)
			}
		}
	},
}

func init() {
	RootCmd.AddCommand(listCmd)
}