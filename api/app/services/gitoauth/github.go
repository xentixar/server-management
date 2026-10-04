package gitoauth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"api/app/constants"
)

type githubProvider struct {
	clientID     string
	clientSecret string
}

func (p githubProvider) Name() string { return constants.GitProviderGitHub }

func (p githubProvider) AuthorizeURL(state, redirectURI string) (string, error) {
	q := url.Values{}
	q.Set("client_id", p.clientID)
	q.Set("redirect_uri", redirectURI)
	q.Set("scope", "read:user repo")
	q.Set("state", state)
	return "https://github.com/login/oauth/authorize?" + q.Encode(), nil
}

func (p githubProvider) Exchange(ctx context.Context, code, redirectURI string) (TokenSet, error) {
	form := url.Values{}
	form.Set("client_id", p.clientID)
	form.Set("client_secret", p.clientSecret)
	form.Set("code", code)
	form.Set("redirect_uri", redirectURI)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://github.com/login/oauth/access_token", strings.NewReader(form.Encode()))
	if err != nil {
		return TokenSet{}, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := httpClient().Do(req)
	if err != nil {
		return TokenSet{}, err
	}
	body, err := readBody(resp)
	if err != nil {
		return TokenSet{}, err
	}
	if resp.StatusCode >= 300 {
		return TokenSet{}, fmt.Errorf("github token exchange failed: %s", string(body))
	}
	var raw struct {
		AccessToken string `json:"access_token"`
		Scope       string `json:"scope"`
		TokenType   string `json:"token_type"`
		Error       string `json:"error"`
		ErrorDesc   string `json:"error_description"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return TokenSet{}, err
	}
	if raw.AccessToken == "" {
		if raw.Error != "" {
			return TokenSet{}, fmt.Errorf("github token exchange: %s (%s)", raw.Error, raw.ErrorDesc)
		}
		return TokenSet{}, fmt.Errorf("github token exchange: empty access token")
	}
	return TokenSet{AccessToken: raw.AccessToken, Scopes: raw.Scope}, nil
}

func (p githubProvider) Refresh(ctx context.Context, refreshToken string) (TokenSet, error) {
	return TokenSet{}, fmt.Errorf("github oauth apps do not support refresh tokens")
}

func (p githubProvider) Account(ctx context.Context, accessToken string) (Account, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/user", nil)
	if err != nil {
		return Account{}, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	resp, err := httpClient().Do(req)
	if err != nil {
		return Account{}, err
	}
	body, err := readBody(resp)
	if err != nil {
		return Account{}, err
	}
	if resp.StatusCode >= 300 {
		return Account{}, fmt.Errorf("github user failed: %s", string(body))
	}
	var raw struct {
		ID        int64  `json:"id"`
		Login     string `json:"login"`
		Name      string `json:"name"`
		AvatarURL string `json:"avatar_url"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return Account{}, err
	}
	return Account{
		ID:        strconv.FormatInt(raw.ID, 10),
		Login:     raw.Login,
		Name:      raw.Name,
		AvatarURL: raw.AvatarURL,
	}, nil
}

func (p githubProvider) ListRepos(ctx context.Context, accessToken, query string, page int) ([]Repo, error) {
	page = ensurePage(page)
	u := url.URL{Scheme: "https", Host: "api.github.com", Path: "/user/repos"}
	q := url.Values{}
	q.Set("per_page", "50")
	q.Set("page", strconv.Itoa(page))
	q.Set("sort", "updated")
	q.Set("affiliation", "owner,collaborator,organization_member")
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	resp, err := httpClient().Do(req)
	if err != nil {
		return nil, err
	}
	body, err := readBody(resp)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("github repos failed: %s", string(body))
	}
	var raw []struct {
		ID            int64  `json:"id"`
		Name          string `json:"name"`
		FullName      string `json:"full_name"`
		Private       bool   `json:"private"`
		DefaultBranch string `json:"default_branch"`
		CloneURL      string `json:"clone_url"`
		HTMLURL       string `json:"html_url"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, err
	}
	query = strings.ToLower(strings.TrimSpace(query))
	out := make([]Repo, 0, len(raw))
	for _, r := range raw {
		repo := Repo{
			ID:            strconv.FormatInt(r.ID, 10),
			Name:          r.Name,
			FullName:      r.FullName,
			Private:       r.Private,
			DefaultBranch: r.DefaultBranch,
			CloneURL:      r.CloneURL,
			HTMLURL:       r.HTMLURL,
			Provider:      constants.GitProviderGitHub,
		}
		if query != "" && !strings.Contains(strings.ToLower(repo.FullName), query) {
			continue
		}
		out = append(out, repo)
	}
	return out, nil
}
