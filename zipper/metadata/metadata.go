package metadata

import (
	"sync"

	"go.uber.org/zap"

	"github.com/go-graphite/carbonapi/limiter"
	"github.com/go-graphite/carbonapi/zipper/types"
)

type md struct {
	sync.RWMutex
	SupportedProtocols       map[string]struct{}
	ProtocolInits            map[string]func(*zap.Logger, types.BackendV2, bool, bool) (types.BackendServer, error)
	ProtocolInitsWithLimiter map[string]func(*zap.Logger, types.BackendV2, bool, bool, limiter.ServerLimiter) (types.BackendServer, error)
}

var Metadata = md{
	SupportedProtocols:       make(map[string]struct{}),
	ProtocolInits:            make(map[string]func(*zap.Logger, types.BackendV2, bool, bool) (types.BackendServer, error)),
	ProtocolInitsWithLimiter: make(map[string]func(*zap.Logger, types.BackendV2, bool, bool, limiter.ServerLimiter) (types.BackendServer, error)),
}
