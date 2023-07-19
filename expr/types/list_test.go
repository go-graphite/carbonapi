package types

import (
	"encoding/json"
	"math"
	"reflect"
	"testing"
)

func TestSuggestion_MarshalJSON(t *testing.T) {
	for _, tc := range []struct {
		name                 string
		s                    *Suggestion
		expectedJSON         string
		expectedUnmarshalled *Suggestion // used when unmarshalling the result JSON doesn't equal the original suggestion
	}{
		{
			name:                 "int",
			s:                    NewSuggestion(1234),
			expectedJSON:         `1234`,
			expectedUnmarshalled: NewSuggestion(1234.), // JSON numbers are always unmarshalled as floats

		},
		{
			name:         "float",
			s:            NewSuggestion(12.34),
			expectedJSON: `12.34`,
		},
		{
			name:         "inf",
			s:            NewSuggestion(math.Inf(1)),
			expectedJSON: `1e9999`,
		},
		{
			name:         "-inf",
			s:            NewSuggestion(math.Inf(-1)),
			expectedJSON: `-1e9999`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := json.Marshal(tc.s)
			if err != nil {
				t.Fatalf("Error marshalling suggestion, err: %s", err.Error())
			}
			actualStr := string(actual)
			if actualStr != tc.expectedJSON {
				t.Fatalf("Marshalled JSON not equal: got\n%+v\nwant\n%+v", actualStr, tc.expectedJSON)
			}
			var unmarshalled Suggestion
			err = json.Unmarshal(actual, &unmarshalled)
			if err != nil {
				t.Fatalf("Error unmarshalling suggestion, err: %s", err.Error())
			}
			expectedUnmarshalled := *tc.s
			if tc.expectedUnmarshalled != nil {
				expectedUnmarshalled = *tc.expectedUnmarshalled
			}
			if !reflect.DeepEqual(unmarshalled, expectedUnmarshalled) {
				t.Fatalf("Unmarshalled JSON not equal: got\n%+v\nwant\n%+v", unmarshalled, expectedUnmarshalled)
			}
		})
	}
}
