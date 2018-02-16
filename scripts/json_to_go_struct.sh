#!/bin/bash

# Disclaimer:
#   This code is veeeeery ugly, but it works. If it hurt your feelings, please rewrite it in some normal language.
#   I'm truly sorry that I decided to write that in bash and I really hope that next time I'll make better descision
#
# Main idea of this file is to get graphite-web's json and autogenerate description as a Go-struct.

usage() {
	echo "${0} func_list_output.json function_dir code_path"
	echo
	echo "Main idea of this file is to get graphite-web's json and autogenerate description as a Go-struct."
	echo "Example: ./${0} ./func.json below ~/go/gopath_third_party/src/github.com/go-graphite/carbonapi"
}

JSON_FILE=${1}
if [[ -z ${JSON_FILE} ]]; then
	usage
	exit 1
fi
shift
FUNCTION_FILE="${1}"
if [[ -z ${FUNCTION_FILE} ]]; then
	usage
	exit 1
fi

shift
# ~/go/gopath_third_party/src/github.com/go-graphite/carbonapi
CODE_PATH="${1}"
if [[ -z ${CODE_PATH} ]]; then
	usage
	exit 1
fi

shift

FUNCTIONS=$(egrep 'RegisterFunction|functions :=' "${CODE_PATH}"/expr/functions/"${FUNCTION_FILE}"/function.go  | grep -v 'RegisterFunction(f,'  | egrep -o '"[^"]+"' | tr -d '"')

{
echo
echo "// Description is auto-generated description, based on output of https://github.com/graphite-project/graphite-web"
echo "func (f *${FUNCTION_FILE}) Description() map[string]*types.FunctionDescription {"
echo "return map[string]*types.FunctionDescription{"
for NAME in ${FUNCTIONS}; do
	JSON=$(jq ".[\"${NAME}\"]" "${JSON_FILE}")
	echo "\"${NAME}\": {"
	awk '{l[NR] = $0} END {for (i=2; i<=NR-1; i++) print l[i]}' <<< "${JSON}" | sed -r 's/^(\s*)"([^"]+)":(.*)/\1\2:\3/g;s/^\s*\<./\U&/g;s/\]/}/g;s/("|})\s*$/\1,/g;s/Params: \[/Params: []types.FunctionParam{/;s#Type: "([^"]+)"#Type: types.\u\1#g;s#(Options|Suggestions): \[#\1: []string{#g;s#      Default: ([^"]+),$#      Default: "\1",#g;s#(\s*)([.0-9]+),\s*$#\1"\2",#g'
	echo
	echo "},"
done
echo "}"
echo "}"
} # | gofmt
