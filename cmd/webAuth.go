package cmd

import (
	"fmt"
	"tidalcmd/api"

	"github.com/spf13/cobra"
)

var appMode string

var webAuthCmd = &cobra.Command{
	Use:   "webAuth [email] [password] [client id] [redirect uri]",
	Short: "Authenticates client using web flow (eg. Web/Desktop/Mobile clients)",
	Args:  cobra.ExactArgs(4),
	Run: func(cmd *cobra.Command, args []string) {
		email := args[0]
		password := args[1]
		clientId := args[2]
		redirectUri := args[3]

		if clientSecret == "" && appMode == "" {
			appMode = "WEB"
		}

		_, clientName, err := api.WebAuth(email, password, clientId, clientSecret, redirectUri, appMode)
		if err != nil {
			panic(err)
		}

		fmt.Println("Successfully logged in")
		fmt.Println("Client name:", clientName)
	},
}

func init() {
	rootCmd.AddCommand(webAuthCmd)

	webAuthCmd.Flags().StringVarP(&clientSecret, "client_secret", "s", "", "Client secret (optional)")
	webAuthCmd.Flags().StringVarP(&appMode, "app_mode", "a", "", "App mode (web/desktop/android/ios) (might be required on tidal clients)")
}
