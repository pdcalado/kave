package cmd

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

func createRequestUrl(cmd *cobra.Command, key string) (*url.URL, error) {
	urlStr := cmd.Flag(kaveFlagUrl).Value.String()
	basePath := cmd.Flag(kaveFlagRouterBasePath).Value.String()

	if !strings.HasPrefix(urlStr, "http") && !strings.HasPrefix(urlStr, "https") {
		urlStr = "http://" + urlStr
	}

	u, err := url.Parse(fmt.Sprintf("%s%s/%s", urlStr, basePath, url.PathEscape(key)))
	if err != nil {
		return nil, err
	}

	if u.Scheme == "" {
		u.Scheme = "http"
	}

	return u, nil
}

// obtainRefreshToken reads token from cli, or from cache
// If expired, obtains new token from provider and writes to cache
func obtainRefreshToken(cmd *cobra.Command) (string, error) {
	if cmd.Flag(kaveFlagToken).Changed {
		return cmd.Flag(kaveFlagToken).Value.String(), nil
	}

	token, ok := readTokenFromCache()
	if ok {
		return token, nil
	}

	token, err := obtainAccessToken(cmd)
	if err != nil {
		return "", err
	}

	return token, writeTokenToCache(token)
}

func isAuthEnabled(cmd *cobra.Command) bool {
	return cmd.Flag(kaveFlagAuth0ClientID).Value.String() != ""
}

func setAuthorizationHeader(cmd *cobra.Command, req *http.Request) {
	if !isAuthEnabled(cmd) {
		return
	}

	token, err := obtainRefreshToken(cmd)
	if err != nil {
		panic("failed to obtain token: " + err.Error())
	}

	req.Header.Set("Authorization", "Bearer "+token)
}

func getTokenCachePath() (string, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}

	return path.Join(cacheDir, kaveCacheDir, kaveCacheTokenFile), nil
}

// readTokenFromCache reads token from cache file and validates it.
// 15 minutes to expiration considers the token invalid
func readTokenFromCache() (string, bool) {
	tokenCachePath, err := getTokenCachePath()
	if err != nil {
		return "", false
	}

	file, err := os.Open(tokenCachePath)
	if err != nil {
		return "", false
	}
	defer file.Close()

	buf, err := io.ReadAll(file)
	if err != nil {
		return "", false
	}

	token := strings.TrimSpace(string(buf))

	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return "", false
	}

	jbuf, err := base64.RawStdEncoding.DecodeString(parts[1])
	if err != nil {
		return "", false
	}

	tmp := struct {
		Exp int64 `json:"exp"`
	}{}
	err = json.Unmarshal(jbuf, &tmp)
	if err != nil {
		return "", false
	}

	return token, tmp.Exp > time.Now().Unix()+15*60
}

func writeTokenToCache(token string) error {
	tokenCachePath, err := getTokenCachePath()
	if err != nil {
		return err
	}

	file, err := os.Create(tokenCachePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// set proper permissions
	err = file.Chmod(0600)
	if err != nil {
		return err
	}

	_, err = io.WriteString(file, token)
	return err
}

func obtainAccessToken(cmd *cobra.Command) (string, error) {
	audience := cmd.Flag(kaveFlagAuth0Audience).Value.String()
	domain := cmd.Flag(kaveFlagAuth0Domain).Value.String()
	domain = strings.TrimPrefix(domain, "https://")
	clientID := cmd.Flag(kaveFlagAuth0ClientID).Value.String()

	secret := auth0ClientSecret
	if secret == "" {
		return "", fmt.Errorf("empty auth0 client secret")
	}

	urlStr := fmt.Sprintf("https://%s/oauth/token", domain)

	data := map[string]string{
		"client_id":     clientID,
		"client_secret": secret,
		"audience":      audience,
		"grant_type":    "client_credentials",
	}
	buf, _ := json.Marshal(data)

	req, err := http.NewRequest(http.MethodPost, urlStr, bytes.NewBuffer(buf))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req.WithContext(cmd.Context()))
	if err != nil {
		return "", err
	}

	if resp.StatusCode/100 != 2 {
		return "", fmt.Errorf("failed to obtain token: %s", resp.Status)
	}

	tokenResponse := struct {
		AccessToken string `json:"access_token"`
	}{}

	err = json.NewDecoder(resp.Body).Decode(&tokenResponse)
	if err != nil {
		return "", err
	}

	return tokenResponse.AccessToken, nil
}
