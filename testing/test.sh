#!/bin/bash
echo "fixing permissions"
sudo find /usr/src/coreutils/ \
  \( ! -user user -o ! -group user \) \
  -exec chown user:user {} +
echo "fixing permissions done"

echo "copying not implemented binaries"
find /usr/src/coreutils/build/src \
  -type f -executable \
  ! -name 'dcgen' \
  ! -name 'du-tests' \
  ! -name 'libstdbuf.so' \
  ! -name 'make-prime-list' \
  ! -name 'getlimits' \
  ! -name 'ginstall' \
  -exec cp /build/NotImplemented {} \;
echo "copying not implemented binaries done"

echo "copying our binaries into src"
cp /build/* src/
echo "copying our binaries into src done"

echo "copying build files into tmpfs"
cp -r /usr/src/coreutils/build/* /usr/src/coreutils/tmpfs/
cd /usr/src/coreutils/tmpfs
echo "copying build files into tmpfs done"

echo "running tests"
if [ -n "$TESTS" ]; then
  test_result=$(make check -j "$(nproc)" TESTS=$TESTS)
else
  test_result=$(make check -j "$(nproc)")
fi
echo "running tests done"

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

cp ./tests/test-suite.log /build/test-suite.log
echo "test-suite.log copied to /build/test-suite.log"