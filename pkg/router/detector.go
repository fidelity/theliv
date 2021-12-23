package router

import (
	"context"
	golog "log"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/fidelity/theliv/internal/problem"
	"github.com/fidelity/theliv/pkg/config"
	theliverr "github.com/fidelity/theliv/pkg/err"
	"github.com/fidelity/theliv/pkg/kubeclient"
	observability "github.com/fidelity/theliv/pkg/observability"
	datadog "github.com/fidelity/theliv/pkg/observability/datadog"
	k8s "github.com/fidelity/theliv/pkg/observability/k8s"
	"github.com/fidelity/theliv/pkg/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

func Detector(r chi.Router) {
	r.Get("/{cluster}/{namespace}/detect", detect)
}

func detect(w http.ResponseWriter, r *http.Request) {
	defer theliverr.PanicHandler(w)
	con, err := service.Detect(createDetectorInputWithContext(r))
	if err != nil {
		// TODO error
		w.WriteHeader(theliverr.GetStatusCode(err))
		render.JSON(w, r, err)
	}
	render.JSON(w, r, con)
}

func createDetectorInputWithContext(r *http.Request) context.Context {
	ctx := r.Context()
	// namespace
	cluster := chi.URLParam(r, "cluster")
	namespace := chi.URLParam(r, "namespace")

	// Get kubeconfig for the specified cluster
	conf := config.GetConfigLoader().GetKubernetesConfig(cluster)
	if conf == nil {
		return ctx
	}

	k8sconfig := conf.GetKubeConfig()
	awsconfig := conf.GetAwsConfig()
	var ac aws.Config
	if awsconfig != nil {
		ac = *awsconfig
	}
	thelivcfg := config.GetThelivConfig()

	// The kubeclient will be used for k8s logs/events driver.
	client, err := kubeclient.NewKubeClient(k8sconfig)
	if err != nil {
		golog.Printf("ERROR - Got error when getting deployment client with kubeclient, error is %s", err)
	}

	eventClient, logClient := getLogDriver(thelivcfg, client)
	eventDeeplinkClient, logDeeplinkClient := getDeeplinkDrivers(thelivcfg)

	input := &problem.DetectorCreationInput{
		Kubeconfig:             k8sconfig,
		ClusterName:            cluster,
		Namespace:              namespace,
		EventRetriever:         eventClient,
		LogRetriever:           logClient,
		EventDeeplinkRetriever: eventDeeplinkClient,
		LogDeeplinkRetriever:   logDeeplinkClient,
		AwsConfig:              ac,
	}

	return service.SetDetectorInput(ctx, input)
}

/*
Get LogDriver and EventsDriver.
The implementation will be decided by ThelivConfig.LogDriver and ThelivConfig.EventDriver.
If no implementation is specified, the K8s API driver will be used.
*/
func getLogDriver(thelivcfg *config.ThelivConfig, kubeclient *kubeclient.KubeClient) (eventRetriever observability.EventRetriever, logRetriever observability.LogRetriever) {

	datadogConfig := getDataDogConfig(thelivcfg)

	switch thelivcfg.EventDriver {
	case config.DriverDatadog:
		eventRetriever = datadog.NewDatadogEventRetriever(datadogConfig)
	default:
		eventRetriever = k8s.NewK8sEventRetriever(kubeclient)
	}

	switch thelivcfg.LogDriver {
	case config.DriverDatadog:
		logRetriever = datadog.NewDatadogLogRetriever(datadogConfig)
	default:
		logRetriever = k8s.NewK8sLogRetriever(kubeclient)
	}
	return
}

func getDeeplinkDrivers(thelivcfg *config.ThelivConfig) (
	eventDeeplinkRetriever observability.EventDeeplinkRetriever,
	logDeeplinkRetriever observability.LogDeeplinkRetriever) {
	switch thelivcfg.EventDeeplinkDriver {
	case config.DriverDatadog:
		eventDeeplinkRetriever = datadog.DatadogEventDeeplinkRetriever{
			DatadogHost: thelivcfg.Datadog.DatadogHost,
		}
	default:
		eventDeeplinkRetriever = k8s.K8sEventDeeplinkRetriever{}
	}

	switch thelivcfg.LogDeeplinkDriver {
	case config.DriverDatadog:
		logDeeplinkRetriever = datadog.DatadogLogDeeplinkRetriever{
			DatadogHost: thelivcfg.Datadog.DatadogHost,
		}
	default:
		logDeeplinkRetriever = k8s.K8sLogDeeplinkRetriever{}
	}
	return
}

func getDataDogConfig(thelivcfg *config.ThelivConfig) datadog.DatadogConfig {
	return datadog.DatadogConfig{
		ClientApiKey: thelivcfg.Datadog.ApiKey,
		ClientAppkey: thelivcfg.Datadog.AppKey,
		AppId:        thelivcfg.Datadog.Index,
		MaxRecords:   int32(thelivcfg.Datadog.MaxRecords),
		Debug:        thelivcfg.Datadog.Debug,
		DatadogHost:  thelivcfg.Datadog.DatadogHost,
	}
}
