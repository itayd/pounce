#!/bin/bash

set -euo pipefail

cd "$(dirname $0)"

export TESTROOT="$(pwd)"
export POUNCE="${TESTROOT}/../pounce"

TESTS="${@-*.clitest}"

for t in ${TESTS}; do
  echo "== test ${t}:"
  "./clitest" "${t}"
done
