/*
 * Copyright FMR LLC <opensource@fidelity.com>
 *
 * SPDX-License-Identifier: Apache
 */
package service

import (
	"context"
	"encoding/json"

	invest "github.com/fidelity/theliv/internal/investigators"
	"github.com/fidelity/theliv/pkg/database/etcd"
)

const (
	KeyPath  = "/theliv/clusters/"
	KubeConf = "/kubeconf"
	BasicKey = "/basic"

	TokenTemplate = `apiVersion: v1
clusters:
- cluster:
    certificate-authority-data: {{.CA}}
    server: {{.Url}}
  name: {{.Name}}
contexts:
- context:
    cluster: {{.Name}}
    user: {{.Name}}
  name: {{.Name}}
current-context: {{.Name}}
kind: Config
preferences: {}
users:
- name: {{.Name}}
  user:
    token: {{.Token}}`
)

type ClusterBasic struct {
	Url   string
	CA    string
	Name  string
	Token string
	AWS   AWS
}

type AWS struct {
	Account string `json:"account"`
}

// Insert or update 1 record, to /theliv/clusters/{name}/kubeconf.
func RegisterCluster(ctx context.Context, basic ClusterBasic) error {

	clusterType := basic.Name[:3]

	value, err := invest.ExecGoTemplate(ctx, TokenTemplate, basic)
	if err != nil {
		return err
	}

	// if aws account id present, convert to json and insert in db
	if basic.AWS != (AWS{}) {
		basicJson, err := json.Marshal(basic.AWS)
		if err != nil {
			return err
		}

		if err := etcd.PutStr(KeyPath+clusterType+"/"+basic.Name+BasicKey, string(basicJson)); err != nil {
			return err
		}
	}

	return etcd.PutStr(KeyPath+clusterType+"/"+basic.Name+KubeConf, value)

}
