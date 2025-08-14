package commonbuilder

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/golang/glog"
	"github.com/openshift-kni/eco-goinfra/pkg/msg"
)

// New returns CommonBuilder from a CommonBuilderInterface.
func New(builder CommonBuilderInterface) CommonBuilder {
	return CommonBuilder{
		CommonBuilderInterface: builder,
	}
}

// Validate will check that the builder and builder definition are properly initialized before
// accessing any member fields.
func (c *CommonBuilder) Validate() (bool, error) {
	if c == nil || reflect.ValueOf(c).IsNil() {
		glog.V(100).Infof("The builder is uninitialized")

		return false, fmt.Errorf("error: received nil builder")
	}

	resourceCRD := strings.ToLower(c.GetKind())

	if c.GetDefinition() == nil || reflect.ValueOf(c.GetDefinition()).IsNil() {
		glog.V(100).Infof("The %s is undefined", resourceCRD)

		return false, fmt.Errorf("%s", msg.UndefinedCrdObjectErrString(resourceCRD))
	}

	if c.GetClient() == nil || reflect.ValueOf(c.GetClient()).IsNil() {
		glog.V(100).Infof("The %s builder apiclient is nil", resourceCRD)

		return false, fmt.Errorf("%s builder cannot have nil apiClient", resourceCRD)
	}

	if c.GetErrorMsg() != "" {
		glog.V(100).Infof("The %s builder has error message: %s", resourceCRD, c.GetErrorMsg())

		return false, fmt.Errorf("%s", c.GetErrorMsg())
	}

	return true, nil
}
