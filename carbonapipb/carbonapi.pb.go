package carbonapipb

type AccessLogDetails struct {
	Handler                       string            `json:"handler,omitempty"`
	CarbonapiUUID                 string            `json:"carbonapi_uuid,omitempty"`
	Username                      string            `json:"username,omitempty"`
	URL                           string            `json:"url,omitempty"`
	PeerIP                        string            `json:"peer_ip,omitempty"`
	PeerPort                      string            `json:"peer_port,omitempty"`
	Host                          string            `json:"host,omitempty"`
	Referer                       string            `json:"referer,omitempty"`
	Format                        string            `json:"format,omitempty"`
	UseCache                      bool              `json:"use_cache,omitempty"`
	Targets                       []string          `json:"targets,omitempty"`
	CacheTimeout                  int32             `json:"cache_timeout,omitempty"`
	Metrics                       []string          `json:"metrics,omitempty"`
	HaveNonFatalErrors            bool              `json:"have_non_fatal_errors,omitempty"`
	Runtime                       float64           `json:"runtime,omitempty"`
	HTTPCode                      int32             `json:"http_code,omitempty"`
	CarbonzipperResponseSizeBytes int64             `json:"carbonzipper_response_size_bytes,omitempty"`
	CarbonapiResponseSizeBytes    int64             `json:"carbonapi_response_size_bytes,omitempty"`
	Reason                        string            `json:"reason,omitempty"`
	SendGlobs                     bool              `json:"send_globs,omitempty"`
	From                          int64             `json:"from,omitempty"`
	Until                         int64             `json:"until,omitempty"`
	MaxDataPoints                 int64             `json:"max_data_points,omitempty"`
	Tz                            string            `json:"tz,omitempty"`
	FromRaw                       string            `json:"from_raw,omitempty"`
	UntilRaw                      string            `json:"until_raw,omitempty"`
	URI                           string            `json:"uri,omitempty"`
	FromCache                     bool              `json:"from_cache"`
	UsedBackendCache              bool              `json:"used_backend_cache"`
	ZipperRequests                uint64            `json:"zipper_requests,omitempty"`
	TotalMetricsCount             uint64            `json:"total_metrics_count,omitempty"`
	RequestHeaders                map[string]string `json:"request_headers"`
}
