// Package common provides a set of common functionality that can be used to reduce the boilerplate required for adding
// builders for new resources.
//
// The Builder interface is mainly for use within the common package and describes the methods that builders must
// implement to take advantage of the common functions. These common functions may be used to implement the typical CRUD
// methods of a builder.
//
// There is also groundwork for abstracting the logging and error handling. Currently, this is just a series of
// functions that mostly work the same as the direct logging calls and error returns. Having them separated allows us to
// standardize them and make future improvements easier.
package common
