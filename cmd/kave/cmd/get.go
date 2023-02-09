/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/spf13/cobra"
)

// getCmd gets a key value from a kave server
var getCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Get a key value from a kave server",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]

		u, err := createRequestUrl(cmd, key)
		if err != nil {
			return err
		}

		req, err := http.NewRequest(http.MethodGet, u.String(), nil)
		if err != nil {
			return err
		}

		setAuthorizationHeader(cmd, req)

		resp, err := http.DefaultClient.Do(req.WithContext(cmd.Context()))
		if err != nil {
			return err
		}

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("failed to get key: %s", resp.Status)
		}

		_, err = io.Copy(os.Stdout, resp.Body)
		return err
	},
}

func init() {
	rootCmd.AddCommand(getCmd)

	getCmd.Flags().String(kaveFlagToken, "", "token to use for authorization")
}
