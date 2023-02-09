package cmd

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/spf13/cobra"
)

// setCmd set a key value in a kave server
var setCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a value for a key in a kave server",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]

		u, err := createRequestUrl(cmd, key)
		if err != nil {
			return err
		}

		body := bytes.NewBuffer([]byte(args[1]))

		req, err := http.NewRequest(http.MethodPost, u.String(), body)
		if err != nil {
			return err
		}

		setAuthorizationHeader(cmd, req)

		resp, err := http.DefaultClient.Do(req.WithContext(cmd.Context()))
		if err != nil {
			return err
		}

		if resp.StatusCode/100 != 2 {
			return fmt.Errorf("failed to set key value: %s", resp.Status)
		}

		_, err = io.Copy(os.Stdout, resp.Body)
		return err
	},
}

func init() {
	rootCmd.AddCommand(setCmd)

	setCmd.Flags().String(kaveFlagToken, "", "token to use for authorization")
}
