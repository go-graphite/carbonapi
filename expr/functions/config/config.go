package config

import "time"

var Config = struct {
	ExtractTagsFromArgs bool
	DefaultTimeZone     *time.Location
}{
	DefaultTimeZone: time.UTC,
}
