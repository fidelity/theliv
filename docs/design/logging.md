# Logging Driver

The logging driver will extract logs from backend logging service such as Datadog, Splunk, etc. The logging configuration will be get from configuration service.

## Package
We will use **log** as the package name. If user want to use go log package, they need to use alias name for the build-in log package.

## Interface and data strucutre
```
type LogRecord struct {
	Message  string
	Metadata map[string]string
	// DeepLink will be extremely useful when you want to attach a link to your analysis (as an out of
	// problem detector). Ideally user would click on that link, authenticate with the central logging system
	// to view the logs (& other log analysis features that the central logging system might provide)
	DeepLink url.URL
}

type FilterCriteria interface {
	FilterCriteria() map[string]string
	//Use a generic string type here, since some logging system supports syntax like "now-15m".
	//Standard time can be converted to string and vice versa.
	StartTime() string
	EndTime() string
	//Filter []LogRecord with the regular expression.
	RegularExpression() string
}

type LogRetriever interface {
	Retrieve(FilterCriteria) *LogDataRef
}

type LogDataRef interface {
	GetRecords() []LogRecord
	Count() int64
	Error() []LogRecord
	Info() []LogRecord
	DeepLink() string
}
```

## Datadog Implementation
For now we use Datadog as the initial implementation, in future we will add Splunk. User can implement the interface to support more logging services. Below are the structs of the Datadog log implementation, DatadogLogConfig contains Datadog related configurations, DatadogLog contains Datadog API client (datadog-api-client-go provided by Datadog), Context and DatadogLogConfig. It implements interface LogRetriever.

```
type DataDogLog struct {
	ctx              context.Context
	apiClient        datadog.APIClient
	datadogLogConfig DatadogLogConfig
}

type DatadogLogConfig struct {
	ClientApiKey string
	ClientAppkey string
	AppId        string
	From         string
	To           string
	MaxRecords   int32
	Debug        bool
}
```