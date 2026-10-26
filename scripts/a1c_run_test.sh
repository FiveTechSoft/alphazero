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
rm -rf $T; exit $fail
