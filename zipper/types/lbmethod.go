package types

import (
	"encoding/json"
	"fmt"
	"strings"
)

var ErrUnknownLBMethodFmt = "unknown lb method: '%v', supported: %v"

type LBMethod int

const (
	RoundRobinLB LBMethod = iota
	BroadcastLB
)

func (p LBMethod) keys(m map[string]LBMethod) []string {
	res := make([]string, 0)
	for k := range m {
		res = append(res, k)
	}
	return res
}

var supportedLBMethods = map[string]LBMethod{
	"roundrobin": RoundRobinLB,
	"rr":         RoundRobinLB,
	"any":        RoundRobinLB,
	"broadcast":  BroadcastLB,
	"all":        BroadcastLB,
}

func (m *LBMethod) FromString(method string) error {
	var ok bool
	if *m, ok = supportedLBMethods[strings.ToLower(method)]; !ok {
		return fmt.Errorf(ErrUnknownLBMethodFmt, method, m.keys(supportedLBMethods))
	}
	return nil
}

func (m *LBMethod) UnmarshalJSON(data []byte) error {
	method := strings.ToLower(string(data))
	return m.FromString(method)
}

func (m *LBMethod) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var method string
	err := unmarshal(&method)
	if err != nil {
		return err
	}

	return m.FromString(method)
}

func (m LBMethod) MarshalJSON() ([]byte, error) {
	switch m {
	case RoundRobinLB:
		return json.Marshal("RoundRobin")
	case BroadcastLB:
		return json.Marshal("Broadcast")
	}

	return nil, fmt.Errorf(ErrUnknownLBMethodFmt, m, m.keys(supportedLBMethods))
}
