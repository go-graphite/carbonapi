#!/usr/bin/env bash

_RUNNING_OS=$(uname -s)

function get_listeners() {
  case "${_RUNNING_OS}" in
    "Linux")
      ss -ltpn | grep "${1}" | awk '{print $4}' | sort -u
      ;;
    "Darwin")
      lsof -nP -iTCP -sTCP:LISTEN | grep "${1}" | awk '{print $9}' | sort -u
      ;;
    *)
      echo "OS ${_RUNNING_OS} currently not supported to run this test" >2
      exit 130
      ;;
  esac
}

cleanup() {
        local pids=$(jobs -pr)
        if [[ -n "${pids}" ]]; then
          kill -9 ${pids} 2>/dev/null ||:
        fi
}