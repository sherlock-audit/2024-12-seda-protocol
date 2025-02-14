package pluginaws

import (
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/aws/request"
)

type CustomRetryer struct {
	client.DefaultRetryer
}

func (r CustomRetryer) MaxRetries() int {
	return r.DefaultRetryer.MaxRetries()
}

func (r CustomRetryer) RetryRules(req *request.Request) time.Duration {
	return r.DefaultRetryer.RetryRules(req)
}

func (r CustomRetryer) ShouldRetry(req *request.Request) bool {
	if strings.Contains(req.Error.Error(), "read: connection reset") {
		return true
	}

	// Fallback to SDK's built in retry rules
	return r.DefaultRetryer.ShouldRetry(req)
}

func AddRetryToConfig(cfg *aws.Config) *aws.Config {
	request.WithRetryer(cfg, CustomRetryer{DefaultRetryer: client.DefaultRetryer{
		NumMaxRetries:    client.DefaultRetryerMaxNumRetries,
		MinRetryDelay:    client.DefaultRetryerMinRetryDelay,
		MaxRetryDelay:    client.DefaultRetryerMaxRetryDelay,
		MinThrottleDelay: client.DefaultRetryerMinThrottleDelay,
		MaxThrottleDelay: client.DefaultRetryerMaxThrottleDelay,
	}})

	return cfg
}
