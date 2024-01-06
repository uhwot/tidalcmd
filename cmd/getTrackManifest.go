package cmd

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"tidalcmd/api"

	"github.com/spf13/cobra"
)

var quality string
var playbackMode string

var getTrackManifestCmd = &cobra.Command{
	Use:   "getTrackManifest [track id]",
	Short: "Gets track stream manifest",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("requires track id argument")
		}
		_, err := strconv.Atoi(args[0])
		if err != nil {
			return errors.New("track id needs to be an integer")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		api, err := api.NewClientFromConfig()
		if err != nil {
			panic(err)
		}

		trackId, _ := strconv.Atoi(args[0])

		info, err := api.GetTrackPlaybackInfo(trackId, quality, playbackMode)
		if err != nil {
			panic(err)
		}

		fmt.Println("Audio quality:", info.AudioQuality)
		fmt.Println("Audio mode:", info.AudioMode)
		fmt.Println("Manifest MIME type:", info.ManifestMimeType)

		fmt.Print("\n")

		manifest, err := base64.StdEncoding.DecodeString(info.Manifest)
		if err != nil {
			panic(err)
		}

		fmt.Println(string(manifest))
	},
}

func init() {
	rootCmd.AddCommand(getTrackManifestCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	getTrackManifestCmd.Flags().StringVarP(&quality, "quality", "q", "HI_RES_LOSSLESS", "Audio quality")
	getTrackManifestCmd.Flags().StringVarP(&playbackMode, "playback_mode", "m", "STREAM", "Playback mode (STREAM/OFFLINE)")
}
