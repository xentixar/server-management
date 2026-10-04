package gitoauth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"api/app/constants"
)

type gitlabProvider struct {
	clientID     string
	clientSecret string
}

func (p gitlabProvider) Name() string { return constants.GitProviderGitLab }

func (p gitlabProvider) AuthorizeURL(state, redirectURI string) (string, error) {
	q := url.Values{}
	q.Set("client_id", p.clientID)
	q.Set("redirect_uri", redirectURI)
	q.Set("response_type", "code")
	q.Set("scope", "read_api")
	q.Set("state", state)
	return "https://gitlab.com/oauth/authorize?" + q.Encode(), nil
}

func (p gitlabProvider) Exchange(ctx context.Context, code, redirectURI string) (TokenSet, error) {
	return p.tokenRequest(ctx, url.Values{
		"client_id":     {p.clientID},
		"client_secret": {p.clientSecret},
		"code":          {code},
		"grant_type":    {"authorization_code"},
		"redirect_uri":  {redirectURI},
	})
}

func (p gitlabProvider) Refresh(ctx context.Context, refreshToken string) (TokenSet, error) {
	if refreshToken == "" {
		return TokenSet{}, fmt.Errorf("gitlab refresh token missing")
	}
	return p.tokenRequest(ctx, url.Values{
		"client_id":     {p.clientID},
		"client_secret": {p.clientSecret},
		"refresh_token": {refreshToken},
		"grant_type":    {"refresh_token"},
	})
}

func (p gitlabProvider) tokenRequest(ctx context.Context, form url.Values) (TokenSet, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://gitlab.com/oauth/token", strings.NewReader(form.Encode()))
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
		return TokenSet{}, fmt.Errorf("gitlab token failed: %s", string(body))
	}
	var raw struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
		Scope        string `json:"scope"`
		Error        string `json:"error"`
		ErrorDesc    string `json:"error_description"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return TokenSet{}, err
	}
	if raw.AccessToken == "" {
		if raw.Error != "" {
			return TokenSet{}, fmt.Errorf("gitlab token: %s (%s)", raw.Error, raw.ErrorDesc)
		}
		return TokenSet{}, fmt.Errorf("gitlab token: empty access token")
	}
	var expires *time.Time
	if raw.ExpiresIn > 0 {
		t := time.Now().UTC().Add(time.Duration(raw.ExpiresIn) * time.Second)
		expires = &t
	}
	return TokenSet{
		AccessToken:  raw.AccessToken,
		RefreshToken: raw.RefreshToken,
		ExpiresAt:    expires,
		Scopes:       raw.Scope,
	}, nil
}

func (p gitlabProvider) Account(ctx context.Context, accessToken string) (Account, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://gitlab.com/api/v4/user", nil)
	if err != nil {
		return Account{}, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := httpClient().Do(req)
	if err != nil {
		return Account{}, err
	}
	body, err := readBody(resp)
	if err != nil {
		return Account{}, err
	}
	if resp.StatusCode >= 300 {
		return Account{}, fmt.Errorf("gitlab user failed: %s", string(body))
	}
	var raw struct {
		ID        int64  `json:"id"`
		Username  string `json:"username"`
		Name      string `json:"name"`
		AvatarURL string `json:"avatar_url"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return Account{}, err
	}
	return Account{
		ID:        strconv.FormatInt(raw.ID, 10),
		Login:     raw.Username,
		Name:      raw.Name,
		AvatarURL: raw.AvatarURL,
	}, nil
}

func (p gitlabProvider) ListRepos(ctx context.Context, accessToken, query string, page int) ([]Repo, error) {
	page = ensurePage(page)
	u := url.URL{Scheme: "https", Host: "gitlab.com", Path: "/api/v4/projects"}
	q := url.Values{}
	q.Set("membership", "true")
	q.Set("simple", "true")
	q.Set("order_by", "last_activity_at")
	q.Set("per_page", "50")
	q.Set("page", strconv.Itoa(page))
	if strings.TrimSpace(query) != "" {
		q.Set("search", strings.TrimSpace(query))
	}
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := httpClient().Do(req)
	if err != nil {
		return nil, err
	}
	body, err := readBody(resp)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("gitlab repos failed: %s", string(body))
	}
	var raw []struct {
		ID                int64  `json:"id"`
		Name              string `json:"name"`
		PathWithNamespace string `json:"path_with_namespace"`
		Visibility        string `json:"visibility"`
		DefaultBranch     string `json:"default_branch"`
		HTTPURLToRepo     string `json:"http_url_to_repo"`
		WebURL            string `json:"web_url"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, err
	}
	out := make([]Repo, 0, len(raw))
	for _, r := range raw {
		out = append(out, Repo{
			ID:            strconv.FormatInt(r.ID, 10),
			Name:          r.Name,
			FullName:      r.PathWithNamespace,
			Private:       r.Visibility != "public",
			DefaultBranch: r.DefaultBranch,
			CloneURL:      r.HTTPURLToRepo,
			HTMLURL:       r.WebURL,
			Provider:      constants.GitProviderGitLab,
		})
	}
	return out, nil
}
