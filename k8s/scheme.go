package main

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	Group   = "ingress"
	Version = "v1"
)

var (
	internalGV              = schema.GroupVersion{Group: Group, Version: runtime.APIVersionInternal}
	externalGV              = schema.GroupVersion{Group: Group, Version: Version}
	embeddedTestExternalGVK = externalGV.WithKind("timestampData")

	SchemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)
	AddToScheme   = SchemeBuilder.AddToScheme
)

// Adds the list of known types to Scheme.
func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(internalGV, &TimestampData{})
	scheme.AddKnownTypeWithName(embeddedTestExternalGVK, &TimestampData{})
	return nil
}
