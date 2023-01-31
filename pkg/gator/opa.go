package gator

import (
	constraintclient "github.com/open-policy-agent/frameworks/constraint/pkg/client"
	"github.com/open-policy-agent/frameworks/constraint/pkg/client/drivers/simple"
	"github.com/open-policy-agent/frameworks/constraint/pkg/client/drivers/rego"
	"github.com/open-policy-agent/gatekeeper/pkg/target"
)

func NewOPAClient(includeTrace bool) (Client, error) {
	opaDriver, err := rego.New(local.Tracing(includeTrace))
	if err != nil {
		return nil, err
	}

	c, err := constraintclient.NewClient(constraintclient.Targets(&target.K8sValidationTarget{}), constraintclient.Driver(simple.New()), constraintclient.Driver(driver))
	if err != nil {
		return nil, err
	}

	return c, nil
}
