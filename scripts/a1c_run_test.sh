#!/bin/bash
# Tests of a1c_run.sh and pilot_a1_stage.sh with a SUBSTITUTE binary (no training, no evaluation, no model).
# The fake binary writes $DIR/$STAGE.json and behaves as told by FAKE_MODE: ok | sleep (exceeds the cap) | rss (allocates).
HERE=$(cd "$(dirname "$0")" && pwd); T=$(mktemp -d); fail=0
FB=$T/fake; cat > $FB <<'EOS'
#!/bin/bash
# args: d1a1train -seed S -perround 40 -cap 100 -evalcap 100 -bank g3 -stage ST -dir D
while [ $# -gt 0 ]; do case $1 in -stage) ST=$2;; -dir) D=$2;; -bank) BK=$2;; -seed) SD=$2;; esac; shift; done
echo "$BK $SD $ST" >> "$D/../calls.txt"
sleep "${FAKE_DELAY:-0}"
case "${FAKE_MODE:-ok}" in
 jsonsleep) if [ "$ST" = "${FAKE_AT:-train-A}" ] && [ "$SD" = "${FAKE_SEED:-19}" ]; then echo '{"stage":"'$ST'"}' > "$D/$ST.json"; sleep 30; fi;;
 sleep) [ "$ST" = "${FAKE_AT:-train-A}" ] && [ "$SD" = "${FAKE_SEED:-19}" ] && sleep 30;;
 rss) [ "$ST" = "${FAKE_AT:-pre-g3}" ] && python3 -c "
import time
a=[]
for i in range(60):
    a.append(bytearray(10*1024*1024)); [x for x in a[-1][::4096]]; time.sleep(0.05)
time.sleep(5)";;
esac
echo '{"stage":"'$ST'"}' > "$D/$ST.json"
EOS
chmod +x $FB; SHA=$(sha256sum $FB | cut -d' ' -f1)
check() { if [ "$2" = "$3" ]; then echo "PASS $1"; else echo "FAIL $1 (got '$2' want '$3')"; fail=1; fi; }
run() { env "$@" BIN=$FB BIN_SHA=$SHA bash "$HERE/a1c_run.sh"; }
# 1 happy path with 2 seeds: 5 stages each, bank g3, total budget accounted
R=$T/r1; out=$(run ROOT=$R SEEDS="19 20" STAGE_CAP_S=10 GLOBAL_BUDGET_S=200 RSS_MIB=900 2>&1)
check "happy: calls" "$(wc -l < $R/calls.txt)" 10; check "happy: bank g3 only" "$(grep -vc '^g3 ' $R/calls.txt)" 0
check "happy: statuses" "$(cat $R/seed-19/status_* | grep -c outcome=COMPLETED)" 5
# 2 stage exceeding its cap: censored, later stages of that seed refuse, next seed still runs, no retry
R=$T/r2; out=$(run ROOT=$R SEEDS="19 20" STAGE_CAP_S=5 GLOBAL_BUDGET_S=200 FAKE_MODE=sleep FAKE_AT=train-A FAKE_SEED=19 2>&1)
check "censor: train-A status" "$(grep -c CENSORED_TIMEOUT $R/seed-19/status_train-A)" 1
check "censor: seed 19 stops there" "$(ls $R/seed-19 | grep -c '^status_')" 2
check "censor: seed 20 complete" "$(cat $R/seed-20/status_* | grep -c outcome=COMPLETED)" 5
check "censor: train-A called once (no retry)" "$(grep -c '^g3 19 train-A' $R/calls.txt)" 1
# 3 RSS limit via the guard on the process tree (limit 200 MiB; fake allocates ~600 MiB)
R=$T/r3; out=$(run ROOT=$R SEEDS="19" STAGE_CAP_S=20 GLOBAL_BUDGET_S=200 RSS_MIB=200 FAKE_MODE=rss FAKE_AT=pre-g3 2>&1)
check "rss: killed, not COMPLETED" "$(grep -c COMPLETED $R/seed-19/status_pre-g3)" 0
check "rss: status mentions RSS" "$(grep -c 'RSS_LIMIT' $R/seed-19/status_pre-g3)" 1
# 4 global budget: cap 10, global 12, each fake stage takes 2 s => stops after at most 3 stages with the exhaustion message
R=$T/r4; out=$(run ROOT=$R SEEDS="19 20 21" STAGE_CAP_S=10 GLOBAL_BUDGET_S=12 FAKE_DELAY=2 2>&1)
echo "$out" | grep -q GLOBAL_BUDGET_EXHAUSTED; check "budget: exhausted message" $? 0
check "budget: stages started <= 3" "$([ $(wc -l < $R/calls.txt) -le 3 ] && echo ok)" ok
# 5 wrong binary hash refuses
R=$T/r5; BIN=$FB BIN_SHA=deadbeef ROOT=$R SEEDS=19 STAGE_CAP_S=10 GLOBAL_BUDGET_S=100 bash "$HERE/a1c_run.sh" >/dev/null 2>&1
check "wrong sha: nothing ran" "$(ls $R/calls.txt 2>/dev/null | wc -l)" 0
# 6 civil deadline smaller than one stage cap: nothing runs, exit 6, separate message
R=$T/r6; out=$(run ROOT=$R SEEDS=19 STAGE_CAP_S=10 GLOBAL_BUDGET_S=200 CIVIL_DEADLINE_S=5 2>&1); rc=$?
check "civil: exit 6" "$rc" 6; check "civil: message" "$(echo "$out" | grep -c CIVIL_DEADLINE_EXHAUSTED)" 1; check "civil: nothing ran" "$(ls $R/calls.txt 2>/dev/null | wc -l)" 0
# 7 civil clock unreadable: fail closed
mkdir -p $T/fakedate; printf '#!/bin/sh\necho oops\n' > $T/fakedate/date; chmod +x $T/fakedate/date
R=$T/r7; out=$(PATH=$T/fakedate:$PATH run ROOT=$R SEEDS=19 STAGE_CAP_S=10 2>&1); rc=$?
check "civil clock fail: exit 7" "$rc" 7
# 8 wrapper: civil deadline DURING the stage => CENSORED_CIVIL, reserve kept as max(elapsed,reserved), no retry
R=$T/r8/seed-19; mkdir -p $T/r8; n=$(date +%s)
env DIR=$R BIN=$FB BIN_SHA=$SHA SEED=19 BANK=g3 STAGE_CAP_S=10 TOTAL_BUDGET_S=50 GUARD_CIVIL_END=$((n+2)) FAKE_MODE=sleep FAKE_AT=pre-g3 FAKE_SEED=19 bash "$HERE/pilot_a1_stage.sh" pre-g3 >/dev/null 2>&1; rc=$?
check "civil in stage: wrapper exit 6" "$rc" 6; check "civil in stage: outcome" "$(grep -c outcome=CENSORED_CIVIL $R/status_pre-g3)" 1
check "civil in stage: reserve kept (>=10)" "$([ $(cat $R/budget_used_s) -ge 10 ] && echo y)" y
# 9 config mismatch on restart: fail closed (exit 7), nothing new ran
R=$T/r1; out=$(run ROOT=$R SEEDS="19 20" STAGE_CAP_S=11 GLOBAL_BUDGET_S=200 RSS_MIB=900 2>&1); rc=$?
check "config mismatch: exit 7" "$rc" 7; check "config mismatch/restart: rejected" "$(echo "$out" | grep -c "RESTART_REJECTED\|CONFIG_MISMATCH")" 1
# 10 driver records clocks and frozen config; defaults are 300/7200/14400
check "driver clocks file: one start only" "$(grep -c '^start operative' $T/r1/CLOCKS.txt)" 1
check "defaults consistent" "$(grep -c 'STAGE_CAP_S:-300.*' $HERE/a1c_run.sh)$(grep -c 'GLOBAL_BUDGET_S:-7200' $HERE/a1c_run.sh)$(grep -c 'CIVIL_DEADLINE_S:-14400' $HERE/a1c_run.sh)" 111
# 11 existing stage file not overwritten (stage already has status)
R=$T/r1/seed-19; out=$(env DIR=$R BIN=$FB BIN_SHA=$SHA SEED=19 BANK=g3 STAGE_CAP_S=10 TOTAL_BUDGET_S=50 bash "$HERE/pilot_a1_stage.sh" pre-g3 2>&1); check "no overwrite: refused" "$(echo "$out" | grep -c 'already has status')" 1
# 12 TIMEOUT after JSON was written is NOT COMPLETED (wrapper requires guard status=exit rc=0)
R=$T/r12/seed-19; mkdir -p $T/r12
env DIR=$R BIN=$FB BIN_SHA=$SHA SEED=19 BANK=g3 STAGE_CAP_S=4 TOTAL_BUDGET_S=50 FAKE_MODE=jsonsleep FAKE_AT=pre-g3 FAKE_SEED=19 bash "$HERE/pilot_a1_stage.sh" pre-g3 >/dev/null 2>&1; rc=$?
check "json then timeout: not completed" "$(grep -c outcome=CENSORED_TIMEOUT $R/status_pre-g3)" 1; check "json then timeout: exit 6" "$rc" 6
# 13 driver lock held / restart rejected / wrapper hard failure stops ALL seeds
R=$T/r13; mkdir -p $R/.driver.lock; out=$(run ROOT=$R SEEDS=19 STAGE_CAP_S=10 2>&1); rc=$?; check "driver lock held: exit 7" "$rc" 7; check "driver lock: nothing ran" "$(ls $R/calls.txt 2>/dev/null | wc -l)" 0
R=$T/r1; out=$(run ROOT=$R SEEDS="19 20" STAGE_CAP_S=10 GLOBAL_BUDGET_S=200 RSS_MIB=900 2>&1); rc=$?; check "restart rejected: exit 7" "$rc" 7; check "restart rejected: message" "$(echo "$out" | grep -c RESTART_REJECTED)" 1
R=$T/r14; out=$(env BIN=$FB BIN_SHA=deadbeef ROOT=$R SEEDS="19 20" STAGE_CAP_S=10 bash "$HERE/a1c_run.sh" 2>&1); rc=$?; check "binary mismatch: global stop exit 7" "$rc" 7; check "binary mismatch: seed 20 never started" "$(echo "$out" | grep -c 'seed 20')" 0
# 14 empty operative clock in driver: fails closed
mkdir -p $T/noawk; printf '#!/bin/sh\nexit 1\n' > $T/noawk/awk; chmod +x $T/noawk/awk
R=$T/r15; out=$(PATH=$T/noawk:$PATH run ROOT=$R SEEDS=19 STAGE_CAP_S=10 2>&1); rc=$?; check "empty operative clock: exit 7" "$rc" 7
rm -rf $T; exit $fail
