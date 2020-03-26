package types

// Stats provides zipper-related statistics
type Stats struct {
	Timeouts          int64
	FindRequests      int64
	FindErrors        int64
	FindTimeouts      int64
	RenderRequests    int64
	RenderErrors      int64
	RenderTimeouts    int64
	InfoRequests      int64
	InfoErrors        int64
	InfoTimeouts      int64
	SearchRequests    int64
	SearchCacheHits   int64
	SearchCacheMisses int64
	ZipperRequests    int64
	TotalMetricsCount int64

	MemoryUsage int64

	CacheMisses int64
	CacheHits   int64

	Servers       []string
	FailedServers []string
}

func (s *Stats) Merge(stats *Stats) {
	s.Timeouts += stats.Timeouts
	s.FindRequests += stats.FindRequests
	s.FindTimeouts += stats.FindTimeouts
	s.FindErrors += stats.FindErrors
	s.RenderRequests += stats.RenderRequests
	s.RenderTimeouts += stats.RenderTimeouts
	s.RenderErrors += stats.RenderErrors
	s.InfoRequests += stats.InfoRequests
	s.InfoTimeouts += stats.InfoTimeouts
	s.InfoErrors += stats.InfoErrors
	s.SearchRequests += stats.SearchRequests
	s.SearchCacheHits += stats.SearchCacheHits
	s.SearchCacheMisses += stats.SearchCacheMisses
	s.MemoryUsage += stats.MemoryUsage
	s.CacheMisses += stats.CacheMisses
	s.CacheHits += stats.CacheHits

	s.Servers = append(s.Servers, stats.Servers...)
	s.FailedServers = append(s.FailedServers, stats.FailedServers...)
}
