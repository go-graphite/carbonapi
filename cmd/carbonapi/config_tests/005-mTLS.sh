#!/usr/bin/env bash

CURL_VERSION=$(curl --version | head -n 1 | awk '{print $2}')

CURL_MAJOR_V=$(cut -d. -f 1 <<< "${CURL_VERSION}")
CURL_MINOR_V=$(cut -d. -f 2 <<< "${CURL_VERSION}")

if [[ ${CURL_MAJOR_V} -le 7 ]]; then
	if [[ ${CURL_MAJOR_V} -lt 7 ]] || [[ ${CURL_MINOR_V} -lt 54 ]]; then
		echo "curl >= 7.54 is required"
		exit 2
	fi
fi

CURL_SSL=$(curl --version | grep -E -o "[^ ]+SSL" | head -n 1)
if [[ "${CURL_SSL}" == "LibreSSL" ]]; then
  echo "CURL with LibreSSL is known to fail with ed25519 curves required for tls 1.3, so mTLS test WILL fail"
  sleep 1
fi

set -e

source "$(dirname "${0}")/common.sh"

TEST_DIR=$(dirname "${0}")
TEST_NAME=$(basename "${0}")
STATUS=0
echo ${TEST_NAME/.sh/.yaml}

EXPECTED_LISTENERS=(
	"127.0.0.1:8082"
)

trap "cleanup" SIGINT SIGTERM EXIT INT QUIT TERM EXIT
echo "carbonapi -config \"${TEST_DIR}/${TEST_NAME/.sh/.yaml}\" &"
./carbonapi -config "${TEST_DIR}/${TEST_NAME/.sh/.yaml}" &
sleep 2

LISTENERS=$(get_listeners "carbonapi")

set +e

cnt=0
for l in ${LISTENERS}; do
	cnt=$((cnt+1))
	found=0
	for el in ${EXPECTED_LISTENERS[@]}; do
		if [[ "${el}" == "${l}" ]]; then
			found=1
			break
		fi
	done
	if [[ ${found} -eq 0 ]]; then
		echo "Listener ${l} is not expected"
		STATUS=1
	fi
done

if [[ ${cnt} -ne ${#EXPECTED_LISTENERS[@]} ]]; then
	echo "Expected listener count mismatch, got ${cnt}, expected ${#EXPECTED_LISTENERS[@]}"
	STATUS=1
fi

if [[ ${STATUS} -ne 0 ]]; then
	echo "${TEST_NAME} FAIL"
	kill %1
	wait
	exit ${STATUS}
fi

# CURL should fail as we haven't provided client certificate
OUT=$(curl -kvvI https://127.0.0.1:8082 2>&1)
CURL_STATUS=${?}
if [[ ${CURL_STATUS} -eq 0 ]]; then
	echo "${OUT}"
	echo "${TEST_NAME} FAIL"
	STATUS=1
	kill %1
	wait
	exit ${STATUS}
fi

EXPECTED_CURL_OUTPUT=(
	"Failed sending HTTP2 data"
)

OLD_IFS="${IFS}"
IFS=$'\n'
for t in ${EXPECTED_CURL_OUTPUT[@]}; do
    IFS="${OLD_IFS}"
	echo "Testing for ${t}"
	grep -q "${t}" <<< "${OUT}"
	if [[ ${?} -ne 0 ]]; then
		echo
		echo "Test for '${t}' in output failed"
		echo "${OUT}"
		echo "${TEST_NAME} FAIL"
		STATUS=1
	fi
done
IFS="${OLD_IFS}"

# CURL should succeed as we've provided client certificate
OUT=$(curl --cacert ./cmd/carbonapi/config_tests/mTLS-server.crt --key ./cmd/carbonapi/config_tests/mTLS-client.key --cert ./cmd/carbonapi/config_tests/mTLS-client.crt -kvvI https://127.0.0.1:8082 2>&1)
CURL_STATUS=${?}
if [[ ${CURL_STATUS} -ne 0 ]]; then
	echo "${OUT}"
	echo "${TEST_NAME} FAIL"
	STATUS=1
	kill %1
	wait
	exit ${STATUS}
fi

EXPECTED_CURL_OUTPUT=(
	"subject: CN=carbonapi-test1"
	"HTTP/2 200"
)

OLD_IFS="${IFS}"
IFS=$'\n'
for t in ${EXPECTED_CURL_OUTPUT[@]}; do
	IFS="${OLD_IFS}"
	echo "Testing for ${t}"
	grep -q "${t}" <<< "${OUT}"
	if [[ ${?} -ne 0 ]]; then
		echo
		echo "Test for '${t}' in output failed"
		echo "${OUT}"
		echo "${TEST_NAME} FAIL"
		STATUS=1
	fi
done
IFS="${OLD_IFS}"

kill %1
wait

if [[ ${STATUS} -eq 0 ]]; then
	echo "${TEST_NAME} OK"
else
	echo "${TEST_NAME} FAIL"
fi

exit ${STATUS}
