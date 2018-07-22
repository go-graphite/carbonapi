package features

import (
	"fmt"
	"net/http"
)

/*
 * There are two types of flags:
 *   1. Runtime (mutable) - flags that are safe to change during runtime (via API)
 *   2. Config (immutable) - flags that shouldn't be changed during runtime, they must be set via config.
        In code they must be set with SetFlagByNameForced function when config is parsed.
*/

var (
	// ErrAlreadyExists is an error when you are trying to register feature that already registered
	ErrAlreadyExists = fmt.Errorf("feature flag already registered")
	// ErrFeatureNotRegistered is an error that's returned when there is no flag registered with specified name
	ErrFeatureNotRegistered = fmt.Errorf("feature flag not registered")
)

// ChangeRequestByID is a type that's used by FlagPatchByIDHandler
type ChangeRequestByID struct {
	ID    int64
	Value bool
}

// ChangeRequestByName is a type that's used by FlagPatchByNameHandler
type ChangeRequestByName struct {
	Name  string
	Value bool
}

// Features - interface to provide a simple feature flag framework
// Package should declare new feature flags in it's init function.
type Features interface {
	// RegisterRuntime registers feature that can be set in runtime
	// You must provide feature name and you'll get FeatureID in the end
	RegisterRuntime(name string, def bool) (int64, error)

	// RegisterConfig registers feature that can be set only in config
	// You must provide feature name and you'll get FeatureID in the end
	RegisterConfig(name string, def bool) (int64, error)

	// IsEnabledID allows to check if this feature was enabled by it's ID
	// if flag not found returns false
	IsEnabledID(id int64) bool

	// IsEnabledName allows to check if this feature was enabled by it's name
	// if flag not found returns false
	IsEnabledName(name string) bool

	// SetFlagByID updates flag status by flag id
	// returns true on success and false if flag not found
	SetFlagByID(id int64, enabled bool) bool

	// SetFlagByName updates flag status by name
	// returns true on success and false if flag not found
	SetFlagByName(name string, enabled bool) bool

	// GetIDByName gets id of feature if exists by it's name
	// Useful to do conditional tests, when you don't know what
	// ID feature flag got
	GetIDByName(name string) (int64, error)

	// FlagListHandler is an HTTP Handler that provides current flag state
	//
	FlagListHandler(w http.ResponseWriter, r *http.Request)

	// FlagPatchByIDHandler is an HTTP Handler controls current flag configuration
	// Handler supports PATCH requests only. Accepts []ChangeRequestByID
	//
	//   If user tries to change flag that can't be change in a runtime
	//   HTTP 400 error will be produced and no changes will be applied
	FlagPatchByIDHandler(w http.ResponseWriter, r *http.Request)

	// FlagPatchByIDHandler is an HTTP Handler controls current flag configuration
	// Handler supports PATCH requests only. Accepts []ChangeRequestByName
	//
	//   If user tries to change flag that can't be change in a runtime
	//   HTTP 400 error will be produced and no changes will be applied
	FlagPatchByNameHandler(w http.ResponseWriter, r *http.Request)

	// SetFlagByNameForced is a function that should be during initialization
	// It's not thread-safe (on purpose) and allows to change even config-time flags
	// returns error if there is no such flag
	SetFlagByNameForced(name string, enabled bool) error
}

/*
TODO:
 1. Think of a way to sync flags
 2. Find a way to improve performance for check-mostly cases. E.x. treat whole config as immutable and migrate to atomics
 3. Implement a way to do A/B tests, e.x. by user name.
*/
