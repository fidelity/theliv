/*
 * Copyright FMR LLC <opensource@fidelity.com>
 *
 * SPDX-License-Identifier: Apache
 */
package service

import (
	"context"
	"fmt"

	"github.com/fidelity/theliv/pkg/common"
	com "github.com/fidelity/theliv/pkg/common"
	theErr "github.com/fidelity/theliv/pkg/err"
	log "github.com/fidelity/theliv/pkg/log"
	"go.uber.org/zap"

	"github.com/fidelity/theliv/pkg/config"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func ListNs(ctx context.Context, clusterName string) ([]string, error) {

	contact := fmt.Sprintf(com.Contact, config.GetThelivConfig().TeamName)

	l := log.SWithContext(ctx).With(
		zap.String("cluster", clusterName),
	)

	conf, err := config.GetConfigLoader().GetKubernetesConfig(ctx, clusterName)
	if err != nil {
		return nil, theErr.NewCommonError(ctx, 4, com.LoadKubeConfigFailed+contact)
	}
	kconf, err := conf.GetKubeConfig(ctx)
	if err != nil {
		l.Errorf("Failed to load kubeconfig: %v.", err)
		return nil, theErr.NewCommonError(ctx, 4, com.LoadKubeConfigFailed+contact)
	}
	clientset, err := kubernetes.NewForConfig(kconf)
	if err != nil {
		l.Errorf("Failed to init kubeconfig: %v.", err)
		return nil, theErr.NewCommonError(ctx, 4, com.LoadKubeConfigFailed+contact)
	}
	nsList, err := clientset.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		l.Errorf("Failed to list namespaces: %v.", err)
		return nil, theErr.NewCommonError(ctx, 4, com.ListNamespacesFailed+contact)
	}
	var names []string
	for _, n := range nsList.Items {
		names = append(names, n.Name)
	}
	return names, nil
}

func GetClusters(ctx context.Context) ([]string, error) {
	results, err := config.GetConfigLoader().GetK8SClusterNames(ctx)
	if err != nil {
		return results, theErr.NewCommonError(ctx, 2,
			common.DatabaseNoConnection+fmt.Sprintf(com.Contact, config.GetThelivConfig().TeamName))
	} else {
		return results, nil
	}
}
