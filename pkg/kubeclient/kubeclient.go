/*
 * Copyright FMR LLC <opensource@fidelity.com>
 *
 * SPDX-License-Identifier: Apache
 */
package kubeclient

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/fidelity/theliv/pkg/config"
	log "github.com/fidelity/theliv/pkg/log"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

type schemeGetter interface {
	Scheme() *runtime.Scheme
}

type reader interface {
	schemeGetter
	Get(context.Context, runtime.Object, NamespacedName, metav1.GetOptions) error
}

const (
	QPS_KEY       = "CLIENTGO_QPS"
	BURST_KEY     = "CLIENTGO_BURST"
	QPS_DEFAULT   = 50.0
	BURST_DEFAULT = 100
)

var _ reader = (*KubeClient)(nil)

type KubeClient struct {
	dynamicCli  dynamic.Interface
	discoverCli *discovery.DiscoveryClient
	mapper      *restmapper.DeferredDiscoveryRESTMapper
	scheme      *runtime.Scheme
}

const RetrieveErrorMessage = "unable retrieve the resource using dynamic client"

func NewKubeClient(ctx context.Context, cfg *restclient.Config, opts ...func(*KubeClient)) (*KubeClient, error) {
	thelivConfig := config.GetThelivConfig()
	if thelivConfig.QPS != 0.0 {
		cfg.QPS = thelivConfig.QPS
	} else {
		cfg.QPS = QPS_DEFAULT
	}

	if thelivConfig.Burst != 0 {
		cfg.Burst = thelivConfig.Burst
	} else {
		cfg.Burst = BURST_DEFAULT
	}
	log.SWithContext(ctx).Infof("Client-go configured with QPS = %f, Burst = %d", cfg.QPS, cfg.Burst)

	kc := &KubeClient{}
	dynamicClient, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("unable to prepare dynamic client: %w", err)
	}
	kc.dynamicCli = dynamicClient

	dc, err := discovery.NewDiscoveryClientForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("unable to prepare discovery client: %w", err)
	}
	kc.discoverCli = dc

	kc.mapper = restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(dc))

	kc.scheme = runtime.NewScheme()
	if err := clientgoscheme.AddToScheme(kc.scheme); err != nil {
		return nil, fmt.Errorf("error while trying to add default client-go schemes: %w", err)
	}

	// Functional options pattern which will enable the caller to specific any custom
	// modifications to the client. E.g caller might want to add a CRD to scheme.
	//
	// kcli, err := NewKubeClient(config, func(kc *Kubeclient) {
	//						_ = velerov1.AddToScheme(kc.Scheme())
	//					})

	for _, opt := range opts {
		opt(kc)
	}

	return kc, err

}

func (kc *KubeClient) Scheme() *runtime.Scheme {
	return kc.scheme
}

type NamespacedName struct {
	Name      string
	Namespace string
}

// Usage examples:
// For k8s native resources:
// kubeCli, err := NewKubeClient(config)
// dep := &appsv1.Deployment{}
// err := kubeCli.Get(context.TODO(), dep, NamespacedName{}, metav1.GetOptions{})
// dep.Status etc
//
// CRD example:
// kubeCli, err := NewKubeClient(config, func(kc *Kubeclient) {_ = velerov1.AddToScheme(kc.Scheme())})
// veleroBackup := (), &velerov1.Backup{}
// kubeCli.Get(context.TODO, veleroBackup, NamespacedName{}, metav1.GetOptions{})
// veleroBackup.Spec etc
func (kc *KubeClient) Get(ctx context.Context, obj runtime.Object, resName NamespacedName, getOps metav1.GetOptions) error {
	return kc.getResource(ctx, obj, resName, &getOps, GetSingleResourceName, GetSingleResoure)
}

func (kc *KubeClient) List(ctx context.Context, obj runtime.Object, resName NamespacedName, listOps metav1.ListOptions) error {
	return kc.getResource(ctx, obj, resName, &listOps, GetListResourceName, GetListResource)
}

type GetResourceName func(gvk *schema.GroupVersionKind)

// Return original gvk for single resource.
func GetSingleResourceName(gvk *schema.GroupVersionKind) {
}

// Replace the suffix "List" for resource list.
func GetListResourceName(gvk *schema.GroupVersionKind) {
	if strings.HasSuffix(gvk.Kind, "List") {
		gvk.Kind = gvk.Kind[:len(gvk.Kind)-4]
	}
}

type GetTargetResource func(ctx context.Context, dr dynamic.ResourceInterface, ops runtime.Object, resName NamespacedName) ([]byte, error)

func GetSingleResoure(ctx context.Context, dr dynamic.ResourceInterface, ops runtime.Object, resName NamespacedName) ([]byte, error) {
	getOps := *ops.(*metav1.GetOptions)
	resource, err := dr.Get(ctx, string(resName.Name), getOps)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", RetrieveErrorMessage, err)
	}
	return resource.MarshalJSON()
}

func GetListResource(ctx context.Context, dr dynamic.ResourceInterface, ops runtime.Object, resName NamespacedName) ([]byte, error) {
	listOps := *ops.(*metav1.ListOptions)
	resource, err := dr.List(ctx, listOps)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", RetrieveErrorMessage, err)
	}
	return resource.MarshalJSON()
}

// GetOwner returns the owner resource
func (kc *KubeClient) GetOwner(ctx context.Context, or metav1.OwnerReference, ns string) (metav1.Object, error) {

	gvk := schema.FromAPIVersionAndKind(or.APIVersion, or.Kind)

	var dr dynamic.ResourceInterface
	mapping, err := kc.mapper.RESTMapping(gvk.GroupKind(), gvk.GroupVersion().Version)
	if err != nil {
		return nil, fmt.Errorf("unable retrieve the restmapping from mapper: %w", err)
	}

	if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
		//namespaced resources should specify the namespace
		dr = kc.dynamicCli.Resource(mapping.Resource).Namespace(ns)
	} else {
		// for cluster-wide resources
		dr = kc.dynamicCli.Resource(mapping.Resource)
	}
	obj, err := dr.Get(ctx, or.Name, metav1.GetOptions{})
	return obj, err
}

func (kc *KubeClient) getResource(ctx context.Context, obj runtime.Object, resName NamespacedName, ops runtime.Object,
	getResourceName GetResourceName, getTargetResource GetTargetResource) error {
	gvk, err := apiutil.GVKForObject(obj, kc.scheme)
	getResourceName(&gvk)
	if err != nil {
		return fmt.Errorf("unable retrieve gvk for obj using apiutil: %w", err)
	}
	var dr dynamic.ResourceInterface
	mapping, err := kc.mapper.RESTMapping(gvk.GroupKind(), gvk.GroupVersion().Version)
	if err != nil {
		return fmt.Errorf("unable retrieve the restmapping from mapper: %w", err)
	}

	if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
		//namespaced resources should specify the namespace
		dr = kc.dynamicCli.Resource(mapping.Resource).Namespace(resName.Namespace)
	} else {
		// for cluster-wide resources
		dr = kc.dynamicCli.Resource(mapping.Resource)
	}
	data, err := getTargetResource(ctx, dr, ops, resName)
	if err != nil && strings.HasPrefix(err.Error(), RetrieveErrorMessage) {
		return fmt.Errorf(err.Error())
	} else if err != nil {
		return fmt.Errorf("error while marshalling json using unstructured in dynamic client: %w", err)
	}
	if err := json.Unmarshal(data, obj); err != nil {
		return fmt.Errorf("error while unmarshalling into resource struct: %w", err)
	}

	return err
}
