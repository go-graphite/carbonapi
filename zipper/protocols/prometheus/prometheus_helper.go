package prometheus

import (
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type tag struct {
	TagValue string
	OP       string
}

type prometheusValue struct {
	Timestamp float64
	Value     float64
}

type prometheusResult struct {
	Metric map[string]string `json:"metric"`
	Values []prometheusValue `json:"values"`
}

type prometheusData struct {
	Result     []prometheusResult `json:"result"`
	ResultType string             `json:"resultType"`
}

type prometheusResponse struct {
	Status    string         `json:"status"`
	ErrorType string         `json:"errorType"`
	Error     string         `json:"error"`
	Data      prometheusData `json:"data"`
}

func (p *prometheusValue) UnmarshalJSON(data []byte) error {
	arr := make([]interface{}, 0)
	err := json.Unmarshal(data, &arr)
	if err != nil {
		return err
	}

	if len(arr) != 2 {
		return fmt.Errorf("length mismatch, got %v, expected 2", len(arr))
	}

	var ok bool
	p.Timestamp, ok = arr[0].(float64)
	if !ok {
		return fmt.Errorf("type mismatch for element[0/1], expected 'float64', got '%T', str=%v", arr[0], string(data))
	}

	str, ok := arr[1].(string)
	if !ok {
		return fmt.Errorf("type mismatch for element[1/1], expected 'string', got '%T', str=%v", arr[1], string(data))
	}

	switch str {
	case "NaN":
		p.Value = math.NaN()
		return nil
	case "+Inf":
		p.Value = math.Inf(1)
		return nil
	case "-Inf":
		p.Value = math.Inf(-1)
		return nil
	default:
		p.Value, err = strconv.ParseFloat(str, 64)
		if err != nil {
			return err
		}
	}

	return nil
}

type prometheusTagResponse struct {
	Status    string   `json:"status"`
	ErrorType string   `json:"errorType"`
	Error     string   `json:"error"`
	Data      []string `json:"data"`
}

type prometheusFindResponse struct {
	Status    string              `json:"status"`
	ErrorType string              `json:"errorType"`
	Error     string              `json:"error"`
	Data      []map[string]string `json:"data"`
}

// accept 'tag=value' or 'tag=~value' string and return sanitized version of it
func (c *PrometheusGroup) promethizeTagValue(tagValue string) (string, tag) {
	// Handle = and =~
	t := tag{}
	idx := strings.Index(tagValue, "=")
	if idx != -1 {
		if tagValue[idx+1] == '~' {
			t.OP = "=~"
			t.TagValue = tagValue[idx+2:]
		} else {
			t.OP = "="
			t.TagValue = tagValue[idx+1:]
		}
	} else {
		// Handle != and !=~
		idx = strings.Index(tagValue, "!")
		if tagValue[idx+2] == '~' {
			t.OP = "!~"
			t.TagValue = tagValue[idx+3:]
		} else {
			t.OP = "!="
			t.TagValue = tagValue[idx+2:]
		}
	}

	return tagValue[:idx], t
}

// TODO: Move to separate package
func (c *PrometheusGroup) splitTagValues(query string) map[string]tag {
	tags := strings.Split(query, ",")
	result := make(map[string]tag)
	for _, tvString := range tags {
		tvString = strings.TrimSpace(tvString)
		name, tag := c.promethizeTagValue(tvString[1 : len(tvString)-1])
		result[name] = tag
	}
	return result
}

func (c *PrometheusGroup) promMetricToGraphite(metric map[string]string) string {
	var res strings.Builder

	res.WriteString(metric["__name__"])
	delete(metric, "__name__")

	keys := make([]string, 0, len(metric))
	for k := range metric {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	for _, k := range keys {
		res.WriteString(";" + k + "=" + metric[k])
	}

	return res.String()
}

// will return step if __step__ is passed
func (c *PrometheusGroup) seriesByTagToPromQL(step, target string) (string, string) {
	firstTag := true
	var queryBuilder strings.Builder
	tagsString := target[len("seriesByTag(") : len(target)-1]
	tvs := c.splitTagValues(tagsString)
	// It's ok to have empty "__name__"
	if v, ok := tvs["__name__"]; ok {
		if v.OP == "=" {
			queryBuilder.WriteString(v.TagValue)
		} else {
			firstTag = false
			queryBuilder.WriteByte('{')
			queryBuilder.WriteString("__name__" + v.OP + "\"" + v.TagValue + "\"")
		}

		delete(tvs, "__name__")
	}
	for tagName, t := range tvs {
		if tagName == "__step__" {
			step = t.TagValue
			continue
		}
		if firstTag {
			firstTag = false
			queryBuilder.WriteByte('{')
			queryBuilder.WriteString(tagName + t.OP + "\"" + t.TagValue + "\"")
		} else {
			queryBuilder.WriteString(", " + tagName + t.OP + "\"" + t.TagValue + "\"")
		}

	}
	if !firstTag {
		queryBuilder.WriteByte('}')
	}
	return step, queryBuilder.String()
}

func convertGraphiteTargetToPromQL(query string) string {
	var sb strings.Builder

	for {
		n := strings.IndexAny(query, "*[{")
		if n < 0 {
			sb.WriteString(regexp.QuoteMeta(query))
			return sb.String()
		}

		sb.WriteString(regexp.QuoteMeta(query[:n]))
		ch := query[n]
		query = query[n+1:]

		switch ch {
		case '*':
			if len(query) == 0 {
				// needed to support find requests when asterisk is the last character and dots should be included
				sb.WriteString(".*")
				break
			}

			sb.WriteString("[^.]*?")

		case '[':
			n = strings.Index(query, "]")
			if n < 0 {
				sb.WriteString(regexp.QuoteMeta("[" + query))
				return sb.String()
			}
			sb.WriteString("[" + query[:n+1])
			query = query[n+1:]

		case '{':
			n = strings.Index(query, "}")
			if n < 0 {
				sb.WriteString(regexp.QuoteMeta("{" + query))
				return sb.String()
			}
			alts := strings.Split(query[:n], ",")
			query = query[n+1:]
			for i := range alts {
				alts[i] = regexp.QuoteMeta(alts[i])
			}
			sb.WriteString("(" + strings.Join(alts, "|") + ")")
		}
	}
}

// inserts math.NaN() in place of gaps in data from Prometheus
func alignValues(startTime, stopTime, step int64, promValues []prometheusValue) []float64 {
	var (
		promValuesCtr = 0
		resValues     = make([]float64, (stopTime-startTime)/step)
	)

	for i := range resValues {
		nextTimestamp := float64(startTime + int64(i)*step)

		if promValuesCtr < len(promValues) && promValues[promValuesCtr].Timestamp == nextTimestamp {
			resValues[i] = promValues[promValuesCtr].Value
			promValuesCtr++
			continue
		}

		resValues[i] = math.NaN()
	}

	return resValues
}
