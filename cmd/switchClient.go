package cmd

import (
	"fmt"
	"tidalcmd/api"

	"github.com/spf13/cobra"
)

var switchClientCmd = &cobra.Command{
	Use:   "switchClient [client id]",
	Short: "Switches client (used for speaker/chromecast clients)",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		api, err := api.NewClientFromConfig()
		if err != nil {
			panic(err)
		}

		clientName, err := api.SwitchClient(args[0])
		if err != nil {
			panic(err)
		}

		fmt.Println("Successfully logged in")
		fmt.Println("Client name:", clientName)
	},
}

func init() {
	rootCmd.AddCommand(switchClientCmd)
}
