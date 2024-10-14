/*
 * Copyright FMR LLC <opensource@fidelity.com>
 *
 * SPDX-License-Identifier: Apache
 */
package oidcmethod

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/fidelity/theliv/internal/rbac"
	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/oauth2"

	"github.com/fidelity/theliv/pkg/config"
	log "github.com/fidelity/theliv/pkg/log"

	theliverr "github.com/fidelity/theliv/pkg/err"
)

var verifier *oidc.IDTokenVerifier

func SSO(r chi.Router) {
	r.Get("/callback", callback)
}

func callback(w http.ResponseWriter, r *http.Request) {
	if err := ExchangeToken(w, r); err != nil {
		processError(w, r, err)
		return
	}
<<<<<<< HEAD
	oicdConfig := config.GetThelivConfig().Oidc
	host := oicdConfig.CallBackHost
	// redirect, by default to /
	state := r.URL.Query().Get("state")
=======
	// redirect, by default to /
	state := r.URL.Query().Get("state")
	host := r.Referer()[:len(r.Referer())-1]
	log.S().Infof("Call back endpoint is %s", host)
>>>>>>> a2bcfd2 (feat: use oidc)
	http.Redirect(w, r, host+state, http.StatusFound)
}

func ExchangeToken(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	oauth2Token, err := getOauthConfig(r).Exchange(r.Context(), r.URL.Query().Get("code"))
	if err != nil {
		msg := "failed to exchange token"
		log.SWithContext(ctx).Error(msg)
		return theliverr.NewCommonError(ctx, 1, msg)
	}
	rawIDToken, ok := oauth2Token.Extra(IDTokenKey).(string)
	if !ok {
		msg := "No id_token field in oauth2 token."
		log.SWithContext(ctx).Error(msg)
		return theliverr.NewCommonError(ctx, 1, msg)
	}
	idtoken, err := verify(ctx, rawIDToken)
	if err != nil {
		msg := "failed to verify ID token"
		log.SWithContext(ctx).Error(msg)
		return theliverr.NewCommonError(ctx, 1, msg)
	}
	thelivUser := &rbac.ThelivUser{}
	if err := idtoken.Claims(thelivUser); err != nil {
		msg := "failed to unmarshal user from id_token"
		log.SWithContext(ctx).Error(msg)
		return theliverr.NewCommonError(ctx, 1, msg)
	}
	// user id
	user := rbac.UserInfo{}
	user.Email = thelivUser.Email
	if thelivUser.ExpiresAt != nil {
		user.ExpiresAt = thelivUser.ExpiresAt.UnixMilli()
	}
	user.FamilyName = thelivUser.Surname
	user.GivenName = thelivUser.GivenName
	user.Issuer = thelivUser.Issuer
	user.Name = thelivUser.DisplayName
	user.Upn = thelivUser.Upn
	user.Id = strings.ToLower(strings.Split(user.Upn, "@")[0])
	user.Roles = thelivUser.Groups

	// set access_token cookie
	acc := &http.Cookie{
		Path:     "/",
		Name:     AccessTokenKey,
		Value:    oauth2Token.AccessToken,
		Expires:  oauth2Token.Expiry,
		Secure:   r.TLS != nil,
		HttpOnly: true,
	}
	http.SetCookie(w, acc)

	return setUserCookie(r.Context(), user, w)
}

func getOauthConfig(r *http.Request) *oauth2.Config {
	ctx := r.Context()
	host := r.Host
	if oc, ok := oauthConfigs[host]; ok {
		return oc
	}
	oicdConfig := config.GetThelivConfig().Oidc
	log.SWithContext(ctx).Infof("oauth-config for host %s does not exist, create new one", host)
	// create new one
	callback := GetHost(r) + oicdConfig.CallBack
	oauthConfigs[host] = &oauth2.Config{
		ClientID:     oicdConfig.ClientID,
		ClientSecret: oicdConfig.ClientSecret,
		Endpoint:     provider.Endpoint(),
		RedirectURL:  callback,
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}
	log.SWithContext(ctx).Infof("successfully created new oauth config")
	return oauthConfigs[host]
}

func setUserCookie(ctx context.Context, user rbac.UserInfo, w http.ResponseWriter) error {

	userinfo, err := userJwt(ctx, &user)
	if err != nil {
		return err
	}
	//set id tokens
	cookies := splitCookie(IDTokenKey, userinfo)
	for _, ck := range cookies {
		http.SetCookie(w, ck)
	}
	return nil
}

func userJwt(ctx context.Context, user *rbac.UserInfo) (string, error) {

	// sign userinfo JWT
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, user)
	jwt, err := jwtToken.SignedString(jwtKey)
	if err != nil {
		msg := "Failed to sign jwt"
		log.SWithContext(ctx).Error()
		return "", theliverr.NewCommonError(ctx, 1, msg)
	}
	return jwt, nil
}

func GetHost(req *http.Request) string {
	proto := "https"
	if req.TLS == nil {
		proto = "http"
	}
	return fmt.Sprintf("%s://%s", proto, req.Host)
}

var verify = func(ctx context.Context, rawIDToken string) (*oidc.IDToken, error) {
	return verifier.Verify(ctx, rawIDToken)
}
