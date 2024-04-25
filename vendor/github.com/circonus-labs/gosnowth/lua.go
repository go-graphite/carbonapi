package gosnowth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

// ExtensionParam values contain information about an extension parameter.
type ExtensionParam struct {
	Type        string      `json:"type"`
	Optional    bool        `json:"optional"`
	Default     interface{} `json:"default,omitempty"`
	Description string      `json:"description,omitempty"`
	AliasList   []string    `json:"alias_list,omitempty"`
	Name        string      `json:"name,omitempty"`
}

// LuaExtension values contain information about a Lua extension.
type LuaExtension struct {
	Documentation    string                     `json:"documentation,omitempty"`
	Method           string                     `json:"method"`
	ParseJSONPayload bool                       `json:"PARSE_JSON_PAYLOAD,omitempty"`
	Params           map[string]*ExtensionParam `json:"params"`
	Man              string                     `json:"man,omitempty"`
	Name             string                     `json:"name,omitempty"`
	Description      string                     `json:"description"`
}

// LuaExtensions is a map of information about Lua extensions.
type LuaExtensions map[string]*LuaExtension

// UnmarshalJSON decodes a byte slice of JSON data into a LuaExtensions map.
func (le *LuaExtensions) UnmarshalJSON(b []byte) error { //nolint:gocyclo
	r := map[string]interface{}{}

	err := json.NewDecoder(bytes.NewBuffer(b)).Decode(&r)
	if err != nil {
		return err
	}

	*le = make(map[string]*LuaExtension, len(r))

	for k, ext := range r {
		(*le)[k] = &LuaExtension{Name: k}

		if v, ok := ext.(map[string]interface{}); ok { //nolint:nestif
			for key, val := range v {
				switch key {
				case "documentation":
					if value, ok := val.(string); ok {
						(*le)[k].Documentation = value
					}
				case "method":
					if value, ok := val.(string); ok {
						(*le)[k].Method = value
					}
				case "PARSE_JSON_PAYLOAD":
					if value, ok := val.(bool); ok {
						(*le)[k].ParseJSONPayload = value
					}
				case "params":
					if value, ok := val.(map[string]interface{}); ok {
						(*le)[k].Params = make(map[string]*ExtensionParam,
							len(value))

						for pk, pv := range value {
							if pMap, ok := pv.(map[string]interface{}); ok {
								(*le)[k].Params[pk] = &ExtensionParam{Name: pk}

								for pKey, pVal := range pMap {
									switch pKey {
									case "type":
										if pV, ok := pVal.(string); ok {
											(*le)[k].Params[pk].Type = pV
										}
									case "optional":
										if pV, ok := pVal.(bool); ok {
											(*le)[k].Params[pk].Optional = pV
										}
									case "default":
										(*le)[k].Params[pk].Default = pVal
									case "description":
										if pV, ok := pVal.(string); ok {
											(*le)[k].Params[pk].Description = pV
										}
									case "alias_list":
										if pV, ok := pVal.([]interface{}); ok {
											(*le)[k].Params[pk].AliasList = make([]string, len(pV))

											for n, iV := range pV {
												if itV, ok := iV.(string); ok {
													(*le)[k].Params[pk].
														AliasList[n] = itV
												}
											}
										}
									case "name":
										if pV, ok := pVal.(string); ok {
											(*le)[k].Params[pk].Name = pV
										}
									}
								}
							}
						}
					}
				case "man":
					if value, ok := val.(string); ok {
						(*le)[k].Man = value
					}
				case "name":
					if value, ok := val.(string); ok {
						(*le)[k].Name = value
					}
				case "description":
					if value, ok := val.(string); ok {
						(*le)[k].Description = value
					}
				}
			}
		}
	}

	return nil
}

// GetLuaExtensions retrieves information about available Lua extensions.
func (sc *SnowthClient) GetLuaExtensions(nodes ...*SnowthNode) (LuaExtensions,
	error,
) {
	return sc.GetLuaExtensionsContext(context.Background(), nodes...)
}

// GetLuaExtensionsContext is the context aware version of GetLuaExtensions.
func (sc *SnowthClient) GetLuaExtensionsContext(ctx context.Context,
	nodes ...*SnowthNode,
) (LuaExtensions, error) {
	var node *SnowthNode
	if len(nodes) > 0 && nodes[0] != nil {
		node = nodes[0]
	} else {
		node = sc.GetActiveNode()
	}

	if node == nil {
		return nil, fmt.Errorf("unable to get active node")
	}

	u := "/extension/lua"
	r := LuaExtensions{}

	body, _, err := sc.DoRequestContext(ctx, node, "GET", u, nil, nil)
	if err != nil {
		return nil, err
	}

	if err := decodeJSON(body, &r); err != nil {
		return nil, fmt.Errorf("unable to decode IRONdb response: %w", err)
	}

	return r, err
}

// ExtParam values contain parameter values to use when executing extensions.
type ExtParam struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// ExecLuaExtension executes the specified Lua extension and returns the
// response as a JSON map.
func (sc *SnowthClient) ExecLuaExtension(name string,
	params []ExtParam, nodes ...*SnowthNode,
) (map[string]interface{}, error) {
	return sc.ExecLuaExtensionContext(context.Background(), name,
		params, nodes...)
}

// ExecLuaExtensionContext is the context aware version of ExecLuaExtension.
func (sc *SnowthClient) ExecLuaExtensionContext(ctx context.Context,
	name string, params []ExtParam,
	nodes ...*SnowthNode,
) (map[string]interface{}, error) {
	var node *SnowthNode
	if len(nodes) > 0 && nodes[0] != nil {
		node = nodes[0]
	} else {
		node = sc.GetActiveNode()
	}

	if node == nil {
		return nil, fmt.Errorf("unable to get active node")
	}

	u := "/extension/lua/" + name

	if len(params) > 0 {
		qp := url.Values{}
		for _, p := range params {
			qp.Add(p.Name, p.Value)
		}

		u += "?" + qp.Encode()
	}

	r := map[string]interface{}{}

	body, _, err := sc.DoRequestContext(ctx, node, "GET", u, nil, nil)
	if err != nil {
		return nil, err
	}

	if err := decodeJSON(body, &r); err != nil {
		return nil, fmt.Errorf("unable to decode IRONdb response: %w", err)
	}

	return r, err
}
