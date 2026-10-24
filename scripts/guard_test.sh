#!/bin/bash
# Deterministic, light tests of guard.sh cleanup paths. Each test prints PASS/FAIL. Uses unique sleep durations.
G=${1:-./guard.sh}; fail=0
alive() { pgrep -a -x sleep | grep -c " $1\$"; }
check() { if [ "$2" = "$3" ]; then echo "PASS $1"; else echo "FAIL $1 (got $2 want $3)"; fail=1; fi; }
# 1 leader exits normally, descendant left behind
out=$($G 20 900 /dev/null bash -c 'sleep 71 & exit 0'); sleep 0.5
check "leader exit + descendant: descendant killed" "$(alive 71)" 0; echo "$out" | grep -q exit_with_descendants_killed; check "leader exit: reported" $? 0
# 2 leader fails (exit 3), descendant left behind
out=$($G 20 900 /dev/null bash -c 'sleep 72 & exit 3'); rc=$?; sleep 0.5
check "leader fail + descendant: descendant killed" "$(alive 72)" 0
# 3 signals to the supervisor with descendants
for sig in TERM INT HUP; do
  n=$((73+$(echo $sig | cksum | cut -c1-1)))
  $G 30 900 /dev/null bash -c "sleep $n & sleep $n" > /tmp/gsig.out & g=$!
  sleep 1; check "$sig: descendants running before signal" "$(alive $n)" 2; kill -$sig $g; wait $g 2>/dev/null; sleep 0.5
  check "$sig to supervisor: group killed" "$(alive $n)" 0
done
# 4 timeout with descendants
out=$($G 2 900 /dev/null bash -c 'sleep 81 & sleep 82'); sleep 0.5
check "timeout: descendants killed" "$(alive 81)$(alive 82)" 00
# 5 RSS limit with descendants (child allocates ~150 MiB/s)
out=$($G 30 150 /dev/null bash -c 'python3 -c "
import time
a=[]
while True:
    a.append(bytearray(15*1024*1024)); [x for x in a[-1][::4096]]; time.sleep(0.1)
" & wait'); sleep 0.5
echo "$out" | grep -q 'status=RSS_LIMIT'; check "RSS limit: status" $? 0
check "RSS limit: allocator killed" "$(pgrep -f 'bytearray\(15' | wc -l)" 0
# 6 sampler cannot read /proc -> fail closed
out=$(GUARD_PROC=/nonexistent $G 20 900 /dev/null bash -c 'sleep 83 & sleep 84'); sleep 0.5
echo "$out" | grep -q 'status=SAMPLER_FAILED'; check "sampler failure: status" $? 0
check "sampler failure: group killed" "$(alive 83)$(alive 84)" 00
# 7 supervisor SIGKILLed: time layer ends the group within lim+5 s (memory NOT enforced)
$G 2 900 /dev/null sleep 85 & g=$!; sleep 1; kill -KILL $g; wait $g 2>/dev/null
sleep 8; check "supervisor SIGKILL: time layer ended child" "$(alive 85)" 0
exit $fail
