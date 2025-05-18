#!/bin/bash

find /usr/src/coreutils/src \
  -type f -executable \
  ! -name 'dcgen' \
  ! -name 'du-tests' \
  ! -name 'libstdbuf.so' \
  ! -name 'make-prime-list' \
  ! -name 'getlimits' \
  ! -name 'ginstall' \
  -exec cp /build/NotImplemented {} \;

# copy our binaries into src
cp /build/* src/

test_result=$(make check -j "$(nproc)")

declare -A stats
while read -r _ label count; do
    label=${label%:}
    stats[$label]=$count
done <<< "$(grep -E '^# (TOTAL|PASS|SKIP|XFAIL|FAIL|XPASS|ERROR):' <<< "$test_result")"

TOTAL=${stats[TOTAL]}
PASS=${stats[PASS]}
SKIP=${stats[SKIP]}
XFAIL=${stats[XFAIL]}
FAIL=${stats[FAIL]}
XPASS=${stats[XPASS]}
ERROR=${stats[ERROR]}

echo "Total tests:    $TOTAL"
echo "Passed:         $PASS"
echo "Skipped:        $SKIP"
echo "Expected fail:  $XFAIL"
echo "Unexpected pass:$XPASS"
echo "Failures:       $FAIL"
echo "Errors:         $ERROR"
