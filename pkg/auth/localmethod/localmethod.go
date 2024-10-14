/*
 * Copyright FMR LLC <opensource@fidelity.com>
 *
 * SPDX-License-Identifier: Apache
 */
package localmethod

import (
	"errors"
	"net/http"

	"github.com/fidelity/theliv/internal/rbac"
	"github.com/fidelity/theliv/pkg/database/etcd"
)

var accesskeyPrefix = "/theliv/accesskeys/"

func CheckAuthorization(r *http.Request) (*http.Request, error) {
	accesskey := r.Header.Get("ACCESSKEY")
	if len(accesskey) > 0 {
		content, err := etcd.Get(r.Context(), accesskeyPrefix+accesskey)
		if err != nil {
			return r, err
		}
		if content != nil {
			return r, nil
		}

	}
	return r, errors.New("not this Auth method")
}

type Localinfo struct {
}

func (Localinfo) GetUser(r *http.Request, getAd bool) (*rbac.User, error) {
	userinfo := &rbac.User{}
	accesskey := r.Header.Get("ACCESSKEY")
	err := etcd.GetObject(accesskeyPrefix+accesskey, userinfo)
	if err == nil {
		return userinfo, nil
	}
	return nil, err
}
