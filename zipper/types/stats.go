package types

// Stats provides zipper-related statistics
type Stats struct {
	Timeouts          uint64
	FindRequests      uint64
	FindErrors        uint64
	FindTimeouts      uint64
	ExpandRequests    uint64
	ExpandErrors      uint64
	ExpandTimeouts    uint64
	RenderRequests    uint64
	RenderErrors      uint64
	RenderTimeouts    uint64
	InfoRequests      uint64
	InfoErrors        uint64
	InfoTimeouts      uint64
	SearchRequests    uint64
	SearchCacheHits   uint64
	SearchCacheMisses uint64
	ZipperRequests    uint64
	TotalMetricsCount uint64

	MemoryUsage int64

	CacheMisses uint64
	CacheHits   uint64

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
