package metadata

import (
	"sync"

	"github.com/go-graphite/carbonzipper/limiter"
	"github.com/go-graphite/carbonzipper/zipper/errors"
	"github.com/go-graphite/carbonzipper/zipper/types"
	"go.uber.org/zap"
)

type md struct {
	sync.RWMutex
	SupportedProtocols       map[string]struct{}
	ProtocolInits            map[string]func(*zap.Logger, types.BackendV2) (types.ServerClient, *errors.Errors)
	ProtocolInitsWithLimiter map[string]func(*zap.Logger, types.BackendV2, *limiter.ServerLimiter) (types.ServerClient, *errors.Errors)
}

var Metadata = md{
	SupportedProtocols:       make(map[string]struct{}),
	ProtocolInits:            make(map[string]func(*zap.Logger, types.BackendV2) (types.ServerClient, *errors.Errors)),
	ProtocolInitsWithLimiter: make(map[string]func(*zap.Logger, types.BackendV2, *limiter.ServerLimiter) (types.ServerClient, *errors.Errors)),
}
