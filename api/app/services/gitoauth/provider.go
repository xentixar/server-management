package gitoauth

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"api/app/constants"
	"api/config"
)

var (
	ErrProviderNotConfigured = errors.New("git provider is not configured")
	ErrUnknownProvider       = errors.New("unknown git provider")
)

type TokenSet struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    *time.Time
	Scopes       string
}

type Account struct {
	ID        string
	Login     string
	Name      string
	AvatarURL string
}

type Repo struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	FullName      string `json:"full_name"`
	Private       bool   `json:"private"`
	DefaultBranch string `json:"default_branch"`
	CloneURL      string `json:"clone_url"`
	HTMLURL       string `json:"html_url"`
	Provider      string `json:"provider"`
}

type Provider interface {
	Name() string
	AuthorizeURL(state, redirectURI string) (string, error)
	Exchange(ctx context.Context, code, redirectURI string) (TokenSet, error)
	Refresh(ctx context.Context, refreshToken string) (TokenSet, error)
	Account(ctx context.Context, accessToken string) (Account, error)
	ListRepos(ctx context.Context, accessToken, query string, page int) ([]Repo, error)
}

func Get(provider string) (Provider, error) {
	switch strings.TrimSpace(strings.ToLower(provider)) {
	case constants.GitProviderGitHub:
		if config.GitHubOAuthClientID() == "" || config.GitHubOAuthClientSecret() == "" {
			return nil, ErrProviderNotConfigured
		}
		return githubProvider{
			clientID:     config.GitHubOAuthClientID(),
			clientSecret: config.GitHubOAuthClientSecret(),
		}, nil
	case constants.GitProviderGitLab:
		if config.GitLabOAuthClientID() == "" || config.GitLabOAuthClientSecret() == "" {
			return nil, ErrProviderNotConfigured
		}
		return gitlabProvider{
			clientID:     config.GitLabOAuthClientID(),
			clientSecret: config.GitLabOAuthClientSecret(),
		}, nil
	case constants.GitProviderBitbucket:
		if config.BitbucketOAuthClientID() == "" || config.BitbucketOAuthClientSecret() == "" {
			return nil, ErrProviderNotConfigured
		}
		return bitbucketProvider{
			clientID:     config.BitbucketOAuthClientID(),
			clientSecret: config.BitbucketOAuthClientSecret(),
		}, nil
	default:
		return nil, ErrUnknownProvider
	}
}

func CallbackURL(provider string) (string, error) {
	base := strings.TrimRight(config.PublicAPIURL(), "/")
	if base == "" {
		return "", errors.New("PUBLIC_API_URL is not configured")
	}
	if !constants.IsValidGitProvider(provider) {
		return "", ErrUnknownProvider
	}
	return fmt.Sprintf("%s/api/v1/git/oauth/callback/%s", base, provider), nil
}

func httpClient() *http.Client {
	return &http.Client{Timeout: 20 * time.Second}
}

func readBody(resp *http.Response) ([]byte, error) {
	defer resp.Body.Close()
	return io.ReadAll(io.LimitReader(resp.Body, 4<<20))
}

func ensurePage(page int) int {
	if page < 1 {
		return 1
	}
	return page
}
