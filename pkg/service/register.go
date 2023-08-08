/*
 * Copyright FMR LLC <opensource@fidelity.com>
 *
 * SPDX-License-Identifier: Apache
 */
package service

import (
	"context"
	"strings"

	invest "github.com/fidelity/theliv/internal/investigators"
	"github.com/fidelity/theliv/pkg/database/etcd"
	log "github.com/fidelity/theliv/pkg/log"
	"go.uber.org/zap"
)

const (
	KeyPath     = "/theliv/clusters/"
	KubeConfKey = "/kubeconf"
	BasicKey    = "/basic"

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

	BasicTemplate = `{
	"provider": "aws",
	"account": "{{.Account}}",
	"role": "arn:aws:iam::{{.Account}}:role/EKS_Theliv",
	"name": "{{.Name}}",
	"region": "{{.Region}}"
}`
)

type ClusterBasic struct {
	Url     string
	CA      string
	Name    string
	Token   string
	Account string
	Region  string
}

// Insert or update 1 record, to /theliv/clusters/{name}/kubeconf.
func RegisterCluster(ctx context.Context, basic ClusterBasic) error {
	clusterType := basic.Name[:3]
	etcdBaseKey := KeyPath + clusterType + "/" + basic.Name

	l := log.SWithContext(ctx).With(
		zap.String("cluster", basic.Name),
	)
	l.Info("Registering cluster with Theliv")

	// if aws account id present, convert to json and insert in db
	if basic.Account != "" {
		l.Infof("AWS config found with account ID: %s", basic.Account)
		if urlSlice := strings.Split(basic.Url, "."); len(urlSlice) > 5 {
			basic.Region = urlSlice[len(urlSlice)-4]
		}

		basicJson, err := invest.ExecGoTemplate(ctx, BasicTemplate, basic)
		if err != nil {
			return err
		}

		if err := etcd.PutStr(etcdBaseKey+BasicKey, basicJson); err != nil {
			return err
		}
	}

	value, err := invest.ExecGoTemplate(ctx, TokenTemplate, basic)
	if err != nil {
		return err
	}

	return etcd.PutStr(etcdBaseKey+KubeConfKey, value)

}
