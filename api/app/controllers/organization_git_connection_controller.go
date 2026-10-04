package controllers

import (
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"api/app/constants"
	"api/app/helpers"
	"api/app/models"
	"api/app/policies"
	"api/app/resources"
	"api/app/services/gitoauth"
	"api/config"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type OrganizationGitConnectionController struct{}

func (OrganizationGitConnectionController) List(ctx *gin.Context) {
	organizationID, ok := uintParam(ctx, "organization", "Invalid organization ID")
	if !ok {
		return
	}
	_, member, ok := requireOrgMember(ctx, organizationID)
	if !ok {
		return
	}
	if !denyUnless(ctx, (policies.OrganizationPolicy{}).ViewGitConnections(member)) {
		return
	}

	rows, err := (models.OrganizationGitConnection{}).ListByOrganization(ctx, organizationID)
	if err != nil {
		respond(ctx).fail("Failed to list git connections", gin.H{}, http.StatusInternalServerError)
		return
	}
	respond(ctx).ok("Git connections fetched", resources.OrganizationGitConnectionCollection(rows), http.StatusOK)
}

func (OrganizationGitConnectionController) Authorize(ctx *gin.Context) {
	organizationID, ok := uintParam(ctx, "organization", "Invalid organization ID")
	if !ok {
		return
	}
	_, member, ok := requireOrgMember(ctx, organizationID)
	if !ok {
		return
	}
	if !denyUnless(ctx, (policies.OrganizationPolicy{}).ManageGitConnections(member)) {
		return
	}

	provider := strings.TrimSpace(strings.ToLower(ctx.Param("provider")))
	if !constants.IsValidGitProvider(provider) {
		respond(ctx).fail("Invalid provider", gin.H{}, http.StatusBadRequest)
		return
	}
	client, err := gitoauth.Get(provider)
	if err != nil {
		if errors.Is(err, gitoauth.ErrProviderNotConfigured) {
			respond(ctx).fail("Git provider is not configured on this server", gin.H{}, http.StatusServiceUnavailable)
			return
		}
		respond(ctx).fail("Invalid provider", gin.H{}, http.StatusBadRequest)
		return
	}
	redirectURI, err := gitoauth.CallbackURL(provider)
	if err != nil {
		respond(ctx).fail(err.Error(), gin.H{}, http.StatusServiceUnavailable)
		return
	}

	stateToken, err := helpers.GenerateToken()
	if err != nil {
		respond(ctx).fail("Failed to start OAuth", gin.H{}, http.StatusInternalServerError)
		return
	}
	_ = (models.GitOAuthState{}).PurgeExpired(ctx)
	state := models.GitOAuthState{
		State:          stateToken,
		OrganizationID: organizationID,
		UserID:         currentUser(ctx).ID,
		Provider:       provider,
		ExpiresAt:      time.Now().UTC().Add(10 * time.Minute),
	}
	if err := state.Create(ctx); err != nil {
		respond(ctx).fail("Failed to start OAuth", gin.H{}, http.StatusInternalServerError)
		return
	}

	authorizeURL, err := client.AuthorizeURL(stateToken, redirectURI)
	if err != nil {
		respond(ctx).fail("Failed to build authorize URL", gin.H{}, http.StatusInternalServerError)
		return
	}
	respond(ctx).ok("Authorize URL ready", gin.H{"authorize_url": authorizeURL}, http.StatusOK)
}

func (OrganizationGitConnectionController) Callback(ctx *gin.Context) {
	provider := strings.TrimSpace(strings.ToLower(ctx.Param("provider")))
	if !constants.IsValidGitProvider(provider) {
		gitOAuthRedirect(ctx, "", provider, "invalid_provider")
		return
	}
	code := strings.TrimSpace(ctx.Query("code"))
	stateToken := strings.TrimSpace(ctx.Query("state"))
	if errMsg := strings.TrimSpace(ctx.Query("error")); errMsg != "" {
		gitOAuthRedirect(ctx, "", provider, errMsg)
		return
	}
	if code == "" || stateToken == "" {
		gitOAuthRedirect(ctx, "", provider, "missing_code")
		return
	}

	state, err := (models.GitOAuthState{}).Consume(ctx, stateToken, provider)
	if err != nil {
		gitOAuthRedirect(ctx, "", provider, "invalid_state")
		return
	}

	org, err := (models.Organization{}).FindByID(ctx, state.OrganizationID)
	if err != nil {
		gitOAuthRedirect(ctx, "", provider, "organization_not_found")
		return
	}

	client, err := gitoauth.Get(provider)
	if err != nil {
		gitOAuthRedirect(ctx, org.Slug, provider, "provider_not_configured")
		return
	}
	redirectURI, err := gitoauth.CallbackURL(provider)
	if err != nil {
		gitOAuthRedirect(ctx, org.Slug, provider, "misconfigured")
		return
	}

	tokens, err := client.Exchange(ctx, code, redirectURI)
	if err != nil {
		gitOAuthRedirect(ctx, org.Slug, provider, "token_exchange_failed")
		return
	}
	account, err := client.Account(ctx, tokens.AccessToken)
	if err != nil {
		gitOAuthRedirect(ctx, org.Slug, provider, "account_fetch_failed")
		return
	}

	encKey, err := helpers.DecodeEncryptionKey(config.AppEncryptionKey())
	if err != nil {
		gitOAuthRedirect(ctx, org.Slug, provider, "encryption_not_configured")
		return
	}
	accessEnc, err := helpers.EncryptAESGCM(encKey, tokens.AccessToken)
	if err != nil {
		gitOAuthRedirect(ctx, org.Slug, provider, "encrypt_failed")
		return
	}
	refreshEnc := ""
	if tokens.RefreshToken != "" {
		refreshEnc, err = helpers.EncryptAESGCM(encKey, tokens.RefreshToken)
		if err != nil {
			gitOAuthRedirect(ctx, org.Slug, provider, "encrypt_failed")
			return
		}
	}

	existing, err := (models.OrganizationGitConnection{}).FindByProviderAccount(ctx, org.ID, provider, account.ID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		gitOAuthRedirect(ctx, org.Slug, provider, "persist_failed")
		return
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		conn := models.OrganizationGitConnection{
			OrganizationID:    org.ID,
			Provider:          provider,
			ExternalAccountID: account.ID,
			AccountLogin:      account.Login,
			AccountName:       account.Name,
			AvatarURL:         account.AvatarURL,
			AccessTokenEnc:    accessEnc,
			RefreshTokenEnc:   refreshEnc,
			TokenExpiresAt:    tokens.ExpiresAt,
			Scopes:            tokens.Scopes,
			Status:            constants.GitConnectionStatusActive,
			LastError:         "",
		}
		if err := conn.Create(ctx); err != nil {
			gitOAuthRedirect(ctx, org.Slug, provider, "persist_failed")
			return
		}
	} else {
		existing.AccountLogin = account.Login
		existing.AccountName = account.Name
		existing.AvatarURL = account.AvatarURL
		existing.AccessTokenEnc = accessEnc
		if refreshEnc != "" {
			existing.RefreshTokenEnc = refreshEnc
		}
		existing.TokenExpiresAt = tokens.ExpiresAt
		existing.Scopes = tokens.Scopes
		existing.Status = constants.GitConnectionStatusActive
		existing.LastError = ""
		if err := existing.Save(ctx); err != nil {
			gitOAuthRedirect(ctx, org.Slug, provider, "persist_failed")
			return
		}
	}

	gitOAuthRedirect(ctx, org.Slug, provider, "")
}

func (OrganizationGitConnectionController) Delete(ctx *gin.Context) {
	organizationID, ok := uintParam(ctx, "organization", "Invalid organization ID")
	if !ok {
		return
	}
	_, member, ok := requireOrgMember(ctx, organizationID)
	if !ok {
		return
	}
	if !denyUnless(ctx, (policies.OrganizationPolicy{}).ManageGitConnections(member)) {
		return
	}

	conn, err := (models.OrganizationGitConnection{}).FindByPublicID(ctx, ctx.Param("connection"), organizationID)
	if err != nil {
		respond(ctx).fail("Git connection not found", gin.H{}, http.StatusNotFound)
		return
	}
	if err := conn.Delete(ctx); err != nil {
		respond(ctx).fail("Failed to disconnect", gin.H{}, http.StatusInternalServerError)
		return
	}
	respond(ctx).ok("Git connection removed", gin.H{}, http.StatusOK)
}

func (OrganizationGitConnectionController) ListRepos(ctx *gin.Context) {
	organizationID, ok := uintParam(ctx, "organization", "Invalid organization ID")
	if !ok {
		return
	}
	_, member, ok := requireOrgMember(ctx, organizationID)
	if !ok {
		return
	}
	if !denyUnless(ctx, (policies.OrganizationPolicy{}).ViewGitConnections(member)) {
		return
	}

	conn, err := (models.OrganizationGitConnection{}).FindByPublicID(ctx, ctx.Param("connection"), organizationID)
	if err != nil {
		respond(ctx).fail("Git connection not found", gin.H{}, http.StatusNotFound)
		return
	}

	accessToken, err := ensureAccessToken(ctx, &conn)
	if err != nil {
		respond(ctx).fail("Git connection needs re-authorization", gin.H{
			"connection": resources.OrganizationGitConnectionResource(conn),
		}, http.StatusBadGateway)
		return
	}

	client, err := gitoauth.Get(conn.Provider)
	if err != nil {
		respond(ctx).fail("Git provider is not configured on this server", gin.H{}, http.StatusServiceUnavailable)
		return
	}

	page := 1
	if raw := strings.TrimSpace(ctx.Query("page")); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n > 0 {
			page = n
		}
	}
	query := strings.TrimSpace(ctx.Query("q"))
	repos, err := client.ListRepos(ctx, accessToken, query, page)
	if err != nil {
		conn.Status = constants.GitConnectionStatusError
		conn.LastError = err.Error()
		_ = conn.Save(ctx)
		respond(ctx).fail("Failed to list repositories", gin.H{
			"connection": resources.OrganizationGitConnectionResource(conn),
		}, http.StatusBadGateway)
		return
	}

	if conn.Status != constants.GitConnectionStatusActive || conn.LastError != "" {
		conn.Status = constants.GitConnectionStatusActive
		conn.LastError = ""
		_ = conn.Save(ctx)
	}

	respond(ctx).ok("Repositories fetched", gin.H{
		"repos":      repos,
		"page":       page,
		"connection": resources.OrganizationGitConnectionResource(conn),
	}, http.StatusOK)
}

func ensureAccessToken(ctx *gin.Context, conn *models.OrganizationGitConnection) (string, error) {
	encKey, err := helpers.DecodeEncryptionKey(config.AppEncryptionKey())
	if err != nil {
		return "", err
	}
	accessToken, err := helpers.DecryptAESGCM(encKey, conn.AccessTokenEnc)
	if err != nil {
		return "", err
	}

	needsRefresh := conn.TokenExpiresAt != nil && time.Now().UTC().After(conn.TokenExpiresAt.Add(-2*time.Minute))
	if !needsRefresh || conn.RefreshTokenEnc == "" {
		return accessToken, nil
	}

	client, err := gitoauth.Get(conn.Provider)
	if err != nil {
		return "", err
	}
	refreshToken, err := helpers.DecryptAESGCM(encKey, conn.RefreshTokenEnc)
	if err != nil {
		return "", err
	}
	tokens, err := client.Refresh(ctx, refreshToken)
	if err != nil {
		conn.Status = constants.GitConnectionStatusError
		conn.LastError = err.Error()
		_ = conn.Save(ctx)
		return "", err
	}
	accessEnc, err := helpers.EncryptAESGCM(encKey, tokens.AccessToken)
	if err != nil {
		return "", err
	}
	conn.AccessTokenEnc = accessEnc
	if tokens.RefreshToken != "" {
		refreshEnc, err := helpers.EncryptAESGCM(encKey, tokens.RefreshToken)
		if err != nil {
			return "", err
		}
		conn.RefreshTokenEnc = refreshEnc
	}
	conn.TokenExpiresAt = tokens.ExpiresAt
	if tokens.Scopes != "" {
		conn.Scopes = tokens.Scopes
	}
	conn.Status = constants.GitConnectionStatusActive
	conn.LastError = ""
	if err := conn.Save(ctx); err != nil {
		return "", err
	}
	return tokens.AccessToken, nil
}

func gitOAuthRedirect(ctx *gin.Context, orgSlug, provider, errCode string) {
	base := strings.TrimRight(config.PublicDashboardURL(), "/")
	if base == "" {
		respond(ctx).fail("PUBLIC_DASHBOARD_URL is not configured", gin.H{"error": errCode}, http.StatusServiceUnavailable)
		return
	}
	if orgSlug == "" {
		ctx.Redirect(http.StatusFound, base+"/?git_oauth_error="+url.QueryEscape(errCode))
		return
	}
	target := base + "/o/" + url.PathEscape(orgSlug) + "/organization/git"
	q := url.Values{}
	if errCode != "" {
		q.Set("error", errCode)
	} else {
		q.Set("connected", provider)
	}
	ctx.Redirect(http.StatusFound, target+"?"+q.Encode())
}
