/*
 * Copyright FMR LLC <opensource@fidelity.com>
 *
 * SPDX-License-Identifier: Apache
 */
package config

import (
	"encoding/json"
	"fmt"

	log "github.com/fidelity/theliv/pkg/log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type LogDriverType string

const (
	DriverDatadog       LogDriverType = "datadog"
	DriverDefaultDriver LogDriverType = "k8s"
)

// ConfigLoader loads configurations for Theliv system,
// And Kubernetes cluster based on K8S cluster ID.
type ConfigLoader interface {
	// ThelivConfig returns ThelivConfig
	LoadConfigs()
	GetKubernetesConfig(name string) *KubernetesCluster
	GetK8SClusterNames() []string
}

var (
	thelivConfig *ThelivConfig
	configLoader ConfigLoader
)

func GetConfigLoader() ConfigLoader {
	return configLoader
}

// GetThelivConfig returns theliv config
func GetThelivConfig() *ThelivConfig {
	return thelivConfig
}

// ThelivConfig is the configuration for Theliv system,
// Such as port for web service
type ThelivConfig struct {
	Port     int `json:"port"`
	LogLevel int `json:"loglevel"`
	// Only for file configs
	ClusterDir          string              `json:"clusterDir,omitempty"`
	Datadog             *DatadogConfig      `json:"datadog,omitempty"`
	Auth                *AuthConfig         `json:"auth,omitempty"`
	Prometheus          *PrometheusConfig   `json:"prometheus,omitempty"`
	ProblemLevel        *ProblemLevelConfig `json:"problemlevel,omitempty"`
	Ldap                *LdapConfig
	LogDriver           LogDriverType `json:"logDriver,omitempty"`
	EventDriver         LogDriverType `json:"eventDriver,omitempty"`
	LogDeeplinkDriver   LogDriverType `json:"logDeeplinkDriver,omitempty"`
	EventDeeplinkDriver LogDriverType `json:"eventDeeplinkDriver,omitempty"`
	// Only for UI usage
	EmailAddr       string `json:"emailAddr,omitempty"`
	DevelopedByTeam string `json:"developedByTeam,omitempty"`
	VideoLink       string `json:"videoLink,omitempty"`
}

func (c *ThelivConfig) ToMaskString() string {
	return fmt.Sprintf("TheliBasicConfig: Port: %v\n", c.Port)
}

type DatadogConfig struct {
	ApiKey      string `json:"apiKey"`
	AppKey      string `json:"appKey"`
	Index       string `json:"index"`
	MaxRecords  int    `json:"maxRecords"`
	Debug       bool   `json:"debug"`
	DatadogHost string `json:"host"`
}

func (c *DatadogConfig) ToMaskString() string {
	// TODO mask fields
	return fmt.Sprintf("Datadog config: \n ApiKey: *** \n AppKey: *** \n Index: %v\n MaxRecords: %v\n Debug: %v\n DatadogHost: %v\n",
		c.Index, c.MaxRecords, c.Debug, c.DatadogHost)
}

type AuthConfig struct {
	CertPath        string   `json:"certPath"`
	Cert            []byte   `json:"cert"`
	KeyPath         string   `json:"keyPath"`
	Key             []byte   `json:"key"`
	IDPMetadataPath string   `json:"idpmetadataPath"`
	IDPMetadata     []byte   `json:"idpmetadata"`
	IDPMetadataURL  string   `json:"idpmetadataURL"`
	RootURL         string   `json:"rootURL"`
	MetadataURL     string   `json:"metadataURL"`
	AcrURL          string   `json:"acrURL"`
	SloURL          string   `json:"sloURL"`
	EntityID        string   `json:"entityID"`
	WhitelistPath   []string `json:"whitelistpath"`
	ClientID        string   `json:"clientID"`
	ClientSecret    string   `json:"clientSecret"`
}

func (c *AuthConfig) ToMaskString() string {
	return fmt.Sprintf(`Auth config:
	CertPath: %v,
	Cert length: %v,
	ClientID length: %v,
	ClientSecret length: %v.
	KeyPath: %v,
	Key length: %v,
	IDPMetadataPath: %v,
	IDPMetadata length: %v,
	IDPMetadataURL: %v,
	MetadataURL: %v,
	AcrURL: %v,
	SloURL: %v,
	EntityID: %v,
	WhitelistPath: %v
	`, c.CertPath, len(c.Cert), len(c.ClientID), len(c.ClientSecret), c.KeyPath, len(c.Key), c.IDPMetadataPath,
		len(c.IDPMetadata), c.IDPMetadataURL, c.MetadataURL, c.AcrURL, c.SloURL, c.EntityID, c.WhitelistPath)
}

type PrometheusConfig struct {
	Address   string `json:"address"`
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Port      string `json:"port"`
}

type ProblemLevelConfig struct {
	ManagedNamespaces []string `json:"managednamespaces"`
}

type KubernetesCluster struct {
	Basic    ClusterBasicInfo `json:"basic"`
	KubeConf []byte           `json:"kubeconf"`
	AwsConf  []byte           `json:"awsconf"`
}

// Kubernetes cluster basic information
type ClusterBasicInfo struct {
	Name     string `json:"name"`
	Provider string `json:"provider"` // EKS, AKS, RKS ...
	Account  string `json:"account"`  //AWS account, Azure Subscription/resource group/,
	Role     string `json:"role"`     // AWS role arn, Azure AD Group
	Region   string `json:"region"`
}

// GetClusterConfig returns Kubernetes config based on cluster name
func (conf *KubernetesCluster) GetKubeConfig() *restclient.Config {
	client, err := clientcmd.RESTConfigFromKubeConfig(conf.KubeConf)
	if err != nil {
		log.S().Errorf("Failed to load kubernetes config, for cluster %v, error is %v\n", conf.Basic.Name, err)
		return nil
	}
	return client
}

func (conf *KubernetesCluster) GetAwsConfig() *aws.Config {
	awsconf := &AwsConfig{}
	err := json.Unmarshal(conf.AwsConf, awsconf)
	if err != nil {
		log.S().Errorf("Failed to load awsconfig for cluster %v, error is %v\n", conf.Basic.Name, err)
		return nil
	}

	return &aws.Config{
		Region:      conf.Basic.Region,
		Credentials: credentials.NewStaticCredentialsProvider(awsconf.Credentials.AccessKeyID, awsconf.Credentials.SecretAccessKey, awsconf.Credentials.SessionToken),
	}
}

type LdapConfig struct {
	Address string `json:"address"`
	Query   string `json:"query"`
}
