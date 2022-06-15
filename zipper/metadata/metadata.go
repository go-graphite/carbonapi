package metadata

import (
	"sync"

	"github.com/ansel1/merry"

	"github.com/grafana/carbonapi/limiter"
	"github.com/grafana/carbonapi/zipper/types"
	"go.uber.org/zap"
)

type md struct {
	sync.RWMutex
	SupportedProtocols       map[string]struct{}
	ProtocolInits            map[string]func(*zap.Logger, types.BackendV2, bool) (types.BackendServer, merry.Error)
	ProtocolInitsWithLimiter map[string]func(*zap.Logger, types.BackendV2, bool, limiter.ServerLimiter) (types.BackendServer, merry.Error)
}

var Metadata = md{
	SupportedProtocols:       make(map[string]struct{}),
	ProtocolInits:            make(map[string]func(*zap.Logger, types.BackendV2, bool) (types.BackendServer, merry.Error)),
	ProtocolInitsWithLimiter: make(map[string]func(*zap.Logger, types.BackendV2, bool, limiter.ServerLimiter) (types.BackendServer, merry.Error)),
}
