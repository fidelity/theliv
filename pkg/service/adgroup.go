/*
 * Copyright FMR LLC <opensource@fidelity.com>
 *
 * SPDX-License-Identifier: Apache
 */
package service

import (
	"context"
	"strings"

	auth "github.com/fidelity/theliv/pkg/auth/authmiddleware"
	"github.com/fidelity/theliv/pkg/database/etcd"
)

const (
	CLUSTER  = "/clusters/"
	DETECTOR = "/detector/"
)

// This function will add 2 new path, into every role.
func AddGroup(cluster string, namespace string, roles []string) (err error) {

	newPath := getPath(cluster, namespace)

	for _, role := range roles {
		if err = AddPath(role, newPath); err != nil {
			return
		}
	}
	return
}

// This function will remove 2 path, from every role.
func RemoveGroup(cluster string, namespace string, roles []string) (err error) {

	rmvPath := getPath(cluster, namespace)

	for _, role := range roles {
		if err = RemovePath(role, rmvPath); err != nil {
			return
		}
	}
	return
}

// Remove path from existing role
func RemovePath(roleName string, rmvPath []string) (err error) {
	var updatedValue []string
	var value []byte
	rolePath := auth.RolePrefix + roleName
	value, err = etcd.Get(context.Background(), rolePath)
	if err != nil {
		return
	}
	existingValue := string(value)
	if len(existingValue) == 0 {
		return
	} else {
		updatedValue = getRemovedElement(strings.Split(existingValue, SEPARATOR), rmvPath)
	}
	newValue := strings.Join(updatedValue, SEPARATOR)
	if existingValue == newValue {
		return
	}
	err = etcd.PutStr(rolePath, newValue)
	return
}

// Remove elements from existing slice.
func getRemovedElement(oldValue []string, removedValue []string) []string {
	for _, value := range removedValue {
		for index, revVal := range oldValue {
			if value == revVal {
				oldValue = append(oldValue[:index], oldValue[index+1:]...)
				break
			}
		}
	}
	return oldValue
}

// Generate 2 path, /cluster/cluster*, /detector/cluster/namespace*
func getPath(cluster string, namespace string) []string {
	return []string{
		CLUSTER + cluster + "*",
		DETECTOR + cluster + "/" + namespace + "*"}
}
