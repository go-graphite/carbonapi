package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"strings"

	"github.com/go-graphite/carbonapi/expr/types"
)

func isUnsortedStringSlicesEqual(s1, s2 []string) bool {
	if len(s1) != len(s2) {
		return false
	}

	s1Map := make(map[string]struct{})

	for _, v := range s1 {
		s1Map[v] = struct{}{}
	}

	for _, v := range s2 {
		if _, ok := s1Map[v]; !ok {
			return false
		}
	}
	return true
}

func isFunctionParamEqual(fp1, fp2 types.FunctionParam) []string {
	var incompatibilities []string
	if !isUnsortedStringSlicesEqual(fp1.Options, fp2.Options) {
		// TODO(civil): Distingush and flag supersets (where we support more)
		if len(fp1.Options) < len(fp2.Options) {
			incompatibilities = append(incompatibilities, fmt.Sprintf("%v: different amount of parameters, got `%+v`, should be `%+v`", fp1.Name, fp1.Options, fp2.Options))
		}
	}

	if fp1.Name != fp2.Name {
		incompatibilities = append(incompatibilities, fmt.Sprintf("%v: name mismatch: got %v, should be %v", fp1.Name, fp1.Name, fp2.Name))
	}

	if fp1.Multiple != fp2.Multiple {
		incompatibilities = append(incompatibilities, fmt.Sprintf("%v: attribute `multiple` mismatch: got %v, should be %v", fp1.Name, fp1.Multiple, fp2.Multiple))
	}

	if fp1.Type != fp2.Type {
		v1, _ := fp1.Type.MarshalJSON()
		v2, _ := fp2.Type.MarshalJSON()
		incompatibilities = append(incompatibilities, fmt.Sprintf("%v: type mismatch: got %v, should be %v", fp1.Name, string(v1), string(v2)))
	}

	if fp1.Default != nil && fp2.Default != nil {
		if fp1.Default.Type != fp2.Default.Type {
			incompatibilities = append(incompatibilities, fmt.Sprintf("%v: default value's type mismatch: got %v, should be %v", fp1.Name, fp1.Default.Type, fp2.Default.Type))
		}
		v1, _ := fp1.Default.MarshalJSON()
		v2, _ := fp2.Default.MarshalJSON()
		if !bytes.Equal(v1, v2) {
			incompatibilities = append(incompatibilities, fmt.Sprintf("%v: default value mismatch: got %v, should be %v", fp1.Name, string(v1), string(v2)))
		}
	}

	if fp1.Default == nil && fp2.Default != nil {
		v2, _ := fp2.Default.MarshalJSON()
		incompatibilities = append(incompatibilities, fmt.Sprintf("%v: default value mismatch: got %v, should be %v", fp1.Name, "(empty)", string(v2)))
	}

	if fp1.Default != nil && fp2.Default == nil {
		v1, _ := fp1.Default.MarshalJSON()
		incompatibilities = append(incompatibilities, fmt.Sprintf("%v: default value mismatch: got %v, should be %v", fp1.Name, string(v1), "(empty)"))
	}
	return incompatibilities
}

// type FunctionParam struct {
// 	Name        string        `json:"name"`
// 	Multiple    bool          `json:"multiple,omitempty"`
// 	Required    bool          `json:"required,omitempty"`
// 	Type        FunctionType  `json:"type,omitempty"`
// 	Options     []string      `json:"options,omitempty"`
// 	Suggestions []*Suggestion `json:"suggestions,omitempty"`
// 	Default     *Suggestion   `json:"default,omitempty"`
// }
func isFunctionParamsEqual(list1, list2 []types.FunctionParam) []string {
	list1ToMap := make(map[string]types.FunctionParam)
	var incompatibilities []string

	for _, v := range list1 {
		list1ToMap[v.Name] = v
	}

	for _, fp2 := range list2 {
		fp1, ok := list1ToMap[fp2.Name]
		if !ok {
			incompatibilities = append(incompatibilities, fmt.Sprintf("parameter not supported: %v", fp2.Name))
			continue
		}

		incompatibility := isFunctionParamEqual(fp1, fp2)
		if len(incompatibility) != 0 {
			incompatibilities = append(incompatibilities, incompatibility...)
		}
	}

	return incompatibilities
}

func main() {
	srv1 := flag.String("carbonapi", "http://localhost:8079", "first server base url")
	srv2 := flag.String("graphiteweb", "http://localhost:8082", "second server base url")

	flag.Parse()

	res, err := http.Get(*srv1 + "/functions/")
	if err != nil {
		log.Fatal("failed to get response from server 1", err)
	}

	resp1, err := ioutil.ReadAll(res.Body)
	res.Body.Close()

	var firstDescription map[string]types.FunctionDescription

	err = json.Unmarshal(resp1, &firstDescription)
	if err != nil {
		log.Fatal("failed to Unmarshal first description", err)
	}

	res, err = http.Get(*srv2 + "/functions/")
	if err != nil {
		log.Fatal("failed to get response from server 1", err)
	}

	resp2, err := ioutil.ReadAll(res.Body)
	res.Body.Close()

	var secondDescription map[string]types.FunctionDescription

	err = json.Unmarshal(resp2, &secondDescription)
	if err != nil {
		log.Fatal("failed to Unmarshal second description", err)
	}

	var carbonapiFunctions []string
	var supportedFunctions []string
	functionsWithIncompatibilities := make(map[string][]string)
	var unsupportedFunctions []string

	for k, v := range secondDescription {
		if v2, ok := firstDescription[k]; ok && !v.Proxied {
			incompatibilities := isFunctionParamsEqual(v.Params, v2.Params)
			supportedFunctions = append(supportedFunctions, v.Function)
			if len(incompatibilities) != 0 {
				functionsWithIncompatibilities[k] = incompatibilities
			}
		} else {
			unsupportedFunctions = append(unsupportedFunctions, k)
		}
	}

	for k, v := range firstDescription {
		if _, ok := secondDescription[k]; !ok {
			carbonapiFunctions = append(carbonapiFunctions, v.Function)
		}
	}

	sort.Strings(carbonapiFunctions)
	sort.Strings(unsupportedFunctions)
	sort.Strings(supportedFunctions)

	fmt.Printf(`# CarbonAPI compatibility with Graphite

Topics:
* [Default settings](#default-settings)
* [URI Parameters](#uri-params)
* [Graphite-web 1.1 Compatibility](#graphite-web-11-compatibility)
* [Supported Functions](#supported-functions)
* [Features of configuration functions](#functions-features)

<a name="default-settings"></a>
## Default Settings

### Default Line Colors
Default colors for png or svg rendering intentionally specified like it is in graphite-web 1.1.0

You can redefine that in config to be more more precise. In default config example they are defined in the same way as in [original graphite PR to make them right](https://github.com/graphite-project/graphite-web/pull/2239)

Reason behind that change is that on dark background it's much nicer to read old colors than new one

<a name="uri-params"></a>
## URI Parameters

### /render/?...

* ` + "`target` : graphite series, seriesList or function (likely containing series or seriesList)\n" +
		"* `from`, `until` : time specifiers. Eg. \"1d\", \"10min\", \"04:37_20150822\", \"now\", \"today\", ... (**NOTE** does not handle timezones the same as graphite)\n" +
		"* `format` : support graphite values of { json, raw, pickle, csv, png, svg } adds { protobuf } and does not support { pdf }\n" +
		"* `jsonp` : (...)\n" +
		"* `noCache` : prevent query-response caching (which is 60s if enabled)\n" +
		"* `cacheTimeout` : override default result cache (60s)\n" +
		"* `rawdata` -or- `rawData` : true for `format=raw`\n" + `
**Explicitly NOT supported**
* ` + "`_salt`\n" +
		"* `_ts`\n" +
		"* `_t`\n" + `
_When ` + "`format=png`_ (default if not specified)\n" +
		"* `width`, `height` : number of pixels (default: width=330 , height=250)\n" +
		"* `pixelRatio` : (1.0)\n" +
		"* `margin` : (10)\n" +
		"* `logBase` : Y-scale should use. Recognizes \"e\" or a floating point ( >= 1 )\n" +
		"* `fgcolor` : foreground color\n" +
		"* `bgcolor` : background color\n" +
		"* `majorLine` : major line color\n" +
		"* `minorLine` : minor line color\n" +
		"* `fontName` : (\"Sans\")\n" +
		"* `fontSize` : (10.0)\n" +
		"* `fontBold` : (false)\n" +
		"* `fontItalic` : (false)\n" +
		"* `graphOnly` : (false)\n" +
		"* `hideLegend` : (false) (**NOTE** if not defined and >10 result metrics this becomes true)\n" +
		"* `hideGrid` : (false)\n" +
		"* `hideAxes` : (false)\n" +
		"* `hideYAxis` : (false)\n" +
		"* `hideXAxis` : (false)\n" +
		"* `yAxisSide` : (\"left\")\n" +
		"* `connectedLimit` : number of missing points to bridge when `linemode` is not one of { \"slope\", \"staircase\" } likely \"connected\" (4294967296)\n" +
		"* `lineMode` : (\"slope\")\n" +
		"* `areaMode` : (\"none\") also recognizes { \"first\", \"all\", \"stacked\" }\n" +
		"* `areaAlpha` : ( <not defined> ) float value for area alpha\n" +
		"* `pieMode` : (\"average\") also recognizes { \"maximum\", \"minimum\" } (**NOTE** pie graph support is explicitly unplanned)\n" +
		"* `lineWidth` : (1.2) float value for line width\n" +
		"* `dashed` : (false) dashed lines\n" +
		"* `rightWidth` : (1.2) ...\n" +
		"* `rightDashed` : (false)\n" +
		"* `rightColor` : ...\n" +
		"* `leftWidth` : (1.2)\n" +
		"* `leftDashed` : (false)\n" +
		"* `leftColor` : ...\n" +
		"* `title` : (\"\") graph title\n" +
		"* `vtitle` : (\"\") ...\n" +
		"* `vtitleRight` : (\"\") ...\n" +
		"* `colorList` : (\"blue,green,red,purple,yellow,aqua,grey,magenta,pink,gold,rose\")\n" +
		"* `majorGridLineColor` : (\"rose\")\n" +
		"* `minorGridLineColor` : (\"grey\")\n" +
		"* `uniqueLegend` : (false)\n" +
		"* `drawNullAsZero` : (false) (**NOTE** affects display only - does not translate missing values to zero in functions. For that use ...)\n" +
		"* `drawAsInfinite` : (false) ...\n" +
		"* `yMin` : <undefined>\n" +
		"* `yMax` : <undefined>\n" +
		"* `yStep` : <undefined>\n" +
		"* `xMin` : <undefined>\n" +
		"* `xMax` : <undefined>\n" +
		"* `xStep` : <undefined>\n" +
		"* `xFormat` : (\"\") ...\n" +
		"* `minorY` : (1) ...\n" +
		"* `yMinLeft` : <undefined>\n" +
		"* `yMinRight` : <undefined>\n" +
		"* `yMaxLeft` : <undefined>\n" +
		"* `yMaxRight` : <undefined>\n" +
		"* `yStepL` : <undefined>\n" +
		"* `ySTepR` : <undefined>\n" +
		"* `yLimitLeft` : <undefined>\n" +
		"* `yLimitRight` : <undefined>\n" +
		"* `yUnitSystem` : (\"si\") also recognizes { \"binary\" }\n" +
		"* `yDivisors` : (4,5,6) ...\n" + `
### /metrics/find/?

* ` + "`format` : (\"treejson\") also recognizes { \"json\" (same as \"treejson\"), \"completer\", \"raw\" }\n" +
		"* `jsonp` : ...\n" +
		"* `query` : the metric or glob-pattern to find\n" + `

`)
	fmt.Println(`

## Graphite-web 1.1 compatibility
### Unsupported functions
| Function                                                                  |
| :------------------------------------------------------------------------ |`)
	for _, f := range unsupportedFunctions {
		fmt.Printf("| %v |\n", f)
	}

	fmt.Println(`

### Partly supported functions
| Function                 | Incompatibilities                              |
| :------------------------|:---------------------------------------------- |`)

	keys := make([]string, 0, len(functionsWithIncompatibilities))
	for k := range functionsWithIncompatibilities {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, f := range keys {
		fmt.Printf("| %v | %v |\n", f, strings.Join(functionsWithIncompatibilities[f], "\n"))
	}

	fmt.Println(`
## Supported functions
| Function      | Carbonapi-only                                            |
| :-------------|:--------------------------------------------------------- |`)

	for _, f := range supportedFunctions {
		fmt.Printf("| %v | no |\n", f)
	}

	for _, f := range carbonapiFunctions {
		fmt.Printf("| %v | yes |\n", f)
	}

	fmt.Println(`<a name="functions-features"></a>
## Features of configuration functions
### aliasByPostgres
1. Make config for function with pairs key-string - request
` + "```" + `yaml
enabled: true
database:
  "databaseAlias":
    urlDB: "localhost:5432"
    username: "portgres_user"
    password: "postgres_password"
    nameDB: "database_name"
    keyString:
      "resolve_switch_name_byId":
        varName: "var"
        queryString: "SELECT field_with_switch_name FROM some_table_with_switch_names_id_and_other WHERE field_with_switchID like 'var0';"
        matchString: ".*"
      "resolve_interface_description_from_table":
        varName: "var"
        queryString: "SELECT interface_desc FROM some_table_with_switch_data WHERE field_with_hostname like 'var0' AND field_with_interface_id like 'var1';"
        matchString: ".*"
` + "```" + `

#### Examples

We have data series:
` + "```" + `
switches.switchId.CPU1Min
` + "```" + `
We need to get CPU load resolved by switchname, aliasByPostgres( switches.*.CPU1Min, databaseAlias, resolve_switch_name_byId, 1 ) will return series like this:
` + "```" + `
switchnameA
switchnameB
switchnameC
switchnameD
` + "```" + `
We have data series:
` + "```" + `
switches.hostname.interfaceID.scope_of_interface_metrics
` + "```" + `
We want to see interfaces stats sticked to their descriptions, aliasByPostgres(switches.hostname.*.ifOctets.rx, databaseAlias, resolve_interface_description_from_table, 1, 2 )
will return series:
` + "```" + `
InterfaceADesc
InterfaceBDesc
InterfaceCDesc
InterfaceDDesc
` + "```" + `

2. Add to main config path to configuration file
` + "```" + `yaml
functionsConfigs:
        aliasByPostgres: /path/to/funcConfig.yaml
` + "```" + `
-----`)

}
