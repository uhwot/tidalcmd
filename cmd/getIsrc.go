package cmd

import (
	"fmt"
	"tidalcmd/api"

	"github.com/spf13/cobra"
)

var getIsrcCmd = &cobra.Command{
	Use:   "getIsrc [isrc]",
	Short: "Gets track list by ISRC",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		api, err := api.NewClientFromConfig()
		if err != nil {
			panic(err)
		}

		isrc := args[0]

		tracks, err := api.GetTracksByIsrc(isrc)
		if err != nil {
			panic(err)
		}

		//fmt.Printf("%+v", tracks)

		for i, track := range tracks.Items {
			fmt.Printf("%d. %s - %s", i+1, track.Artist.Name, track.Title)
			if track.Explicit {
				fmt.Print(" [E]")
			}
			fmt.Printf(" [%s]", track.Album.Title)
			for _, tag := range track.MediaMetadata.Tags {
				fmt.Printf(" [%s]", tag)
			}
			fmt.Printf(" [%d]\n", track.ID)
		}
	},
}

func init() {
	rootCmd.AddCommand(getIsrcCmd)
}
