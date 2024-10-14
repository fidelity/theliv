/*
 * Copyright FMR LLC <opensource@fidelity.com>
 *
 * SPDX-License-Identifier: Apache
 */
package oidcmethod

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/fidelity/theliv/internal/rbac"
	"github.com/fidelity/theliv/pkg/config"
	log "github.com/fidelity/theliv/pkg/log"
	"github.com/go-chi/render"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/oauth2"
)

const (
	AccessTokenKey string = "access_token"
	IDTokenKey     string = "id_token"
)

var provider *oidc.Provider

var oauthConfigs = make(map[string]*oauth2.Config)
var jwtKey []byte
var ErrNoIDFound = errors.New("no id token found")

type OIDC struct {
}

// Initialize oidc config
func InitAuth() error {
	if provider != nil {
		return nil
	}
	oicdConfig := config.GetThelivConfig().Oidc
	p, err := oidc.NewProvider(context.Background(), oicdConfig.OidcProvider)
	if err != nil {
		log.S().Errorf("Unable to initialize oidc config: %v", err)
		return err
	}
	provider = p

	oidcConfig := &oidc.Config{
		ClientID: oicdConfig.ClientID,
	}
	verifier = provider.Verifier(oidcConfig)
	jwtKey = []byte(oicdConfig.ClientSecret)
	return nil
}

func (OIDC) GetUser(r *http.Request, getAd bool) (*rbac.User, error) {
	c, err := joinCookies(r.Context(), IDTokenKey, r.Cookies())
	if err != nil {
		log.SWithContext(r.Context()).Warnf("Cookie %v does not exist: %v", IDTokenKey, err)
		return nil, err
	}
	extract, err := userFromToken(r.Context(), c)
	if err != nil {
		return nil, err
	}
	user := rbac.User{}
	user.UID = extract.Id
	user.Surname = extract.FamilyName
	user.Givenname = extract.GivenName
	user.Displayname = extract.Name
	user.Emailaddress = extract.Email
	if getAd {
		user.AdGroups = extract.Roles
	}
	return &user, nil
}

func userFromToken(ctx context.Context, auth string) (*rbac.UserInfo, error) {
	user := &rbac.UserInfo{}
	token, err := jwt.ParseWithClaims(auth, user, func(t *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil {
		log.SWithContext(ctx).Errorf("authorization code/cookie invalid")
		return nil, err
	}
	c, ok := token.Claims.(*rbac.UserInfo)
	if !ok {
		msg := "cannot unmarshal ID token claim"
		err := fmt.Errorf(msg)
		log.SWithContext(ctx).Errorf(msg)
		return nil, err
	}
	if e := c.Valid(); e != nil {
		log.SWithContext(ctx).Errorf("failed to validate token")
		return nil, e
	}
	return user, nil
}

func CheckAuthorization(r *http.Request) (*http.Request, error) {
	for _, cookie := range r.Cookies() {
		if strings.HasPrefix(cookie.Name, IDTokenKey) {
			return r, nil
		}
	}
	return r, ErrNoIDFound
}

func HandleStartAuthFlow(w http.ResponseWriter, r *http.Request) {
	state := getRedirectUrl(r.Header.Get("redirect"))
	if state == "" {
		state = "/"
	}
	// redirect to authorization url
	url := getOauthConfig(r).AuthCodeURL(state)
	w.Header().Add("X-Location", url)
	w.WriteHeader(http.StatusUnauthorized)
	render.JSON(w, r, struct {
		Error string `json:"error"`
		Code  int    `json:"code"`
		Login bool   `json:"login"`
	}{url, http.StatusUnauthorized, true})
}

func getRedirectUrl(url string) (redirect string) {
	redirect = "/theliv/"
	path := strings.Split(url, redirect)
	if len(path) == 2 {
		redirect = redirect + path[1]
	}
	return
}
