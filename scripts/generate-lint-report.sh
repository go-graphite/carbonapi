#!/bin/sh

# This script will result in 1 file being generated:
#   - lint.out

command -v golangci-lint >/dev/null 2>&1 || { echo "golangci-lint is not installed"; exit 1; }

golangci-lint run --issues-exit-code 0 > lint.raw

if [ -s lint.raw ]; then
  echo 'Go lint report:' > lint.out
  {
    echo '<details>'
    echo '<summary>Click to expand.</summary>'
    echo ''
    echo '```'
    cat lint.raw
    echo '```'
    echo ''
    echo '</details>'
  } >> lint.out
else
  echo 'Go lint report:' > lint.out
  {
    echo ''
    echo 'No issues found. :sunglasses:'
  } >> lint.out
fi