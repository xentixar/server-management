package gitoauth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"api/app/constants"
)

type bitbucketProvider struct {
	clientID     string
	clientSecret string
}

func (p bitbucketProvider) Name() string { return constants.GitProviderBitbucket }

func (p bitbucketProvider) AuthorizeURL(state, redirectURI string) (string, error) {
	q := url.Values{}
	q.Set("client_id", p.clientID)
	q.Set("response_type", "code")
	q.Set("redirect_uri", redirectURI)
	q.Set("state", state)
	// Bitbucket scopes are configured on the consumer; request via query when supported.
	q.Set("scope", "repository account")
	return "https://bitbucket.org/site/oauth2/authorize?" + q.Encode(), nil
}

func (p bitbucketProvider) Exchange(ctx context.Context, code, redirectURI string) (TokenSet, error) {
	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("code", code)
	form.Set("redirect_uri", redirectURI)
	return p.tokenRequest(ctx, form)
}

func (p bitbucketProvider) Refresh(ctx context.Context, refreshToken string) (TokenSet, error) {
	if refreshToken == "" {
		return TokenSet{}, fmt.Errorf("bitbucket refresh token missing")
	}
	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", refreshToken)
	return p.tokenRequest(ctx, form)
}

func (p bitbucketProvider) tokenRequest(ctx context.Context, form url.Values) (TokenSet, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://bitbucket.org/site/oauth2/access_token", strings.NewReader(form.Encode()))
	if err != nil {
		return TokenSet{}, err
	}
	req.SetBasicAuth(p.clientID, p.clientSecret)
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
		return TokenSet{}, fmt.Errorf("bitbucket token failed: %s", string(body))
	}
	var raw struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
		Scopes       string `json:"scopes"`
		Error        string `json:"error"`
		ErrorDesc    string `json:"error_description"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return TokenSet{}, err
	}
	if raw.AccessToken == "" {
		if raw.Error != "" {
			return TokenSet{}, fmt.Errorf("bitbucket token: %s (%s)", raw.Error, raw.ErrorDesc)
		}
		return TokenSet{}, fmt.Errorf("bitbucket token: empty access token")
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
		Scopes:       raw.Scopes,
	}, nil
}

func (p bitbucketProvider) Account(ctx context.Context, accessToken string) (Account, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.bitbucket.org/2.0/user", nil)
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
		return Account{}, fmt.Errorf("bitbucket user failed: %s", string(body))
	}
	var raw struct {
		UUID        string `json:"uuid"`
		Username    string `json:"username"`
		DisplayName string `json:"display_name"`
		Links       struct {
			Avatar struct {
				Href string `json:"href"`
			} `json:"avatar"`
		} `json:"links"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return Account{}, err
	}
	return Account{
		ID:        raw.UUID,
		Login:     raw.Username,
		Name:      raw.DisplayName,
		AvatarURL: raw.Links.Avatar.Href,
	}, nil
}

func (p bitbucketProvider) ListRepos(ctx context.Context, accessToken, query string, page int) ([]Repo, error) {
	page = ensurePage(page)
	u := url.URL{Scheme: "https", Host: "api.bitbucket.org", Path: "/2.0/repositories"}
	q := url.Values{}
	q.Set("role", "member")
	q.Set("pagelen", "50")
	q.Set("page", fmt.Sprintf("%d", page))
	if strings.TrimSpace(query) != "" {
		q.Set("q", fmt.Sprintf(`name~"%s"`, strings.ReplaceAll(strings.TrimSpace(query), `"`, ``)))
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
		return nil, fmt.Errorf("bitbucket repos failed: %s", string(body))
	}
	var raw struct {
		Values []struct {
			UUID       string `json:"uuid"`
			Name       string `json:"name"`
			FullName   string `json:"full_name"`
			IsPrivate  bool   `json:"is_private"`
			MainBranch *struct {
				Name string `json:"name"`
			} `json:"mainbranch"`
			Links struct {
				HTML struct {
					Href string `json:"href"`
				} `json:"html"`
				Clone []struct {
					Name string `json:"name"`
					Href string `json:"href"`
				} `json:"clone"`
			} `json:"links"`
		} `json:"values"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, err
	}
	out := make([]Repo, 0, len(raw.Values))
	for _, r := range raw.Values {
		cloneURL := ""
		for _, c := range r.Links.Clone {
			if c.Name == "https" {
				cloneURL = c.Href
				break
			}
		}
		branch := "main"
		if r.MainBranch != nil && r.MainBranch.Name != "" {
			branch = r.MainBranch.Name
		}
		out = append(out, Repo{
			ID:            r.UUID,
			Name:          r.Name,
			FullName:      r.FullName,
			Private:       r.IsPrivate,
			DefaultBranch: branch,
			CloneURL:      cloneURL,
			HTMLURL:       r.Links.HTML.Href,
			Provider:      constants.GitProviderBitbucket,
		})
	}
	return out, nil
}
