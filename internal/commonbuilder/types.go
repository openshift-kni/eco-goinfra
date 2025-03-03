package commonbuilder

import (
	goclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// CommonBuilderInterface is an interface implemented by all builders.
type CommonBuilderInterface interface {
	GetClient() goclient.Client
	GetDefinition() goclient.Object
	GetErrorMsg() string
	GetKind() string
}

// CommonBuilder is a struct used to implement common builder behavior.
type CommonBuilder struct {
	CommonBuilderInterface
}
