#!/usr/bin/env bash

TESTS=$(ls cmd/mockbackend/testcases/)

OPTIONS=""
if [[ "$1" == "-break" ]]; then
	OPTIONS="${OPTIONS} -break"
	shift
fi

if [[ ! -z "${1}" ]]; then
	TESTS="${1}"
fi

FAILED_TESTS=""
for t in ${TESTS}; do
	if [[ "${TRAVIS}" == "true" ]]; then
		travis_fold start "test_${t}"
		travis_time_start
	fi

	echo "RUNNING TEST ${t}"
	./mockbackend ${OPTIONS} -test -config cmd/mockbackend/testcases/${t}/${t}.yaml
	status=$?

	if [[ "${TRAVIS}" == "true" ]]; then
		travis_time_finish
		travis_fold end "test_${t}"
	fi

	sleep 2
	if [[ ${status} -ne 0 ]]; then
		FAILED_TESTS="${FAILED_TESTS} ${t}"
		echo "test_${t}: FAIL"
	else
		echo "test_${t}: SUCCESS"
	fi
done

if [[ ! -z ${FAILED_TESTS} ]]; then
	echo "Some e2e tests failed: ${FAILED_TESTS}"
	exit 1
fi
