package authmiddleware

import (
	"errors"
	"net/http"

	"github.com/fidelity/theliv/internal/rbac"
	"github.com/fidelity/theliv/pkg/auth/localmethod"
	"github.com/fidelity/theliv/pkg/auth/samlmethod"
	"github.com/wangli1030/saml/samlsp"
)

var ErrNotThisAuth = errors.New("not this Auth method")
var authMethod rbac.RBACInfo

func StartAuth(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//whitelist path
		if r.URL.Path == "/theliv-api/v1/health" || r.URL.Path == "/auth/saml/acr" || r.URL.Path == "/auth/saml/metadata" {
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
			//saml auth
			r, err = samlmethod.CheckAuthorization(r)
			if err == nil {
				authMethod = samlmethod.Samlinfo{}
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
			if err == samlsp.ErrNoSession {
				samlmethod.HandleStartAuthFlow(w, r)
				return
			}
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
	})
}

func GetUser(r *http.Request) (*rbac.User, error) {
	return authMethod.GetUser(r)
}

func GetADgroups(r *http.Request) ([]string, error) {
	return authMethod.GetADgroups(r)
}
