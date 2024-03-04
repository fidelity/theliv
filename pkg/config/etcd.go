/*
 * Copyright FMR LLC <opensource@fidelity.com>
 *
 * SPDX-License-Identifier: Apache
 */
package config

import (
	"context"
	"errors"
	"fmt"

	"regexp"
	"strings"

	driver "github.com/fidelity/theliv/pkg/database/etcd"
	log "github.com/fidelity/theliv/pkg/log"
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
		log.S().Panicf("Failed to load theliv config, error is %v\n", err)
	}
	if err := ecl.loadDatadogConfig(); err != nil {
		log.S().Errorf("Failed to load datadog config, error is %v\n", err)
	}

	if err := ecl.loadAuthConfig(); err != nil {
		log.S().Errorf("Failed to load auth config, error is %v\n", err)
	}

	if err := ecl.loadPrometheusConfig(); err != nil {
		log.S().Errorf("Failed to load prometheus config, error is %v\n", err)
	}

	if err := ecl.loadThelivLevelConfig(); err != nil {
		log.S().Errorf("Failed to load theliv level config, error is %v\n", err)
	}

	if err := ecl.loadLdapConfig(); err != nil {
		log.S().Errorf("Failed to load ldap config, error is %v\n", err)
	}

	if err := ecl.loadAzureConfig(); err != nil {
		log.S().Errorf("Failed to load azure config, error is %v\n", err)
	}

	if err := ecl.loadAiConfig(); err != nil {
		log.S().Errorf("Failed to load ai config, error is %v\n", err)
	}
}

func (ecl *EtcdConfigLoader) GetKubernetesConfig(ctx context.Context, name string) (*KubernetesCluster, error) {
	env := getK8SEnv(name)
	key := fmt.Sprintf("%v/%v/%v", driver.CLUSTERS_KEY, env, name)
	conf := &KubernetesCluster{}
	err := driver.GetObjectWithSub(ctx, key, conf)
	if err != nil {
		log.SWithContext(ctx).Errorf("Failed to load theliv config from etcd, error is %v\n", err)
		return nil, err
	}
	if len(conf.KubeConf) == 0 {
		return nil, errors.New("empty KubeConf")
	}
	return conf, nil
}

func (ecl *EtcdConfigLoader) GetK8SClusterNames(ctx context.Context) ([]string, error) {
	keys, err := driver.GetKeys(ctx, driver.CLUSTERS_KEY)
	if err != nil {
		log.SWithContext(ctx).Errorf("Failed to load cluster keys, error is %v.", err)
		return keys, err
	}
	tmp := make(map[string]string)
	result := make([]string, 0)
	// Only get the cluster name from keys
	// key looks like /theliv/clusters/eks/cluster-name-with-dash-and-1/kubeconf
	re := regexp.MustCompile(fmt.Sprintf("%v/[era]ks/([0-9a-z-]+)", driver.CLUSTERS_KEY))
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
	return result, nil
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

func getK8SEnv(cluster string) string {
	re := regexp.MustCompile("[era]ks")
	m := re.FindAllStringSubmatch(cluster, -1)
	env := m[0][0]
	return env
}

func (ecl *EtcdConfigLoader) loadThelivConfig() error {
	conf := &ThelivConfig{}
	err := driver.GetObject(driver.THELIV_CONFIG_KEY, conf)
	if err != nil {
		log.S().Errorf("Failed to load theliv config from etcd, error is %v\n", err)
		return err
	}
	thelivConfig = conf
	log.S().Infof("Load theliv config from etcd: %v\n", conf)
	return nil
}

func (ecl *EtcdConfigLoader) loadDatadogConfig() error {
	conf := &DatadogConfig{}
	err := driver.GetObject(driver.DATADOG_CONFIG_KEY, conf)
	if err != nil {
		return err
	}
	thelivConfig.Datadog = conf
	log.S().Infof("Successfully load Datadog config %v\n", conf.ToMaskString())
	return nil
}

func (ecl *EtcdConfigLoader) loadAuthConfig() error {
	conf := &AuthConfig{}
	err := driver.GetObjectWithSub(context.Background(), driver.THELIV_AUTH_KEY, conf)
	if err != nil {
		return err
	}
	thelivConfig.Auth = conf
	log.S().Infof("Successfully load auth config %v\n", conf.ToMaskString())
	return nil
}

func (ecl *EtcdConfigLoader) loadPrometheusConfig() error {
	conf := &PrometheusConfig{}
	err := driver.GetObjectWithSub(context.Background(), driver.PROMETHEUS_GLOBAL_CONFIG_KEY, conf)
	if err != nil {
		return err
	}
	thelivConfig.Prometheus = conf
	log.S().Infof("Successfully load prometheus config")
	return nil
}

func (ecl *EtcdConfigLoader) loadThelivLevelConfig() error {
	conf := &ProblemLevelConfig{}
	err := driver.GetObjectWithSub(context.Background(), driver.THELIV_LEVEL_CONFIG_KEY, conf)
	if err != nil {
		return err
	}
	thelivConfig.ProblemLevel = conf
	log.S().Infof("Successfully load theliv level config")
	return nil
}

func getLastPart(key string) string {
	names := strings.Split(key, "/")
	return names[len(names)-1]
}

func (ecl *EtcdConfigLoader) loadLdapConfig() error {
	conf := &LdapConfig{}
	err := driver.GetObjectWithSub(context.Background(), driver.LDAP_CONFIG_KEY, conf)
	if err != nil {
		return err
	}
	thelivConfig.Ldap = conf
	log.S().Infof("Successfully load ldap config")
	return nil
}

func (ecl *EtcdConfigLoader) loadAzureConfig() error {
	conf := &AzureConfig{}
	err := driver.GetObjectWithSub(context.Background(), driver.AZURE_CONFIG_KEY, conf)
	if err != nil {
		return err
	}
	thelivConfig.Azure = conf
	log.S().Infof("Successfully load azure config")
	return nil
}

func (ecl *EtcdConfigLoader) loadAiConfig() error {
	conf := &AiConfig{}
	err := driver.GetObjectWithSub(context.Background(), driver.AI_CONFIG_KEY, conf)
	if err != nil {
		return err
	}
	thelivConfig.Ai = conf
	log.S().Infof("Successfully load Ai config")
	return nil
}
