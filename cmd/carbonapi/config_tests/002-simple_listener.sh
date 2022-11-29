#!/usr/bin/env bash

set -e

TEST_DIR=$(dirname ${0})
TEST_NAME=$(basename ${0})
STATUS=0
echo ${TEST_NAME/.sh/.yaml}

EXPECTED_LISTENERS=(
	"127.0.0.1:8082"
)


echo "carbonapi -config ${TEST_DIR}/${TEST_NAME/.sh/.yaml} &"
./carbonapi -config ${TEST_DIR}/${TEST_NAME/.sh/.yaml} &

sleep 2

LISTENERS=$(ss -ltpn | grep carbonapi | awk '{print $4}' | sort -u)

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

kill %1
wait

if [[ ${STATUS} -eq 0 ]]; then
	echo "${TEST_NAME} OK"
else
	echo "${TEST_NAME} FAIL"
fi

exit ${STATUS}
