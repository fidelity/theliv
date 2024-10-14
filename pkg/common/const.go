/*
 * Copyright FMR LLC <opensource@fidelity.com>
 *
 * SPDX-License-Identifier: Apache
 */
package common

const (
	SkipTlsVerify = true

	Name          = "name"
	Namespace     = "namespace"
	Pod           = "pod"
	Container     = "container"
	Initcontainer = "initcontainer"
	Deployment    = "deployment"
	Replicaset    = "replicaset"
	Statefulset   = "statefulset"
	Daemonset     = "daemonset"
	Node          = "node"
	Job           = "job"
	Cronjob       = "cronjob"
	Service       = "service"
	Ingress       = "ingress"
	Endpoint      = "endpoint"
	Resourcetype  = "resourcetype"
	Blank         = " "
	Argo          = "Argo"
	Helm          = "Helm"

	NoUserInfo             = "your personal info not found, please try SSO again."
	DatabaseNoConnection   = "database can't be accessed,"
	LoadKubeConfigFailed   = "load cluster config failed, please find the cluster in the dropdown list, or"
	ListNamespacesFailed   = "failed to list namespaces,"
	PrometheusNotAvailable = "prometheus agent is either uninstalled or currently shut down by FinOps,"
	LoadResourceFailed     = "failed to load affected resources,"
	LoadEventsFailed       = "failed to load Kubernetes events in the namespace,"
	UncaughtApiErr         = "error occurred in Theliv API, we will track and fix it soon," + Thanks
	Contact                = " please contact %s for help." + Thanks
	Thanks                 = " Thanks for using Theliv!!"
)
