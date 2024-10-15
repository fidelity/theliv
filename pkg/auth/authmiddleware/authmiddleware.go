/*
 * Copyright FMR LLC <opensource@fidelity.com>
 *
 * SPDX-License-Identifier: Apache
 */
package authmiddleware

import (
	"errors"
	"net/http"
	"net/url"

	"github.com/fidelity/theliv/internal/rbac"
	"github.com/fidelity/theliv/pkg/auth/localmethod"
	"github.com/fidelity/theliv/pkg/auth/oidcmethod"
	"github.com/fidelity/theliv/pkg/config"
)

var ErrNotThisAuth = errors.New("not this Auth method")
var authMethod rbac.RBACInfo

func StartAuth(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//whitelist path
		oidc := config.GetThelivConfig().Oidc
		if r.URL.Path == "/theliv-api/v1/health" || r.URL.Path == "/theliv-api/v1/metrics" || r.URL.Path == getUrlPath(oidc.CallBack) {
			handler.ServeHTTP(w, r)
			return
		}
		//local auth
		r, err := localmethod.CheckAuthorization(r)
		if err == nil {
			authMethod = localmethod.Localinfo{}
			ok, err := checkRBAC(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if ok {
				handler.ServeHTTP(w, r)
			} else {
				http.Error(w, "", http.StatusForbidden)
			}
			return
		}
		if err.Error() == ErrNotThisAuth.Error() {
			//oidc auth
			r, err = oidcmethod.CheckAuthorization(r)
			if err == nil {
				authMethod = oidcmethod.OIDC{}
				ok, err := checkRBAC(r)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				if ok {
					handler.ServeHTTP(w, r)
				} else {
					http.Error(w, "", http.StatusForbidden)
				}
				return
			}
			if err == oidcmethod.ErrNoIDFound {
				oidcmethod.HandleStartAuthFlow(w, r)
				return
			}
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
	})
}

func GetUser(r *http.Request, getAds bool) (*rbac.User, error) {
	return authMethod.GetUser(r, getAds)
}

func getUrlPath(p string) string {
	if u, err := url.Parse(p); err == nil {
		return u.Path
	}
	return ""
}
