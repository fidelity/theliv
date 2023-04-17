/*
 * Copyright FMR LLC <opensource@fidelity.com>
 *
 * SPDX-License-Identifier: Apache
 */
package config

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"path"

	log "github.com/fidelity/theliv/pkg/log"

	"sigs.k8s.io/yaml"
)

var _ ConfigLoader = &FileConfigLoader{}

var k8sConfig = make(map[string]*KubernetesCluster)

// FileConfigLoader loads all the configuration from files
type FileConfigLoader struct {
	ThelivConfigFile string
}

// NewFileConfigLoader creates new FileConfigLoader from configfile
func NewFileConfigLoader(configfile string) *FileConfigLoader {
	loader := &FileConfigLoader{
		ThelivConfigFile: configfile,
	}
	configLoader = loader
	return loader
}

func (l *FileConfigLoader) LoadConfigs(ctx context.Context) {
	if err := l.loadThelivConfig(); err != nil {
		log.SWithContext(ctx).Fatalf("Failed to load theliv config, %v", err)
	}

	if err := l.loadKubernetesConfig(); err != nil {
		log.SWithContext(ctx).Fatalf("Failed to load kubernetes configs, %v", err)
	}
}

func (l *FileConfigLoader) GetKubernetesConfig(ctx context.Context, name string) *KubernetesCluster {
	return k8sConfig[name]
}

func (l *FileConfigLoader) GetK8SClusterNames(ctx context.Context) []string {
	names := make([]string, len(k8sConfig))
	i := 0

	for key, _ := range k8sConfig {
		names[i] = key
		i++
	}
	return names
}

// load kubernetes cluster configs from files under folder `ThelivConfig.ClusterDir`
// file name is exactly cluster name
func (l *FileConfigLoader) loadKubernetesConfig() error {
	// List files
	dir := path.Join(thelivConfig.ClusterDir)
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, file := range files {
		// Ignore folders
		if file.IsDir() {
			continue
		}
		kube, err := ioutil.ReadFile(path.Join(dir, file.Name()))
		if err != nil {
			//TODO log
			continue
		}
		k8sConfig[file.Name()] = &KubernetesCluster{
			Basic: ClusterBasicInfo{
				Name: file.Name(),
			},
			KubeConf: kube,
		}
	}
	return nil
}

// Loads Theliv config
// TODO error handling and log
func (l *FileConfigLoader) loadThelivConfig() error {
	content, err := ioutil.ReadFile(l.ThelivConfigFile)
	if err != nil {
		return err
	}
	jsonContent, err := yaml.YAMLToJSON(content)
	if err != nil {
		return err
	}

	theliv := ThelivConfig{}
	err = json.Unmarshal(jsonContent, &theliv)
	if err != nil {
		return err
	}
	thelivConfig = &theliv

	//set default value for k8s folder, same folder with ThelivConfigFile
	if thelivConfig.ClusterDir == "" {
		thelivConfig.ClusterDir = path.Dir(l.ThelivConfigFile) + "/k8s"
	}
	return nil
}
