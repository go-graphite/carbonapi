package stringutils

import (
	"fmt"
	"strconv"
	"strings"
)

type templateParam []string

func initTemplateParam(s string) templateParam {
	v := []string{s}
	return append(v, strings.Split(s, ".")...)
}

// Template parsed and splited format string (stored in first field)
type Template []interface{}

// InitTemplate parse and split format string (format string stored in first field)
//
// @format Format string like 'string %{param} %{param1.param2}'
func InitTemplate(format string) (Template, error) {
	t := make([]interface{}, 1, 32)
	t[0] = format
	f := format
	for {
		start := strings.IndexByte(f, '%')
		if start == -1 {
			t = append(t, f)
			break
		} else if f[start] == '%' {
			// Try to extract parameter
			if len(f) < start+2 {
				return nil, fmt.Errorf("parse error '%s' at symbol %d: unexpected end", f, start)
			} else if f[start+1] != '{' {
				return nil, fmt.Errorf("parse error '%s' at symbol %d: unexpected %c", f, start+1, f[start+1])
			}
			if start > 0 {
				t = append(t, f[:start])
			}
			f = f[start+2:]
			// Try to find end parameter delimiter '}'
			end := strings.IndexByte(f, '}')
			if end == -1 {
				return nil, fmt.Errorf("parse error '%s': expect }", f)
			}
			name := f[:end]
			t = append(t, initTemplateParam(name))

			if end+1 == len(f) {
				break
			}
			// next iter
			f = f[end+1:]
		}
	}
	return t, nil
}

func lookupParam(p templateParam, params map[string]interface{}) (string, error) {
	var (
		v  interface{}
		ok bool
	)
	lp := len(p) - 1
	mv := params
	for i := 1; i < len(p); i++ {
		if v, ok = mv[p[i]]; ok {
			if i == lp {
				// end of templateParam, so check param value
				switch n := v.(type) {
				case string:
					return n, nil
				case float64:
					return strconv.FormatFloat(n, 'f', -1, 64), nil
				case float32:
					return strconv.FormatFloat(float64(n), 'f', -1, 32), nil
				case int32:
					return strconv.FormatInt(int64(n), 10), nil
				case uint32:
					return strconv.FormatUint(uint64(n), 10), nil
				case int64:
					return strconv.FormatInt(n, 10), nil
				case uint64:
					return strconv.FormatUint(n, 10), nil
				case int16:
					return strconv.FormatInt(int64(n), 10), nil
				case uint16:
					return strconv.FormatUint(uint64(n), 10), nil
				case int8:
					return strconv.FormatInt(int64(n), 10), nil
				case uint8:
					return strconv.FormatUint(uint64(n), 10), nil
				case int:
					return strconv.FormatInt(int64(n), 10), nil
				case uint:
					return strconv.FormatUint(uint64(n), 10), nil
				default:
					return "", fmt.Errorf("incorect field type when lookup template param field %s (%s): '%+v'", p[i], p[0], v)
				}
			}
			if mv, ok = v.(map[string]interface{}); !ok {
				return "", fmt.Errorf("unexpected end of params map when template field lookup %s (%s): '%+v'", p[i], p[0], v)
			}
		}
	}

	return "", fmt.Errorf("unknown field in template param: %+v", p[0])
}

func loopkupTemplateNode(t interface{}, params map[string]interface{}) (string, error) {
	if s, ok := t.(string); ok {
		return s, nil
	} else if p, ok := t.(templateParam); ok {
		return lookupParam(p, params)
	} else {
		return "", fmt.Errorf("unknown field type: %+v", t)
	}
}

// Execute process template with mapped params
//
// @Params Params in map[string]interface{}
//
// 	params := map[string]interface{}{
// 		"param":  "URL",
// 		"param1": map[string]interface{}{ "param2": "2" },
// 	}
func (t *Template) Execute(params map[string]interface{}) (string, error) {
	if len(*t) == 2 {
		return loopkupTemplateNode((*t)[1], params)
	} else if len(*t) > 2 {
		var sb Builder
		sb.Grow(2 * len((*t)[0].(string)))
		for i := 1; i < len(*t); i++ {
			if s, err := loopkupTemplateNode((*t)[i], params); err == nil {
				sb.WriteString(s)
			} else {
				return "", err
			}
		}
		return sb.String(), nil
	}
	return "", nil
}
