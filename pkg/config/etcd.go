/*
 * Copyright FMR LLC <opensource@fidelity.com>
 *
 * SPDX-License-Identifier: Apache
 */
package config

import (
	"fmt"
	"context"

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

func (ecl *EtcdConfigLoader) LoadConfigs(ctx context.Context) {
	driver.InitClientConfig(ctx, ecl.CAFile, ecl.CertFile, ecl.KeyFile, ecl.Endpoints)
	if err := ecl.loadThelivConfig(ctx); err != nil {
		log.SWithContext(ctx).Panicf("Failed to load theliv config, error is %v\n", err)
	}
	if err := ecl.loadDatadogConfig(ctx); err != nil {
		log.SWithContext(ctx).Errorf("Failed to load datadog config, error is %v\n", err)
	}

	if err := ecl.loadAuthConfig(ctx); err != nil {
		log.SWithContext(ctx).Errorf("Failed to load auth config, error is %v\n", err)
	}

	if err := ecl.loadPrometheusConfig(ctx); err != nil {
		log.SWithContext(ctx).Errorf("Failed to load prometheus config, error is %v\n", err)
	}

	if err := ecl.loadThelivLevelConfig(ctx); err != nil {
		log.SWithContext(ctx).Errorf("Failed to load theliv level config, error is %v\n", err)
	}

	if err := ecl.loadLdapConfig(ctx); err != nil {
		log.SWithContext(ctx).Errorf("Failed to load ldap config, error is %v\n", err)
	}
}

func (ecl *EtcdConfigLoader) GetKubernetesConfig(ctx context.Context, name string) *KubernetesCluster {
	env := getK8SEnv(name)
	key := fmt.Sprintf("%v/%v/%v", driver.CLUSTERS_KEY, env, name)
	conf := &KubernetesCluster{}
	err := driver.GetObjectWithSub(ctx, key, conf)
	if err != nil {
		log.SWithContext(ctx).Errorf("Failed to load theliv config from etcd, error is %v\n", err)
		return nil
	}
	if len(conf.KubeConf) == 0 {
		return nil
	}
	return conf
}

func (ecl *EtcdConfigLoader) GetK8SClusterNames(ctx context.Context) []string {
	keys, err := driver.GetKeys(ctx, driver.CLUSTERS_KEY)
	if err != nil {
		log.SWithContext(ctx).Error("Failed to load cluster keys")
		return keys
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

func getK8SEnv(cluster string) string {
	re := regexp.MustCompile("[era]ks")
	m := re.FindAllStringSubmatch(cluster, -1)
	env := m[0][0]
	return env
}

func (ecl *EtcdConfigLoader) loadThelivConfig(ctx context.Context) error {
	conf := &ThelivConfig{}
	err := driver.GetObject(ctx, driver.THELIV_CONFIG_KEY, conf)
	if err != nil {
		log.SWithContext(ctx).Errorf("Failed to load theliv config from etcd, error is %v\n", err)
		return err
	}
	thelivConfig = conf
	log.SWithContext(ctx).Infof("Load theliv config from etcd: %v\n", conf)
	return nil
}

func (ecl *EtcdConfigLoader) loadDatadogConfig(ctx context.Context) error {
	conf := &DatadogConfig{}
	err := driver.GetObject(ctx, driver.DATADOG_CONFIG_KEY, conf)
	if err != nil {
		return err
	}
	thelivConfig.Datadog = conf
	log.SWithContext(ctx).Infof("Successfully load Datadog config %v\n", conf.ToMaskString())
	return nil
}

func (ecl *EtcdConfigLoader) loadAuthConfig(ctx context.Context) error {
	conf := &AuthConfig{}
	err := driver.GetObjectWithSub(ctx, driver.THELIV_AUTH_KEY, conf)
	if err != nil {
		return err
	}
	thelivConfig.Auth = conf
	log.SWithContext(ctx).Infof("Successfully load auth config %v\n", conf.ToMaskString())
	return nil
}

func (ecl *EtcdConfigLoader) loadPrometheusConfig(ctx context.Context) error {
	conf := &PrometheusConfig{}
	err := driver.GetObjectWithSub(ctx, driver.PROMETHEUS_GLOBAL_CONFIG_KEY, conf)
	if err != nil {
		return err
	}
	thelivConfig.Prometheus = conf
	log.SWithContext(ctx).Infof("Successfully load prometheus config")
	return nil
}

func (ecl *EtcdConfigLoader) loadThelivLevelConfig(ctx context.Context) error {
	conf := &ProblemLevelConfig{}
	err := driver.GetObjectWithSub(ctx, driver.THELIV_LEVEL_CONFIG_KEY, conf)
	if err != nil {
		return err
	}
	thelivConfig.ProblemLevel = conf
	log.SWithContext(ctx).Infof("Successfully load theliv level config")
	return nil
}

func getLastPart(key string) string {
	names := strings.Split(key, "/")
	return names[len(names)-1]
}

func (ecl *EtcdConfigLoader) loadLdapConfig(ctx context.Context) error {
	conf := &LdapConfig{}
	err := driver.GetObjectWithSub(ctx, driver.LDAP_CONFIG_KEY, conf)
	if err != nil {
		return err
	}
	thelivConfig.Ldap = conf
	log.SWithContext(ctx).Infof("Successfully load ldap config")
	return nil
}
