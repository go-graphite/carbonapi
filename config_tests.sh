#!/usr/bin/env bash

STATUS=0
for i in cmd/carbonapi/config_tests/*-*.sh; do
	${i}
	NEW_STATUS=$?
	STATUS=$((STATUS + NEW_STATUS))
done

exit ${STATUS}
