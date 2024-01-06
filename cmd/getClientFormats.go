package cmd

import (
	"fmt"
	"tidalcmd/api"
	"tidalcmd/manifest"

	"github.com/spf13/cobra"
)

const MQA_ID = 177525913
const HIRES_FLAC_ID = 296612885
const SONY_360_ID = 203033693
const ATMOS_ID = 233057630

var getClientFormatsCmd = &cobra.Command{
	Use:   "getClientFormats",
	Short: "Gets formats supported by client and extra info",
	Run: func(cmd *cobra.Command, args []string) {
		api, err := api.NewClientFromConfig()
		if err != nil {
			panic(err)
		}

		tracksToTest := []struct {
			TrackId int
			quality string
		}{
			{MQA_ID, "LOW"},
			{MQA_ID, "HIGH"},
			{MQA_ID, "LOSSLESS"},
			{MQA_ID, "HI_RES"},
			{HIRES_FLAC_ID, "HI_RES_LOSSLESS"},
			{SONY_360_ID, "HI_RES"},
			{ATMOS_ID, "HI_RES"},
		}

		fmt.Println("Supported formats:")

		for _, track := range tracksToTest {
			info, err := api.GetTrackPlaybackInfo(track.TrackId, track.quality, playbackMode)
			if err != nil {
				panic(err)
			}

			if track.quality != info.AudioQuality && info.AudioMode == "STEREO" {
				continue
			}

			manifestInfo, err := manifest.GetInfoFromManifest(info.Manifest, info.ManifestMimeType)
			if err != nil {
				panic(err)
			}

			if info.AudioQuality == "HI_RES_LOSSLESS" {
				fmt.Print("24bit flac")
			} else if manifestInfo.Codecs != "" {
				fmt.Print(manifestInfo.Codecs)
			} else {
				fmt.Print(info.AudioQuality)
			}
			if manifestInfo.Encrypted {
				fmt.Print(", encrypted")
			}
			fmt.Print(", ", info.ManifestMimeType)
			fmt.Print("\n")
		}
	},
}

func init() {
	rootCmd.AddCommand(getClientFormatsCmd)

	getClientFormatsCmd.Flags().StringVarP(&playbackMode, "playback_mode", "m", "STREAM", "Playback mode (STREAM/OFFLINE)")
}
