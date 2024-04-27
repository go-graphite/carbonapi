package types

import (
	"github.com/ansel1/merry"
	protov3 "github.com/go-graphite/protocol/carbonapi_v3_pb"
)

type MultiFetchRequestV3 struct {
	protov3.MultiFetchRequest
}

type MultiGlobRequestV3 struct {
	protov3.MultiGlobRequest
}

type MultiMetricsInfoV3 struct {
	protov3.MultiMetricsInfoRequest
}

type CapabilityRequestV3 struct {
	protov3.CapabilityRequest
}

func (request MultiGlobRequestV3) Marshal() ([]byte, error) {
	b, err := request.MultiGlobRequest.Marshal()
	return b, merry.Wrap(err)
}

func (request MultiGlobRequestV3) LogInfo() interface{} {
	return request.MultiGlobRequest
}

func (request MultiFetchRequestV3) Marshal() ([]byte, error) {
	b, err := request.MultiFetchRequest.Marshal()
	return b, merry.Wrap(err)
}

func (request MultiFetchRequestV3) LogInfo() interface{} {
	return request.MultiFetchRequest

}

func (request MultiMetricsInfoV3) Marshal() ([]byte, error) {
	b, err := request.MultiMetricsInfoRequest.Marshal()
	return b, merry.Wrap(err)
}

func (request MultiMetricsInfoV3) LogInfo() interface{} {
	return request.MultiMetricsInfoRequest
}

func (request CapabilityRequestV3) Marshal() ([]byte, error) {
	b, err := request.CapabilityRequest.Marshal()
	return b, merry.Wrap(err)
}

func (request CapabilityRequestV3) LogInfo() interface{} {
	return request.CapabilityRequest
}
