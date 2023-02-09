package cmd

import (
	"fmt"
	"os"
	"os/user"
	"path"

	"github.com/BurntSushi/toml"
	"github.com/pdcalado/kave/internal/version"
	"github.com/spf13/cobra"
)

const (
	kaveConfigFile         = "config.toml"
	kaveConfigDir          = ".kave"
	kaveFlagConfig         = "config"
	kaveFlagUsername       = "username"
	kaveFlagUrl            = "url"
	kaveFlagRouterBasePath = "router-base-path"
	kaveFlagToken          = "token"
)

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

type profile struct {
	Username       string `toml:"username"`
	Url            string `toml:"url"`
	RouterBasePath string `toml:"router_base_path"`
	Auth           struct {
		Enabled  bool   `toml:"enabled"`
		Domain   string `toml:"domain"`
		ClientID string `toml:"client_id"`
	} `toml:"auth"`
}

func init() {
	var cfgFile string
	rootCmd.PersistentFlags().StringVar(&cfgFile, kaveFlagConfig, "", "path to config file")

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

	rootCmd.PersistentFlags().StringVar(&prof.Username, kaveFlagUsername, prof.Username, "username for authentication")
	rootCmd.PersistentFlags().StringVar(&prof.Url, kaveFlagUrl, prof.Url, "url of kave server")
	rootCmd.PersistentFlags().StringVar(&prof.RouterBasePath, kaveFlagRouterBasePath, prof.RouterBasePath, "base path of server router")

	rootCmd.SilenceUsage = true

	rootCmd.AddCommand(initCmd)
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize kave user profile",
	Long:  "Specify username and url of kave server in the profile.",
	RunE: func(cmd *cobra.Command, args []string) error {
		formatFailedToInit := "failed to init profile: %w"

		home, err := getHomeDir()
		if err != nil {
			return fmt.Errorf(formatFailedToInit, err)
		}

		profileDir := path.Join(home, kaveConfigDir)

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

		prof := &profile{
			Username:       cmd.Flags().Lookup(kaveFlagUsername).Value.String(),
			Url:            cmd.Flags().Lookup(kaveFlagUrl).Value.String(),
			RouterBasePath: cmd.Flags().Lookup(kaveFlagRouterBasePath).Value.String(),
		}

		if err := toml.NewEncoder(file).Encode(prof); err != nil {
			return fmt.Errorf(formatFailedToInit, err)
		}
		return nil
	},
}

func loadProfile(path string) (*profile, error) {
	prof := &profile{}
	_, err := toml.DecodeFile(path, prof)
	if err != nil {
		return nil, fmt.Errorf("failed to load profile: %w", err)
	}

	if prof.Username == "" {
		return nil, fmt.Errorf("%s is empty", kaveFlagUsername)
	}

	if prof.Url == "" {
		return nil, fmt.Errorf("%s is empty", kaveFlagUrl)
	}

	return prof, nil
}

func getConfigDefaultPath() (string, error) {
	home, err := getHomeDir()
	if err != nil {
		return "", err
	}
	return path.Join(home, kaveConfigDir, kaveConfigFile), nil
}

func getHomeDir() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	return usr.HomeDir, nil
}
