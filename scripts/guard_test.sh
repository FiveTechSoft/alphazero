#!/bin/bash
# Deterministic, light tests of guard.sh cleanup paths. Each test prints PASS/FAIL. Uses unique sleep durations.
HERE=$(cd "$(dirname "$0")" && pwd); G=${1:-"bash $HERE/guard.sh"}; fail=0
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
# civil deadline enforced DURING a stage, and unreadable civil clock fails closed
n=$(date +%s); o=$(GUARD_CIVIL_END=$((n+2)) $G 60 1024 /tmp/gt_cv.out sleep 30 2>&1); check "civil during stage: CIVIL_DEADLINE" "$(echo "$o" | grep -c 'status=CIVIL_DEADLINE')" 1
FD=$(mktemp -d); printf '#!/bin/sh\necho oops\n' > $FD/date; chmod +x $FD/date
o=$(PATH=$FD:$PATH GUARD_CIVIL_END=$((n+999)) $G 60 1024 /tmp/gt_cv.out sleep 30 2>&1); check "civil clock unreadable: fails closed" "$(echo "$o" | grep -c 'status=CIVIL_CLOCK_FAILED')" 1
# watcher/orphans: supervisor SIGKILLed AND leader SIGKILLed, children alive: cleaned by the watcher within lim+5
$G 3 900 /dev/null bash -c 'sleep 91 & sleep 92 & wait' >/dev/null 2>&1 & sp=$!; sleep 1
kill -9 $sp 2>/dev/null; lp=$(pgrep -f 'sleep 91' | head -1); lg=$(ps -o ppid= -p $lp | tr -d ' '); kill -9 $lg 2>/dev/null
check "orphans alive right after supervisor+leader killed" "$(( $(alive 91) + $(alive 92) ))" 2
sleep 9; check "orphans cleaned by watcher" "$(( $(alive 91) + $(alive 92) ))" 0
# civil clock going backwards => fail closed
FB2=$(mktemp -d); cat > $FB2/date <<EOS
#!/bin/sh
v=\$(cat $FB2/c 2>/dev/null || echo 2000000000); echo \$((v-10)) > $FB2/c; echo \$v
EOS
chmod +x $FB2/date
o=$(PATH=$FB2:$PATH GUARD_CIVIL_END=2100000000 $G 60 1024 /tmp/gt_cv.out sleep 30 2>&1); check "civil clock backward: fails closed" "$(echo "$o" | grep -c 'status=CIVIL_CLOCK_BACKWARD')" 1
check "end line has civil_start/end" "$(echo "$o" | grep -c 'civil_start=.*civil_end=')" 1
# supervisor killed almost immediately (before the group is surely created): watcher must still end the group at lim+5
$G 3 900 /dev/null sleep 93 >/dev/null 2>&1 & sp=$!; sleep 0.05; kill -9 $sp 2>/dev/null; sleep 10
check "early supervisor kill: watcher still cleaned group" "$(alive 93)" 0
# empty operative clock: guard fails closed (CLOCK_FAILED)
ED=$(mktemp -d); printf '#!/bin/sh\nexit 1\n' > $ED/awk; chmod +x $ED/awk
o=$(PATH=$ED:$PATH $G 20 900 /dev/null sleep 94 2>&1); check "empty operative clock: CLOCK_FAILED" "$(echo "$o" | grep -c CLOCK_FAILED)" 1; sleep 1; check "empty operative clock: nothing left" "$(alive 94)" 0
exit $fail
