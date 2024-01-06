package cmd

import (
	"fmt"
	"tidalcmd/api"

	"github.com/spf13/cobra"
)

var getUpcCmd = &cobra.Command{
	Use:   "getUpc [upc]",
	Short: "Gets album list by UPC",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		api, err := api.NewClientFromConfig()
		if err != nil {
			panic(err)
		}

		upc := args[0]

		albums, err := api.GetAlbumsByUpc(upc)
		if err != nil {
			panic(err)
		}

		//fmt.Printf("%+v", albums)

		for i, album := range albums.Items {
			fmt.Printf("%d. %s - %s", i+1, album.Artist.Name, album.Title)
			if album.Explicit {
				fmt.Print(" [E]")
			}
			for _, tag := range album.MediaMetadata.Tags {
				fmt.Printf(" [%s]", tag)
			}
			fmt.Printf(" [%d]\n", album.ID)
		}
	},
}

func init() {
	rootCmd.AddCommand(getUpcCmd)
}
