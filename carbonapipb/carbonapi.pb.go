package carbonapipb

type AccessLogDetails struct {
	Handler                       string   `json:"handler,omitempty"`
	CarbonapiUuid                 string   `json:"carbonapi_uuid,omitempty"`
	Username                      string   `json:"username,omitempty"`
	Url                           string   `json:"url,omitempty"`
	PeerIp                        string   `json:"peer_ip,omitempty"`
	PeerPort                      string   `json:"peer_port,omitempty"`
	Host                          string   `json:"host,omitempty"`
	Referer                       string   `json:"referer,omitempty"`
	Format                        string   `json:"format,omitempty"`
	UseCache                      bool     `json:"use_cache,omitempty"`
	Targets                       []string `json:"targets,omitempty"`
	CacheTimeout                  int32    `json:"cache_timeout,omitempty"`
	Metrics                       []string `json:"metrics,omitempty"`
	HaveNonFatalErrors            bool     `json:"have_non_fatal_errors,omitempty"`
	Runtime                       float64  `json:"runtime,omitempty"`
	HttpCode                      int32    `json:"http_code,omitempty"`
	CarbonzipperResponseSizeBytes int64    `json:"carbonzipper_response_size_bytes,omitempty"`
	CarbonapiResponseSizeBytes    int64    `json:"carbonapi_response_size_bytes,omitempty"`
	Reason                        string   `json:"reason,omitempty"`
	SendGlobs                     bool     `json:"send_globs,omitempty"`
	From                          int64    `json:"from,omitempty"`
	Until                         int64    `json:"until,omitempty"`
	Tz                            string   `json:"tz,omitempty"`
	FromRaw                       string   `json:"from_raw,omitempty"`
	UntilRaw                      string   `json:"until_raw,omitempty"`
	Uri                           string   `json:"uri,omitempty"`
	FromCache                     bool     `json:"from_cache"`
	ZipperRequests                int64    `json:"zipper_requests,omitempty"`
}
