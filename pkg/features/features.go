package features

import (
	"sync"
)

type featuresImpl struct {
	sync.RWMutex
	nameToID      map[string]int64
	state         map[int64]bool
	cursorRuntime int64
	cursorConfig  int64
}

// RegisterRuntime registers feature that can be set in runtime
// You must provide feature name and you'll get FeatureID in the end
// All runtime IDs are >=0
func (f *featuresImpl) RegisterRuntime(name string, def bool) (int64, error) {
	f.Lock()
	defer f.Unlock()
	if id, ok := f.nameToID[name]; ok {
		return id, ErrAlreadyExists
	}

	nextID := f.cursorRuntime
	f.nameToID[name] = nextID
	f.state[nextID] = def
	f.cursorRuntime++

	return nextID, nil
}

// RegisterConfig registers feature that can be set only in config
// You must provide feature name and you'll get FeatureID in the end
// All config-time IDs are negative
func (f *featuresImpl) RegisterConfig(name string, def bool) (int64, error) {
	f.Lock()
	defer f.Unlock()
	if id, ok := f.nameToID[name]; ok {
		return id, ErrAlreadyExists
	}

	nextID := f.cursorConfig
	f.nameToID[name] = nextID
	f.state[nextID] = def
	f.cursorConfig--

	return nextID, nil
}

// IsEnabledID allows to check if this feature was enabled by it's ID
func (f *featuresImpl) IsEnabledID(id int64) bool {
	f.RLock()
	defer f.RUnlock()
	if enabled, ok := f.state[id]; ok {
		return enabled
	}
	return false
}

// IsEnabledName allows to check if this feature was enabled by it's name
func (f *featuresImpl) IsEnabledName(name string) bool {
	f.RLock()
	defer f.RUnlock()
	if id, ok := f.nameToID[name]; ok {
		if enabled, ok := f.state[id]; ok {
			return enabled
		}
	}
	return false
}

// GetIDByName gets id of feature if exists by it's name
// Useful to do conditional tests, when you don't know what
// ID feature flag got
func (f *featuresImpl) GetIDByName(name string) (int64, error) {
	f.RLock()
	defer f.RUnlock()
	if id, ok := f.nameToID[name]; ok {
		return id, nil
	}
	return 0, ErrFeatureNotRegistered
}

// SetFlagByID updates flag status by flag id
func (f *featuresImpl) SetFlagByID(id int64, enabled bool) bool {
	if id < 0 {
		return false
	}
	f.Lock()
	defer f.Unlock()
	if _, ok := f.state[id]; ok {
		f.state[id] = enabled
		return true
	}
	return false
}

// SetFlagByName updates flag status by name
func (f *featuresImpl) SetFlagByName(name string, enabled bool) bool {
	f.Lock()
	defer f.Unlock()
	if id, ok := f.nameToID[name]; ok {
		if id < 0 {
			return false
		}
		f.state[id] = enabled
		return true
	}
	return false
}

// SetFlagByNameForced is a function that should be during initialization
// It's not thread-safe (on purpose) and allows to change even config-time flags
// returns error if there is no such flag
func (f *featuresImpl) SetFlagByNameForced(name string, enabled bool) error {
	if id, ok := f.nameToID[name]; ok {
		f.state[id] = enabled
		return nil
	}
	return ErrFeatureNotRegistered
}

type featuresSingleton struct {
	features Features
}

var featuresInstance *featuresSingleton
var once sync.Once

// GetFeaturesInstance returns or initialize feature instance
func GetFeaturesInstance() Features {
	once.Do(func() {
		featuresInstance = &featuresSingleton{
			features: newFeatures(),
		}
	})
	return featuresInstance.features
}

// NewFeatures creates a new instance of feature flag controller
func newFeatures() Features {
	return &featuresImpl{
		nameToID:     make(map[string]int64),
		state:        make(map[int64]bool),
		cursorConfig: -1,
	}
}
