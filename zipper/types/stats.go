package types

// Stats provides zipper-related statistics
type Stats struct {
	FindErrors     int64
	RenderErrors   int64
	InfoErrors     int64
	ZipperRequests int64

}

func (s *Stats) Merge(stats *Stats) {
	s.FindErrors += stats.FindErrors
	s.RenderErrors += stats.RenderErrors
	s.InfoErrors += stats.InfoErrors
}
