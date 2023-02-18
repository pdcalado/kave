package cmd

import (
	"fmt"
	"os"
	"path"

	"github.com/BurntSushi/toml"
	"github.com/spf13/cobra"

	"github.com/pdcalado/kave/internal/version"
)

const (
	// config related
	kaveConfigFile      = "config.toml"
	kaveConfigDir       = "kave"
	kaveCacheDir        = "kave"
	kaveCredentialsFile = "credentials.toml"
	kaveCacheTokenFile  = "token"

	// flags
	kaveFlagConfig            = "config"
	kaveFlagCredentials       = "credentials"
	kaveFlagUrl               = "url"
	kaveFlagRouterBasePath    = "router-base-path"
	kaveFlagToken             = "token"
	kaveFlagAuth0Audience     = "auth0_audience"
	kaveFlagAuth0Domain       = "auth0_domain"
	kaveFlagAuth0ClientID     = "auth0_client_id"
	kaveFlagAuth0ClientSecret = "auth0_client_secret"

	// env variables
	envAuth0Audience     = "AUTH0_AUDIENCE"
	envAuth0Domain       = "AUTH0_DOMAIN"
	envAuth0ClientID     = "AUTH0_CLIENT_ID"
	envAuth0ClientSecret = "AUTH0_CLIENT_SECRET"
)

var auth0ClientSecret string = ""

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "kave",
	Short:   "Get and Set key values on a kave server",
	Long:    "Make sure you run 'kave init' before using kave.",
	Version: version.Version,
	Run:     func(cmd *cobra.Command, args []string) {},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// Only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

type profileAuth struct {
	Enabled  bool   `toml:"enabled"`
	Domain   string `toml:"domain"`
	ClientID string `toml:"client_id"`
	Audience string `toml:"audience"`
}

type profile struct {
	Url            string      `toml:"url"`
	RouterBasePath string      `toml:"router_base_path"`
	Auth           profileAuth `toml:"auth"`
}

type profileCredentials struct {
	Secret string `toml:"secret"`
}

func init() {
	// load profile from config
	var cfgFile string
	rootCmd.PersistentFlags().StringVarP(&cfgFile, kaveFlagConfig, "c", "", "path to config file")

	if cfgFile == "" {
		var err error
		cfgFile, err = getConfigDefaultPath()
		if err != nil {
			panic(fmt.Errorf("failed to get config path: %w", err))
		}
	}

	prof, err := loadProfile(cfgFile)
	if err != nil {
		// profile not initialized
		prof = &profile{
			RouterBasePath: "/redis",
		}
	}

	// load credentials from credentials file
	var credentialsFile string
	rootCmd.PersistentFlags().StringVar(&credentialsFile, kaveFlagCredentials, "", "path to credentials file")
	if credentialsFile == "" {
		var err error
		credentialsFile, err = getCredentialsDefaultPath()
		if err != nil {
			panic(fmt.Errorf("failed to get credentials path: %w", err))
		}
	}

	creds, err := loadCredentials(credentialsFile)
	if err != nil {
		creds = &profileCredentials{}
	}

	rootCmd.PersistentFlags().StringVar(&prof.Url, kaveFlagUrl, prof.Url, "url of kave server")
	rootCmd.PersistentFlags().StringVar(&prof.RouterBasePath, kaveFlagRouterBasePath, prof.RouterBasePath, "base path of server router")
	rootCmd.PersistentFlags().StringVar(&prof.Auth.Audience, kaveFlagAuth0Audience, defaultEnvIfSet(envAuth0Audience, prof.Auth.Audience), "Auth0 audience setting")
	rootCmd.PersistentFlags().StringVar(&prof.Auth.Domain, kaveFlagAuth0Domain, defaultEnvIfSet(envAuth0Domain, prof.Auth.Domain), "Auth0 domain setting")
	rootCmd.PersistentFlags().StringVar(&prof.Auth.ClientID, kaveFlagAuth0ClientID, defaultEnvIfSet(envAuth0ClientID, prof.Auth.ClientID), "Auth0 client ID setting")

	// do not leak secret in usage
	var clientSecret string
	rootCmd.PersistentFlags().StringVar(&clientSecret, kaveFlagAuth0ClientSecret, "", "Auth0 client secret")
	if clientSecret == "" {
		var ok bool
		clientSecret, ok = os.LookupEnv(envAuth0ClientSecret)
		if !ok {
			clientSecret = creds.Secret
		}
	}
	auth0ClientSecret = clientSecret

	rootCmd.SilenceUsage = true

	rootCmd.AddCommand(initCmd)
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize kave user profile",
	Long:  "Initialize kave user profile. Specify url of kave server in the profile.",
	RunE: func(cmd *cobra.Command, args []string) error {
		formatFailedToInit := "failed to init profile: %w"

		configDir, err := os.UserConfigDir()
		if err != nil {
			return fmt.Errorf(formatFailedToInit, err)
		}

		profileDir := path.Join(configDir, kaveConfigDir)

		_, err = os.Stat(profileDir)
		if os.IsNotExist(err) {
			err = os.Mkdir(profileDir, os.ModePerm)
			if err != nil {
				return fmt.Errorf(formatFailedToInit, err)
			}
		}

		if err != nil {
			return fmt.Errorf(formatFailedToInit, err)
		}

		profilePath := path.Join(profileDir, kaveConfigFile)
		file, err := os.Create(profilePath)
		if err != nil {
			return fmt.Errorf(formatFailedToInit, err)
		}

		auth0Audience, _ := cmd.Flags().GetString(kaveFlagAuth0Audience)
		auth0Domain, _ := cmd.Flags().GetString(kaveFlagAuth0Domain)
		auth0ClientID, _ := cmd.Flags().GetString(kaveFlagAuth0ClientID)
		auth0ClientSecret, _ := cmd.Flags().GetString(kaveFlagAuth0ClientSecret)

		prof := &profile{
			Url:            cmd.Flags().Lookup(kaveFlagUrl).Value.String(),
			RouterBasePath: cmd.Flags().Lookup(kaveFlagRouterBasePath).Value.String(),
			Auth: profileAuth{
				Enabled:  auth0ClientID != "",
				Audience: auth0Audience,
				Domain:   auth0Domain,
				ClientID: auth0ClientID,
			},
		}

		if err := toml.NewEncoder(file).Encode(prof); err != nil {
			return fmt.Errorf(formatFailedToInit, err)
		}

		cmd.PrintErrf("config file written to %s\n", profilePath)

		if auth0ClientSecret == "" {
			return nil
		}

		credentialsPath := path.Join(profileDir, kaveCredentialsFile)
		file, err = os.Create(credentialsPath)
		if err != nil {
			return fmt.Errorf(formatFailedToInit, err)
		}

		credentials := profileCredentials{
			Secret: auth0ClientSecret,
		}

		if err := toml.NewEncoder(file).Encode(credentials); err != nil {
			return fmt.Errorf(formatFailedToInit, err)
		}

		cmd.PrintErrf("credentials file written to %s\n", credentialsPath)

		return nil
	},
}

func loadProfile(path string) (*profile, error) {
	prof := &profile{}
	_, err := toml.DecodeFile(path, prof)
	if err != nil {
		return nil, fmt.Errorf("failed to load profile: %w", err)
	}

	if prof.Url == "" {
		return nil, fmt.Errorf("%s is empty", kaveFlagUrl)
	}

	return prof, nil
}

func loadCredentials(path string) (*profileCredentials, error) {
	creds := &profileCredentials{}
	_, err := toml.DecodeFile(path, creds)
	if err != nil {
		return nil, fmt.Errorf("failed to load profile: %w", err)
	}

	return creds, nil
}

func getConfigDefaultPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return path.Join(configDir, kaveConfigDir, kaveConfigFile), nil
}

func getCredentialsDefaultPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return path.Join(configDir, kaveConfigDir, kaveCredentialsFile), nil
}

func defaultEnvIfSet(envVar string, def string) string {
	envValue, ok := os.LookupEnv(envVar)
	if !ok {
		return def
	}
	return envValue
}
