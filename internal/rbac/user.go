/*
 * Copyright FMR LLC <opensource@fidelity.com>
 *
 * SPDX-License-Identifier: Apache
 */
package rbac

import "net/http"

type User struct {
	Surname      string `json:"Surname,omitempty"`
	Givenname    string `json:"Givenname,omitempty"`
	UID          string `json:"UID,omitempty"`
	Displayname  string `json:"Displayname,omitempty"`
	Emailaddress string `json:"Emailaddress,omitempty"`
}

type RBACInfo interface {
	GetUser(r *http.Request) (*User, error)
	GetADgroups(r *http.Request) ([]string, error)
}
