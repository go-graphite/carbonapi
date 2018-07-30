package types

// Stats provides zipper-related statistics
type Stats struct {
	Timeouts          int64
	FindErrors        int64
	RenderErrors      int64
	InfoErrors        int64
	SearchRequests    int64
	SearchCacheHits   int64
	SearchCacheMisses int64
	ZipperRequests    int64

	MemoryUsage int64

	CacheMisses int64
	CacheHits   int64
}

func (s *Stats) Merge(stats *Stats) {
	s.Timeouts += stats.Timeouts
	s.FindErrors += stats.FindErrors
	s.RenderErrors += stats.RenderErrors
	s.InfoErrors += stats.InfoErrors
	s.SearchRequests += stats.SearchRequests
	s.SearchCacheHits += stats.SearchCacheHits
	s.SearchCacheMisses += stats.SearchCacheMisses
	s.MemoryUsage += stats.MemoryUsage
	s.CacheMisses += stats.CacheMisses
	s.CacheHits += stats.CacheHits
}
