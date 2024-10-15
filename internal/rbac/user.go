/*
 * Copyright FMR LLC <opensource@fidelity.com>
 *
 * SPDX-License-Identifier: Apache
 */
package rbac

import (
	"net/http"

	"github.com/golang-jwt/jwt/v4"
)

type User struct {
	Surname      string   `json:"Surname,omitempty"`
	Givenname    string   `json:"Givenname,omitempty"`
	UID          string   `json:"UID,omitempty"`
	Displayname  string   `json:"Displayname,omitempty"`
	Emailaddress string   `json:"Emailaddress,omitempty"`
	AdGroups     []string `json:"roles,omitempty"`
}

type ThelivUser struct {
	jwt.RegisteredClaims
	DisplayName       string   `json:"displayName,omitempty"`
	GivenName         string   `json:"givenName,omitempty"`
	JobTitle          string   `json:"jobTitle,omitempty"`
	Mail              string   `json:"mail,omitempty"`
	Email             string   `json:"email,omitempty"`
	Surname           string   `json:"surname,omitempty"`
	UserPrincipalName string   `json:"userPrincipalName,omitempty"`
	Upn               string   `json:"upn"`
	Groups            []string `json:"groups,omitempty"`
}

type UserInfo struct {
	Name       string   `json:"name"`
	FamilyName string   `json:"family_name"`
	GivenName  string   `json:"given_name"`
	Email      string   `json:"email"`
	Upn        string   `json:"upn"`
	OnBehalf   string   `json:"onbehalf"` // on behalf of user name -- corpid
	Roles      []string `json:"roles"`
	jwt.StandardClaims
}

type RBACInfo interface {
	GetUser(r *http.Request, getAd bool) (*User, error)
}
