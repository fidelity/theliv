package observability

import (
	"context"
	"net/url"
	"time"
)

// LogRecord represents a single log line
type LogRecord struct {
	Message  string
	Metadata map[string]string
	//May change DeepLink to URL if needed.
	// DeepLink will be extremely useful when you want to attach a link to your analysis (as an out of
	// problem detector). Ideally user would click on that link, authenticate with the central logging system
	// to view the logs (& other log analysis features that the central logging system might provide)
	DeepLink url.URL
}

type LogFilterCriteria struct {
	FilterCriteria map[string]string
	StartTime      time.Time
	EndTime        time.Time
	//Filter []LogRecord with the regular expression.
	RegularExpression string
}

// FilterValidator will validate the filter criteria. Validate method is expected to throw an error
// if any of the filter entries are incorrect. E.g incorrect key, unsupported special character
// in value etc. Different LogRetriever implementation can plug in their own implementation
// This interface may enabled in FilterValidator later.
// type FilterValidator interface {
// 	Validate(FilterCriteria) error
// }

// LogRetriever is aimed at providing a uniform interface for querying the centralized logging systems.
// This will typically be used by various problem detector implementations who might want to query
// the logs to detect an issue and provide the reference to the user. The interface is also aimed at
// staying away from vendor specific features to detect an issue since then the problem detector logic
// might not work for someone who is not using that vendor. When you use this interface, though you can
// retrieve any amount of data you want, it is strongly discouraged to do so. Most of the operations
// need to happen on the central logging system (applying filter, regex etc). The end goal should always
// be to reduce the amount of logs to a bare minimum that might be useful for the users to look at.
type LogRetriever interface {
	Retrieve(LogFilterCriteria) LogDataRef
}

// While LogRetriever implementation will take care of authnz with the central logging system and fetching
// you a pointer or reference to a log data, LogDataRef aims at providing very specific operations that
// can be applied on top of that LogDataRef.
type LogDataRef interface {
	GetRecords(ctx context.Context) ([]LogRecord, error)
}
