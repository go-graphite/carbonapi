package metadata

import (
	"sync"

	"github.com/go-graphite/carbonzipper/limiter"
	"github.com/go-graphite/carbonzipper/zipper/types"
)

type md struct {
	sync.RWMutex
	SupportedProtocols       map[string]struct{}
	ProtocolInits            map[string]func(types.BackendV2) (types.ServerClient, error)
	ProtocolInitsWithLimiter map[string]func(types.BackendV2, limiter.ServerLimiter) (types.ServerClient, error)
}

var Metadata = md{
	SupportedProtocols:       make(map[string]struct{}),
	ProtocolInits:            make(map[string]func(types.BackendV2) (types.ServerClient, error)),
	ProtocolInitsWithLimiter: make(map[string]func(types.BackendV2, limiter.ServerLimiter) (types.ServerClient, error)),
}
