package types

import protov3 "github.com/go-graphite/protocol/carbonapi_v3_pb"

type MultiFetchRequestV3 struct {
	*protov3.MultiFetchRequest
}

type MultiGlobRequestV3 struct {
	*protov3.MultiGlobRequest
}

type MultiMetricsInfoV3 struct {
	*protov3.MultiMetricsInfoRequest
}

type CapabilityRequestV3 struct {
	*protov3.CapabilityRequest
}

func (request MultiGlobRequestV3) Marshal() ([]byte, error) {
	return request.MultiGlobRequest.Marshal()
}

func (request MultiGlobRequestV3) LogInfo() string {
	return request.MultiGlobRequest.GoString()
}

func (request MultiFetchRequestV3) Marshal() ([]byte, error) {
	return request.MultiFetchRequest.Marshal()
}

func (request MultiFetchRequestV3) LogInfo() string {
	return request.MultiFetchRequest.GoString()

}

func (request MultiMetricsInfoV3) Marshal() ([]byte, error) {
	return request.MultiMetricsInfoRequest.Marshal()
}

func (request MultiMetricsInfoV3) LogInfo() string {
	return request.MultiMetricsInfoRequest.GoString()
}

func (request CapabilityRequestV3) Marshal() ([]byte, error) {
	return request.CapabilityRequest.Marshal()
}

func (request CapabilityRequestV3) LogInfo() string {
	return request.CapabilityRequest.GoString()
}
