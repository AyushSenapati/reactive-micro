package policyenforcer

import "errors"

var (
	// this error is returned when an implementation of policy storage
	// interface is required but nil was provided instead
	ErrNilPolicyStorage = errors.New("nil policy storage")

	// policy storage returns this error when update policy for the resource type
	// is requested but the resource type is not supported by the service
	ErrUnsupportedRtype = errors.New("unsupported resource type")

	// policy storage returns this error when update policy for the sub
	// is requested but the sub is not found in the local storage
	ErrSubNotCached = errors.New("sub is not cached")

	// this errors is returned when policy is of invalid type
	ErrInvalidPolicy = errors.New("invalid policy")
)
