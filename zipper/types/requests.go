package types

import (
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
	return request.MultiGlobRequest.Marshal()
}

func (request MultiGlobRequestV3) LogInfo() interface{} {
	return request.MultiGlobRequest
}

func (request MultiFetchRequestV3) Marshal() ([]byte, error) {
	return request.MultiFetchRequest.Marshal()
}

func (request MultiFetchRequestV3) LogInfo() interface{} {
	return request.MultiFetchRequest

}

func (request MultiMetricsInfoV3) Marshal() ([]byte, error) {
	return request.MultiMetricsInfoRequest.Marshal()
}

func (request MultiMetricsInfoV3) LogInfo() interface{} {
	return request.MultiMetricsInfoRequest
}

func (request CapabilityRequestV3) Marshal() ([]byte, error) {
	return request.CapabilityRequest.Marshal()
}

func (request CapabilityRequestV3) LogInfo() interface{} {
	return request.CapabilityRequest
}
