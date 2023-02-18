package cmd

import (
	"github.com/spf13/cobra"
)

// tokenCmd tries to obtain a token from the Auth provider
// Currently tailored for Auth0
var tokenCmd = &cobra.Command{
	Use:   "token",
	Short: "Obtain a token from Auth0",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {

		token, ok := readTokenFromCache()
		if ok {
			cmd.Println(token)
			return nil
		}

		token, err := obtainAccessToken(cmd)
		if err != nil {
			return err
		}

		cmd.Println(token)

		return writeTokenToCache(token)
	},
}

func init() {
	rootCmd.AddCommand(tokenCmd)
}
