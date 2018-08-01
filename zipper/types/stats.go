package types

// Stats provides zipper-related statistics
type Stats struct {
	FindRequests int64
	FindErrors   int64

	RenderRequests int64
	RenderErrors   int64

	InfoRequests int64
	InfoErrors   int64

	ZipperRequests int64
}

func (s *Stats) Merge(stats *Stats) {
	s.FindRequests += stats.FindRequests
	s.FindErrors += stats.FindErrors

	s.RenderRequests += stats.RenderRequests
	s.RenderErrors += stats.RenderErrors

	s.InfoRequests += stats.InfoRequests
	s.InfoErrors += stats.InfoErrors
}
