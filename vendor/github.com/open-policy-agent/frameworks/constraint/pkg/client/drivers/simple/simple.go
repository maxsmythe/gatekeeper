package simple

import (
	"context"
	"errors"
	"strings"
	"sync"

	apiconstraints "github.com/open-policy-agent/frameworks/constraint/pkg/apis/constraints"
	"github.com/open-policy-agent/frameworks/constraint/pkg/client/drivers"
	"github.com/open-policy-agent/frameworks/constraint/pkg/core/templates"
	"github.com/open-policy-agent/frameworks/constraint/pkg/types"
	"github.com/open-policy-agent/opa/storage"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const (
	// Name is the name of the driver
	Name = "Simple"
)

func New() *Driver {
	return &Driver{
		code: make(map[string]string),
	}
}

var _ drivers.Driver = &Driver{}

// Driver is a threadsafe Rego environment for compiling Rego in ConstraintTemplates,
// registering Constraints, and executing queries.
type Driver struct {
	// mtx guards access to the storage and target maps.
	mtx sync.RWMutex

	// code maps the template name to the rejection statement to return
	code map[string]string
}

// Name returns the name of the driver.
func (d *Driver) Name() string {
	return Name
}

// AddTemplate adds templ to Driver. Normalizes modules into usable forms for
// use in queries.
func (d *Driver) AddTemplate(ctx context.Context, ct *templates.ConstraintTemplate) error {
	if len(ct.Spec.Targets) != 1 {
		return errors.New("wrong number of targets defined, only 1 target allowed")
	}
	var rawCode map[string]interface{}
	for _, code := range ct.Spec.Targets[0].Code {
		if code.Engine != Name {
			continue
		}
		objMap, ok := code.Source.Value.(map[string]interface{})
		if !ok {
			return errors.New("SimplePolicy code malformed")
		}
		rawCode = objMap
		break
	}
	if rawCode == nil {
		return errors.New("SimplePolicy code not defined")
	}

	messageRaw, ok := rawCode["rejectWith"]
	if !ok {
		return errors.New("There is no rejection message")
	}

	message, ok := messageRaw.(string)
	if !ok {
		return errors.New("Rejection message is not a string")
	}

	d.mtx.Lock()
	defer d.mtx.Unlock()
	d.code[ct.GetName()] = message
	return nil
}

// RemoveTemplate removes all Compilers and Constraints for templ.
// Returns nil if templ does not exist.
func (d *Driver) RemoveTemplate(ctx context.Context, ct *templates.ConstraintTemplate) error {
	d.mtx.Lock()
	defer d.mtx.Unlock()
	delete(d.code, ct.GetName())
	return nil
}

// AddConstraint adds Constraint to Rego storage. Future calls to Query will
// be evaluated against Constraint if the Constraint's key is passed.
func (d *Driver) AddConstraint(ctx context.Context, constraint *unstructured.Unstructured) error {
	return nil
}

// RemoveConstraint removes Constraint from Rego storage. Future calls to Query
// will not be evaluated against the constraint. Queries which specify the
// constraint's key will silently not evaluate the Constraint.
func (d *Driver) RemoveConstraint(ctx context.Context, constraint *unstructured.Unstructured) error {
	return nil
}

// AddData adds data to Rego storage at data.inventory.path.
func (d *Driver) AddData(ctx context.Context, target string, path storage.Path, data interface{}) error {
	return nil
}

// RemoveData deletes data from Rego storage at data.inventory.path.
func (d *Driver) RemoveData(ctx context.Context, target string, path storage.Path) error {
	return nil
}

func (d *Driver) Query(ctx context.Context, target string, constraints []*unstructured.Unstructured, review interface{}, opts ...drivers.QueryOpt) ([]*types.Result, *string, error) {
	results := []*types.Result{}
	for i := range constraints {
		constraint := constraints[i]
		result := &types.Result{
			Msg:        d.code[strings.ToLower(constraint.GetObjectKind().GroupVersionKind().Kind)],
			Constraint: constraint,
			// TODO: the engine should not determine the enforcement action -- that does not work with CEL KEP
			EnforcementAction: apiconstraints.EnforcementActionDeny,
		}
		results = append(results, result)
	}
	return results, nil, nil
}

func (d *Driver) Dump(ctx context.Context) (string, error) {
	return "", nil
}
