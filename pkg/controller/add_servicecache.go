package controller

import (
	"service-cache-operator/pkg/controller/servicecache"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, servicecache.Add)
}
