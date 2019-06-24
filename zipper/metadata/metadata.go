package metadata

import (
	"sync"

	"github.com/go-graphite/carbonapi/limiter"
	"github.com/go-graphite/carbonapi/zipper/errors"
	"github.com/go-graphite/carbonapi/zipper/types"
	"go.uber.org/zap"
)

type md struct {
	sync.RWMutex
	SupportedProtocols       map[string]struct{}
	ProtocolInits            map[string]func(*zap.Logger, types.BackendV2) (types.BackendServer, *errors.Errors)
	ProtocolInitsWithLimiter map[string]func(*zap.Logger, types.BackendV2, *limiter.ServerLimiter) (types.BackendServer, *errors.Errors)
}

var Metadata = md{
	SupportedProtocols:       make(map[string]struct{}),
	ProtocolInits:            make(map[string]func(*zap.Logger, types.BackendV2) (types.BackendServer, *errors.Errors)),
	ProtocolInitsWithLimiter: make(map[string]func(*zap.Logger, types.BackendV2, *limiter.ServerLimiter) (types.BackendServer, *errors.Errors)),
}
