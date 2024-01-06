package cmd

import (
	"fmt"
	"tidalcmd/api"

	"github.com/spf13/cobra"
)

var changeClientIdCmd = &cobra.Command{
	Use:   "changeClientId [client id]",
	Short: "Tries changing client ID, doesn't work for all clients",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		api, err := api.NewClientFromConfig()
		if err != nil {
			panic(err)
		}

		api.ClientId = args[0]
		api.ClientSecret = clientSecret

		clientName, err := api.Refresh()
		if err != nil {
			panic(err)
		}

		fmt.Println("Successfully logged in")
		fmt.Println("Client name:", clientName)
	},
}

func init() {
	rootCmd.AddCommand(changeClientIdCmd)

	changeClientIdCmd.Flags().StringVarP(&clientSecret, "client_secret", "s", "", "Client secret (optional)")
}
