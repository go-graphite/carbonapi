package features

import (
	"sync"
	"testing"
)

type feature struct {
	name        string
	value       bool
	assignedID  int64
	expectedID  int64
	expectedErr error
}

type query struct {
	byName         string
	byID           int64
	expectedResult bool
}

type featuresBaseTestCase struct {
	name                  string
	runtimeFeatures       []feature
	configFeatures        []feature
	expectedCursorConfig  int64
	expectedCursorRuntime int64

	extraQueries []query
}

func TestFeaturesSingleRoutine(t *testing.T) {
	tests := []featuresBaseTestCase{
		{
			name:            "empty set",
			runtimeFeatures: []feature{},
			configFeatures:  []feature{},
			extraQueries:    []query{},
		},
		{
			name: "some features",
			runtimeFeatures: []feature{
				{
					name:        "test1",
					value:       false,
					expectedErr: nil,
				},
				{
					name:        "test1",
					value:       false,
					expectedErr: ErrAlreadyExists,
				},
				{
					name:        "test2",
					value:       true,
					expectedErr: nil,
				},
			},
			configFeatures: []feature{
				{
					name:        "test3",
					value:       false,
					expectedErr: nil,
				},
				{
					name:        "test1",
					value:       false,
					expectedErr: ErrAlreadyExists,
				},
				{
					name:        "test4",
					value:       false,
					expectedErr: nil,
				},
			},
			extraQueries: []query{
				{
					byName:         "blah",
					byID:           100500,
					expectedResult: false,
				},
				{
					byName:         "test1",
					byID:           0,
					expectedResult: false,
				},
				{
					byName:         "test2",
					byID:           1,
					expectedResult: true,
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			f := NewFeatures()
			realF := f.(*featuresImpl)

			insertedRuntimeFeatures := 0
			for i := range test.runtimeFeatures {
				feature := &test.runtimeFeatures[i]
				id, err := f.RegisterRuntime(feature.name, feature.value)
				if err != feature.expectedErr {
					t.Errorf("unexpected error value, got %v, expected %v", err, feature.expectedErr)
				}
				if err == nil {
					if id < 0 {
						t.Errorf("got negative id for runtime flag, but it's reserved for config flags")
					}
					insertedRuntimeFeatures++
					feature.assignedID = id
				}
			}

			insertedConfigFeatures := 0
			for i := range test.configFeatures {
				feature := &test.configFeatures[i]
				id, err := f.RegisterConfig(feature.name, feature.value)
				if err != feature.expectedErr {
					t.Errorf("unexpected error value, got %v, expected %v", err, feature.expectedErr)
				}
				if err == nil {
					if id >= 0 {
						t.Errorf("got non-negative id for config flag, but it's reserved for runtime flags")
					}
					insertedConfigFeatures++
					feature.assignedID = id
				}
				feature.assignedID = id
			}

			insertedFeatures := insertedConfigFeatures + insertedRuntimeFeatures
			if len(realF.state) != insertedFeatures {
				t.Errorf("len(state) mismatch, got %v, expected %v", len(realF.state), insertedFeatures)
			}

			if len(realF.nameToID) != insertedFeatures {
				t.Errorf("len(nameToID) mismatch, got %v, expected 0", len(realF.nameToID))
			}

			if realF.cursorConfig != -1-int64(insertedConfigFeatures) {
				t.Errorf("cursorConfig mismatch, got %v, expected %v", realF.cursorConfig, -1-int64(insertedConfigFeatures))
			}

			if realF.cursorRuntime != int64(insertedRuntimeFeatures) {
				t.Errorf("cursorRuntime mismatch, got %v, expected %v", realF.cursorRuntime, insertedRuntimeFeatures)
			}

			for _, feature := range test.runtimeFeatures {
				if feature.expectedErr != nil {
					continue
				}
				ok := f.SetFlagByName(feature.name, feature.value)
				if !ok {
					t.Errorf("failed to update existing runtime feature flag '%v' by name", feature.name)
				}

				ok = f.SetFlagByID(feature.assignedID, feature.value)
				if !ok {
					t.Errorf("failed to update existing runtime feature flag '%v' by id", feature.name)
				}
			}

			for _, feature := range test.configFeatures {
				if feature.expectedErr != nil {
					continue
				}
				ok := f.SetFlagByName(feature.name, feature.value)
				if ok {
					t.Errorf("successfully updated config feature flag '%v' by name, when we shouldn't", feature.name)
				}

				ok = f.SetFlagByID(feature.assignedID, feature.value)
				if ok {
					t.Errorf("successfully updated config feature flag '%v' by id '%v', when we shouldn't", feature.name, feature.assignedID)
				}
			}
		})
	}
}

func TestFeaturesMultipleRoutinesReadOnly(t *testing.T) {
	goRoutines := 1000
	test := featuresBaseTestCase{
		name: "some features",
		runtimeFeatures: []feature{
			{
				name:        "test1",
				value:       false,
				expectedErr: nil,
			},
			{
				name:        "test2",
				value:       true,
				expectedErr: nil,
			},
		},
	}

	f := NewFeatures()

	insertedRuntimeFeatures := 0
	for i := range test.runtimeFeatures {
		feature := &test.runtimeFeatures[i]
		id, err := f.RegisterRuntime(feature.name, feature.value)
		if err != feature.expectedErr {
			t.Errorf("unexpected error value, got %v, expected %v", err, feature.expectedErr)
		}
		if err == nil {
			if id < 0 {
				t.Errorf("got negative id for runtime flag, but it's reserved for config flags")
			}
			insertedRuntimeFeatures++
			feature.assignedID = id
		}
	}

	insertedConfigFeatures := 0
	for i := range test.configFeatures {
		feature := &test.configFeatures[i]
		id, err := f.RegisterConfig(feature.name, feature.value)
		if err != feature.expectedErr {
			t.Errorf("unexpected error value, got %v, expected %v", err, feature.expectedErr)
		}
		if err == nil {
			if id >= 0 {
				t.Errorf("got non-negative id for config flag, but it's reserved for runtime flags")
			}
			insertedConfigFeatures++
			feature.assignedID = id
		}
		feature.assignedID = id
	}

	startChan := make(chan struct{})
	wg := sync.WaitGroup{}
	routine := 0
	for j := 0; j < len(test.runtimeFeatures); j++ {
		go func(feature feature) {
			<-startChan

			ok := f.IsEnabledID(feature.assignedID)
			if ok != feature.value {
				t.Errorf("unexpected value, got %v, expected %v", ok, feature.value)
			}
		}(test.runtimeFeatures[j])

		routine++
		if routine == goRoutines {
			break
		}
	}
	close(startChan)
	wg.Wait()
}

func TestFeaturesMultipleRoutinesReadUpdate(t *testing.T) {
	goRoutines := 200
	test := featuresBaseTestCase{
		name: "some features",
		runtimeFeatures: []feature{
			{
				name:        "test1",
				value:       false,
				expectedErr: nil,
			},
			{
				name:        "test2",
				value:       true,
				expectedErr: nil,
			},
		},
		configFeatures: []feature{
			{
				name:        "test3",
				value:       false,
				expectedErr: nil,
			},
			{
				name:        "test4",
				value:       true,
				expectedErr: nil,
			},
		},
	}

	f := NewFeatures()

	insertedRuntimeFeatures := 0
	for i := range test.runtimeFeatures {
		feature := &test.runtimeFeatures[i]
		id, err := f.RegisterRuntime(feature.name, feature.value)
		if err != feature.expectedErr {
			t.Errorf("unexpected error value, got %v, expected %v", err, feature.expectedErr)
		}
		if err == nil {
			if id < 0 {
				t.Errorf("got negative id for runtime flag, but it's reserved for config flags")
			}
			insertedRuntimeFeatures++
			feature.assignedID = id
		}
	}

	insertedConfigFeatures := 0
	for i := range test.configFeatures {
		feature := &test.configFeatures[i]
		id, err := f.RegisterConfig(feature.name, feature.value)
		if err != feature.expectedErr {
			t.Errorf("unexpected error value, got %v, expected %v", err, feature.expectedErr)
		}
		if err == nil {
			if id >= 0 {
				t.Errorf("got non-negative id for config flag, but it's reserved for runtime flags")
			}
			insertedConfigFeatures++
			feature.assignedID = id
		}
		feature.assignedID = id
	}

	startChan := make(chan struct{})
	wg := sync.WaitGroup{}
	routine := 0
	for j := 0; j < len(test.runtimeFeatures); j++ {
		go func(feature feature) {
			<-startChan

			ok := f.IsEnabledID(feature.assignedID)
			if ok != feature.value {
				t.Errorf("unexpected value, got %v, expected %v", ok, feature.value)
			}
		}(test.runtimeFeatures[j])

		routine++
		if routine == goRoutines {
			break
		}
	}

	routine = 0
	for j := 0; j < len(test.configFeatures); j++ {
		go func(feature feature) {
			<-startChan

			ok := f.IsEnabledID(feature.assignedID)
			if ok != feature.value {
				t.Errorf("unexpected value, got %v, expected %v", ok, feature.value)
			}
		}(test.configFeatures[j])

		routine++
		if routine == goRoutines {
			break
		}
	}

	routine = 0
	for j := 0; j < len(test.runtimeFeatures); j++ {
		go func(feature feature) {
			<-startChan

			ok := f.SetFlagByName(feature.name, feature.value)
			if !ok {
				t.Errorf("failed to update runtime feature")
			}
		}(test.runtimeFeatures[j])

		routine++
		if routine == goRoutines {
			break
		}
	}

	routine = 0
	for j := 0; j < len(test.configFeatures); j++ {
		go func(feature feature) {
			<-startChan

			ok := f.SetFlagByName(feature.name, feature.value)
			if ok {
				t.Errorf("updated config feature, but shouldn't")
			}
		}(test.configFeatures[j])

		routine++
		if routine == goRoutines {
			break
		}
	}
	close(startChan)
	wg.Wait()
}

func BenchmarkFeaturesSingleThreadOverheadByID(b *testing.B) {
	test := featuresBaseTestCase{
		name: "some features",
		runtimeFeatures: []feature{
			{
				name:        "test1",
				value:       false,
				expectedErr: nil,
			},
			{
				name:        "test2",
				value:       true,
				expectedErr: nil,
			},
		},
	}

	f := NewFeatures()

	insertedRuntimeFeatures := 0
	for i := range test.runtimeFeatures {
		feature := &test.runtimeFeatures[i]
		id, err := f.RegisterRuntime(feature.name, feature.value)
		if err != feature.expectedErr {
			b.Fatalf("unexpected error value, got %v, expected %v", err, feature.expectedErr)
		}
		if err == nil {
			if id < 0 {
				b.Fatalf("got negative id for runtime flag, but it's reserved for config flags")
			}
			insertedRuntimeFeatures++
			feature.assignedID = id
		}
	}

	insertedConfigFeatures := 0
	for i := range test.configFeatures {
		feature := &test.configFeatures[i]
		id, err := f.RegisterConfig(feature.name, feature.value)
		if err != feature.expectedErr {
			b.Fatalf("unexpected error value, got %v, expected %v", err, feature.expectedErr)
		}
		if err == nil {
			if id >= 0 {
				b.Fatalf("got non-negative id for config flag, but it's reserved for runtime flags")
			}
			insertedConfigFeatures++
			feature.assignedID = id
		}
		feature.assignedID = id
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		id := int64(i % len(test.runtimeFeatures))
		_ = f.IsEnabledID(id)
	}
}
