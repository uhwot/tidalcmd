package cmd

import (
	"fmt"
	"tidalcmd/api"

	"github.com/spf13/cobra"
)

var clientSecret string

var deviceAuthCmd = &cobra.Command{
	Use:   "deviceAuth [client id]",
	Short: "Authenticates client using device flow (eg. TV/Automotive clients)",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		_, clientName, err := api.DeviceAuth(args[0], clientSecret)
		if err != nil {
			panic(err)
		}

		fmt.Println("Successfully logged in")
		fmt.Println("Client name:", clientName)
	},
}

func init() {
	rootCmd.AddCommand(deviceAuthCmd)

	deviceAuthCmd.Flags().StringVarP(&clientSecret, "client_secret", "s", "", "Client secret (optional)")
}
