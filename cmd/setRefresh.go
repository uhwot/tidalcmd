package cmd

import (
	"fmt"
	"tidalcmd/api"

	"github.com/spf13/cobra"
)

var setRefreshCmd = &cobra.Command{
	Use:   "setRefresh [client id] [refresh token]",
	Short: "Sets client + refresh token",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		_, clientName, err := api.RefreshTokenAuth(args[0], clientSecret, args[1])
		if err != nil {
			panic(err)
		}

		fmt.Println("Successfully logged in")
		fmt.Println("Client name:", clientName)
	},
}

func init() {
	rootCmd.AddCommand(setRefreshCmd)

	setRefreshCmd.Flags().StringVarP(&clientSecret, "client_secret", "s", "", "Client secret (optional)")
}
