/*
 * Copyright FMR LLC <opensource@fidelity.com>
 *
 * SPDX-License-Identifier: Apache
 */
package service

import (
	"context"
	"log"

	"github.com/fidelity/theliv/pkg/config"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func ListNs(clusterName string) []string {
	conf := config.GetConfigLoader().GetKubernetesConfig(clusterName)
	if conf == nil {
		return nil
	}
	kconf := conf.GetKubeConfig()
	clientset, err := kubernetes.NewForConfig(kconf)
	if err != nil {
		log.Println("Failed to init kubeconfig for cluster", clusterName, "error is", err)
		return nil
	}
	nsList, err := clientset.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Println("Failed to list namespaces for cluster", clusterName, "error is", err)
		return nil
	}
	var names []string
	for _, n := range nsList.Items {
		names = append(names, n.Name)
	}
	return names
}

func GetClusters() []string {
	return config.GetConfigLoader().GetK8SClusterNames()
}
