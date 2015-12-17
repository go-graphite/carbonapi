package main

import (
	"bytes"
	"math"
	"strconv"
	"time"

	pb "github.com/dgryski/carbonzipper/carbonzipperpb"
	pickle "github.com/kisielk/og-rek"
)

type metricData struct {
	pb.FetchResponse

	// extra options
	xStep          float64
	valuesPerPoint int
	drawAsInfinite bool
	secondYAxis    bool
	dashed         bool // TODO (ikruglov) smth like lineType would be better
	color          string
	lineWidth      float64
}

func marshalCSV(results []*metricData) []byte {

	var b []byte

	for _, r := range results {

		step := r.GetStepTime()
		t := r.GetStartTime()
		for i, v := range r.Values {
			if !r.IsAbsent[i] {
				b = append(b, '"')
				b = append(b, r.GetName()...)
				b = append(b, '"')
				b = append(b, ',')
				b = append(b, time.Unix(int64(t), 0).Format("2006-01-02 15:04:05")...)
				b = append(b, ',')
				b = strconv.AppendFloat(b, v, 'f', -1, 64)
				b = append(b, '\n')
			}
			t += step
		}
	}
	return b
}

func marshalJSON(results []*metricData) []byte {

	var b []byte
	b = append(b, '[')

	var topComma bool
	for _, r := range results {
		if r == nil {
			continue
		}

		if topComma {
			b = append(b, ',')
		}
		topComma = true

		b = append(b, `{"target":`...)
		b = strconv.AppendQuoteToASCII(b, r.GetName())
		b = append(b, `,"datapoints":[`...)

		var innerComma bool
		t := r.GetStartTime()
		for i, v := range r.Values {
			if innerComma {
				b = append(b, ',')
			}
			innerComma = true

			b = append(b, '[')

			if r.IsAbsent[i] || math.IsInf(v, 0) {
				b = append(b, "null"...)
			} else {
				b = strconv.AppendFloat(b, v, 'f', -1, 64)
			}

			b = append(b, ',')

			b = strconv.AppendInt(b, int64(t), 10)

			b = append(b, ']')

			t += r.GetStepTime()
		}

		b = append(b, `]}`...)
	}

	b = append(b, ']')

	return b
}

func marshalPickle(results []*metricData) []byte {

	var p []map[string]interface{}

	for _, r := range results {
		values := make([]interface{}, len(r.Values))
		for i, v := range r.Values {
			if r.IsAbsent[i] {
				values[i] = pickle.None{}
			} else {
				values[i] = v
			}

		}
		p = append(p, map[string]interface{}{
			"name":   r.GetName(),
			"start":  r.GetStartTime(),
			"end":    r.GetStopTime(),
			"step":   r.GetStepTime(),
			"values": values,
		})
	}

	var buf bytes.Buffer

	penc := pickle.NewEncoder(&buf)
	penc.Encode(p)

	return buf.Bytes()
}

func marshalProtobuf(results []*metricData) []byte {
	response := pb.MultiFetchResponse{}
	for _, metric := range results {
		response.Metrics = append(response.Metrics, &((*metric).FetchResponse))
	}
	b, err := response.Marshal()
	if err != nil {
		logger.Logf("proto.Marshal: %v", err)
	}

	return b
}

func marshalRaw(results []*metricData) []byte {

	var b []byte

	for _, r := range results {

		b = append(b, r.GetName()...)

		b = append(b, ',')
		b = strconv.AppendInt(b, int64(r.GetStartTime()), 10)
		b = append(b, ',')
		b = strconv.AppendInt(b, int64(r.GetStopTime()), 10)
		b = append(b, ',')
		b = strconv.AppendInt(b, int64(r.GetStepTime()), 10)
		b = append(b, '|')

		var comma bool
		for i, v := range r.Values {
			if comma {
				b = append(b, ',')
			}
			comma = true
			if r.IsAbsent[i] {
				b = append(b, "None"...)
			} else {
				b = strconv.AppendFloat(b, v, 'f', -1, 64)
			}
		}

		b = append(b, '\n')
	}
	return b
}
