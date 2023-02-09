package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

const (
	envAuth0Audience     = "AUTH0_AUDIENCE"
	envAuth0Domain       = "AUTH0_DOMAIN"
	envAuth0ClientSecret = "AUTH0_CLIENT_SECRET"
)

// tokenCmd tries to obtain a token from the Auth provider
// Currently tailored for Auth0
var tokenCmd = &cobra.Command{
	Use:   "token <client-id>",
	Short: "Obtain a token from Auth0",
	Long: fmt.Sprintf(`Must provide Auth0 client-id by argument.
%s and %s are read from the environment.
Audience is assumed to be the configured kave API URL, unless %s is set.
`, envAuth0Domain, envAuth0ClientSecret, envAuth0Audience),
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		audience := cmd.Flag(kaveFlagUrl).Value.String()
		if aud, ok := os.LookupEnv(envAuth0Audience); ok {
			audience = aud
		}

		envMustBeSetFormat := "%s must be set"

		domain := os.Getenv(envAuth0Domain)
		if domain == "" {
			return fmt.Errorf(fmt.Sprintf(envMustBeSetFormat, envAuth0Domain))
		}

		domain = strings.TrimPrefix(domain, "https://")

		secret := os.Getenv(envAuth0ClientSecret)
		if secret == "" {
			return fmt.Errorf(fmt.Sprintf(envMustBeSetFormat, envAuth0ClientSecret))
		}

		urlStr := fmt.Sprintf("https://%s/oauth/token", domain)

		data := map[string]string{
			"client_id":     args[0],
			"client_secret": secret,
			"audience":      audience,
			"grant_type":    "client_credentials",
		}
		buf, _ := json.Marshal(data)

		req, err := http.NewRequest(http.MethodPost, urlStr, bytes.NewBuffer(buf))
		if err != nil {
			return err
		}

		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req.WithContext(cmd.Context()))
		if err != nil {
			return err
		}

		if resp.StatusCode/100 != 2 {
			return fmt.Errorf("failed to obtain token: %s", resp.Status)
		}

		tokenResponse := struct {
			AccessToken string `json:"access_token"`
		}{}

		err = json.NewDecoder(resp.Body).Decode(&tokenResponse)
		if err != nil {
			return err
		}

		_, err = os.Stdout.WriteString(tokenResponse.AccessToken)
		return err
	},
}

func init() {
	rootCmd.AddCommand(tokenCmd)
}
