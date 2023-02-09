package cmd

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
)

func createRequestUrl(cmd *cobra.Command, key string) (*url.URL, error) {
	urlStr := cmd.Flag(kaveFlagUrl).Value.String()
	basePath := cmd.Flag(kaveFlagRouterBasePath).Value.String()

	if !strings.HasPrefix(urlStr, "http") || !strings.HasPrefix(urlStr, "https") {
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

func setAuthorizationHeader(cmd *cobra.Command, req *http.Request) {
	if cmd.Flag(kaveFlagToken).Changed {
		req.Header.Set("Authorization", "Bearer "+cmd.Flag(kaveFlagToken).Value.String())
	}
}
