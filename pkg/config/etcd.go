/*
 * Copyright FMR LLC <opensource@fidelity.com>
 *
 * SPDX-License-Identifier: Apache
 */
package config

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	driver "github.com/fidelity/theliv/pkg/database/etcd"
)

var _ ConfigLoader = &EtcdConfigLoader{}

type EtcdConfigLoader struct {
	CAFile    string
	CertFile  string
	KeyFile   string
	Endpoints []string
}

func (ecl *EtcdConfigLoader) LoadConfigs() {
	driver.InitClientConfig(ecl.CAFile, ecl.CertFile, ecl.KeyFile, ecl.Endpoints)
	if err := ecl.loadThelivConfig(); err != nil {
		log.Panicf("Failed to load theliv config, error is %v\n", err)
	}
	if err := ecl.loadDatadogConfig(); err != nil {
		log.Printf("Failed to load datadog config, error is %v\n", err)
	}

	if err := ecl.loadAuthConfig(); err != nil {
		log.Printf("Failed to load auth config, error is %v\n", err)
	}

	if err := ecl.loadPrometheusConfig(); err != nil {
		log.Printf("Failed to load prometheus config, error is %v\n", err)
	}

	if err := ecl.loadThelivLevelConfig(); err != nil {
		log.Printf("Failed to load theliv level config, error is %v\n", err)
	}
}

func (ecl *EtcdConfigLoader) GetKubernetesConfig(name string) *KubernetesCluster {
	key := fmt.Sprintf("%v/%v", driver.EKS_CLUSTERS_KEY, name)
	conf := &KubernetesCluster{}
	err := driver.GetObjectWithSub(key, conf)
	if err != nil {
		log.Printf("Failed to load theliv config from etcd, error is %v\n", err)
		return nil
	}
	if len(conf.KubeConf) == 0 {
		return nil
	}
	return conf
}

func (ecl *EtcdConfigLoader) GetK8SClusterNames() []string {
	keys, err := driver.GetKeys(driver.EKS_CLUSTERS_KEY)
	if err != nil {
		log.Println("Failed to load eks cluster keys")
		return keys
	}
	tmp := make(map[string]string)
	result := make([]string, 0)
	// Only get the cluster name from keys
	// key looks like /theliv/clusters/eks/cluster-name-with-dash-and-1/kubeconf
	re := regexp.MustCompile(fmt.Sprintf("%v/([0-9a-z-]+)", driver.EKS_CLUSTERS_KEY))
	for _, key := range keys {
		m := re.FindAllStringSubmatch(key, -1)
		for _, name := range m {
			// m should look like this
			// [["/theliv/clusters/eks/cluster-name-with-dash-and-1", "cluster-name-with-dash-and-1"]]
			// get the second ele as cluster name
			if len(name) < 2 {
				continue
			}
			// if the cluster already present, ignore
			if _, ok := tmp[name[1]]; !ok {
				tmp[name[1]] = name[1]
				result = append(result, name[1])
			}
		}
	}
	return result
}

func NewEtcdConfigLoader(ca, cert, key string, endpoints ...string) *EtcdConfigLoader {
	loader := &EtcdConfigLoader{
		CAFile:    ca,
		CertFile:  cert,
		KeyFile:   key,
		Endpoints: endpoints,
	}
	configLoader = loader
	return loader
}

func (ecl *EtcdConfigLoader) loadThelivConfig() error {
	conf := &ThelivConfig{}
	err := driver.GetObject(driver.THELIV_CONFIG_KEY, conf)
	if err != nil {
		log.Printf("Failed to load theliv config from etcd, error is %v\n", err)
		return err
	}
	thelivConfig = conf
	log.Printf("Load theliv config from etcd: %v\n", conf)
	return nil
}

func (ecl *EtcdConfigLoader) loadDatadogConfig() error {
	conf := &DatadogConfig{}
	err := driver.GetObject(driver.DATADOG_CONFIG_KEY, conf)
	if err != nil {
		return err
	}
	thelivConfig.Datadog = conf
	log.Printf("Successfully load Datadog config %v\n", conf.ToMaskString())
	return nil
}

func (ecl *EtcdConfigLoader) loadAuthConfig() error {
	conf := &AuthConfig{}
	err := driver.GetObjectWithSub(driver.THELIV_AUTH_KEY, conf)
	if err != nil {
		return err
	}
	thelivConfig.Auth = conf
	log.Printf("Successfully load auth config %v\n", conf.ToMaskString())
	return nil
}

func (ecl *EtcdConfigLoader) loadPrometheusConfig() error {
	conf := &PrometheusConfig{}
	err := driver.GetObjectWithSub(driver.PROMETHEUS_CONFIG_KEY, conf)
	if err != nil {
		return err
	}
	thelivConfig.Prometheus = conf
	log.Println("Successfully load prometheus config")
	return nil
}

func (ecl *EtcdConfigLoader) loadThelivLevelConfig() error {
	conf := &ProblemLevelConfig{}
	err := driver.GetObjectWithSub(driver.THELIV_LEVEL_CONFIG_KEY, conf)
	if err != nil {
		return err
	}
	thelivConfig.ProblemLevel = conf
	log.Println("Successfully load theliv level config")
	return nil
}

func getLastPart(key string) string {
	names := strings.Split(key, "/")
	return names[len(names)-1]
}
