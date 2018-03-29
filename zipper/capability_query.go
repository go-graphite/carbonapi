package zipper

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/go-graphite/carbonzipper/limiter"
	"github.com/go-graphite/carbonzipper/zipper/helper"
	"github.com/go-graphite/carbonzipper/zipper/httpHeaders"
	protov3 "github.com/go-graphite/protocol/carbonapi_v3_pb"
	"go.uber.org/zap"
)

type capabilityResponse struct {
	server   string
	protocol string
}

//_internal/capabilities/
func doQuery(query *helper.HttpQuery, server string, payload []byte, resChan chan<- capabilityResponse) {

}

type CapabilityResponse struct {
	ProtoToServers map[string][]string
}

func getBestSupportedProtocol(logger *zap.Logger, servers []string, concurencyLimit int) *CapabilityResponse {
	response := &CapabilityResponse{
		ProtoToServers: make(map[string][]string),
	}
	groupName := "capability query"
	limiter := limiter.NewServerLimiter([]string{groupName}, concurencyLimit)

	httpClient := &http.Client{
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				// TODO: Make that configurable
				Timeout:   200 * time.Millisecond,
				KeepAlive: 30 * time.Second,
				DualStack: true,
			}).DialContext,
		},
	}

	httpQuery := helper.NewHttpQuery(logger, groupName, servers, 3, limiter, httpClient, httpHeaders.ContentTypeCarbonAPIv3PB)

	request := protov3.CapabilityRequest{}
	data, err := request.Marshal()
	if err != nil {
		return nil
	}

	resCh := make(chan capabilityResponse, len(servers))

	for _, srv := range servers {
		go doQuery(httpQuery, srv, data, resCh)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	answeredServers := make(map[string]struct{})
	responseCounts := 0
GATHER:
	for {
		if responseCounts == len(servers) && len(resCh) == 0 {
			break GATHER
		}
		select {
		case res := <-resCh:
			responseCounts++
			answeredServers[res.server] = struct{}{}
			p := response.ProtoToServers[res.protocol]
			response.ProtoToServers[res.protocol] = append(p, res.protocol)
		case <-ctx.Done():
			noAnswer := make([]string, 0)
			for _, s := range servers {
				if _, ok := answeredServers[s]; !ok {
					noAnswer = append(noAnswer, s)
				}
			}
			logger.Warn("timeout waiting for more responses",
				zap.Strings("no_answers_from", noAnswer),
			)
			break GATHER
		}
	}

	return response
}
