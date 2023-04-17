/*
 * Copyright FMR LLC <opensource@fidelity.com>
 *
 * SPDX-License-Identifier: Apache
 */
package service

import (
	"strings"
	"context"

	auth "github.com/fidelity/theliv/pkg/auth/authmiddleware"
	"github.com/fidelity/theliv/pkg/database/etcd"
)

const SEPARATOR = ","

// This function will insert new Role kv into database, if key not found.
// If KV exists in database, append newly added path.
// Value in database is string format, a collection seperated by ",".
// If database operation failed, return error, else return nil.
func AddPath(ctx context.Context, roleName string, newPaths []string) (err error) {
	var updatedValue []string
	var value []byte
	rolePath := auth.RolePrefix + roleName
	value, err = etcd.Get(ctx, rolePath)
	if err != nil {
		return
	}
	existingValue := string(value)
	if len(existingValue) == 0 {
		updatedValue = newPaths
	} else {
		updatedValue = getUpdatedElement(strings.Split(existingValue, SEPARATOR), newPaths)
	}
	newValue := strings.Join(updatedValue, SEPARATOR)
	if existingValue == newValue {
		return
	}
	err = etcd.PutStr(ctx, rolePath, newValue)
	return
}

// This function will append element to the first array,
// which exists in the second array but non-existing in the first array.
func getUpdatedElement(existingElements []string, newElements []string) []string {
	for _, newVal := range newElements {
		matched := false
		for _, oldVal := range existingElements {
			if oldVal == newVal {
				matched = true
			}
		}
		if !matched {
			existingElements = append(existingElements, newVal)
		}
	}
	return existingElements
}
