package authmiddleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/fidelity/theliv/pkg/config"
	"github.com/fidelity/theliv/pkg/database/etcd"
)

const UIDPrefix string = "/theliv/uids/"
const RolePrefix string = "/theliv/roles/"
const URLPrefix string = "/theliv-api/v1"

func getRole(UID string) ([]string, error) {
	UID = UIDPrefix + UID
	value, err := etcd.Get(UID)
	if err != nil {
		return nil, err
	}
	roles := strings.Split(string(value), ",")
	return roles[:], nil
}
func getPath(role string) ([]string, error) {
	role = RolePrefix + role
	value, err := etcd.Get(role)
	if err != nil {
		return nil, err
	}
	if len(value) < 1 {
		return nil, nil
	}
	paths := strings.Split(string(value), ",")
	return paths[:], nil
}

func checkRBAC(r *http.Request) (bool, error) {
	user, err := GetUser(r)
	if err != nil {
		return false, err
	}
	path := r.URL.Path
	path = strings.TrimSuffix(path, "/")
	skip, err := checkPattern([]string{path}, config.GetThelivConfig().Auth.WhitelistPath[:])
	if err != nil {
		return false, err
	}
	if skip {
		return true, err
	}
	if strings.HasPrefix(path, URLPrefix) {
		path = path[14:]
	} else {
		return false, err
	}
	roles, err := getRole(user.UID)
	if err != nil {
		return false, err
	}
	adgroups, err := GetADgroups(r)
	if err != nil {
		return false, err
	}
	roles = append(roles, adgroups...)
	var grantPath []string
	for _, r := range roles {
		path, err := getPath(r)
		if err != nil {
			return false, err
		}
		if path != nil {
			grantPath = append(grantPath, path...)
		}
	}
	if len(grantPath) < 1 {
		return false, nil
	}
	matched, err := checkPattern(grantPath[:], []string{path})
	if err != nil {
		return false, err
	}
	return matched, err
}

//check any of string match any of the pattern
func checkPattern(patterns []string, strings []string) (bool, error) {
	matched := false
	if len(patterns) < 1 || len(strings) < 1 {
		return false, errors.New("pattern or string array is empty")
	}
	for _, s := range strings {
		for _, p := range patterns {
			matched = KeyMatch(s, p)
			if matched {
				return matched, nil
			}
		}
	}
	return matched, nil
}

// KeyMatch determines whether key1 matches the pattern of key2 (similar to RESTful path), key2 can contain a *.
// For example, "/foo/bar" matches "/foo/*"
// Ref: https://github.com/casbin/casbin/blob/master/util/builtin_operators.go
func KeyMatch(key1 string, key2 string) bool {
	i := strings.Index(key2, "*")
	if i == -1 {
		return key1 == key2
	}

	if len(key1) > i {
		return key1[:i] == key2[:i]
	}
	return key1 == key2[:i]
}
