#!/usr/bin/env bash
echo -e "package functions\n\nimport ("
for f in $(ls | egrep -v '(^example|\.go$)'); do
	echo "	_ \"github.com/go-graphite/carbonapi/expr/functions/${f}\""
done
echo ")" 
