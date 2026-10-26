#!/bin/bash
# Artificial tests of transfer_run.sh / transfer_stage.sh / guard.sh with a SUBSTITUTE binary (no model, no solver, no weights).
HERE=$(cd "$(dirname "$0")" && pwd); T=$(mktemp -d); fail=0
FB=$T/fake; cat > $FB <<'EOS'
#!/bin/bash
# args: d1a1transfer -seed S -perround 40 -evalcap 100 -levels L -net N -dir SRC -out FILE
while [ $# -gt 0 ]; do case $1 in -net) N=$2;; -out) O=$2;; -seed) SD=$2;; -dir) SR=$2;; esac; shift; done
echo "$SD $N" >> "$(dirname "$O")/../calls.txt"
if [ "${FAKE_MODE:-ok}" = sleep ] && [ "$N" = "${FAKE_AT:-A}" ] && [ "$SD" = "${FAKE_SEED:-19}" ]; then sleep 30; fi
if [ "${FAKE_MODE:-ok}" = jsonsleep ] && [ "$N" = "${FAKE_AT:-A}" ]; then echo '{}' > "$O"; sleep 30; fi
if [ "${FAKE_MODE:-ok}" = tmpleft ] && [ "$N" = "${FAKE_AT:-A}" ]; then echo '{}' > "$O"; echo x > "$O.tmp"; exit 0; fi
if [ "${FAKE_MODE:-ok}" = rss ] && [ "$N" = "${FAKE_AT:-pre}" ]; then python3 -c "
import time
a=[]
for i in range(60):
    a.append(bytearray(10*1024*1024)); [x for x in a[-1][::4096]]; time.sleep(0.05)
time.sleep(5)"; fi
echo '{"net":"'$N'"}' > "$O.tmp" && mv "$O.tmp" "$O"
EOS
chmod +x $FB; SHA=$(sha256sum $FB | cut -d' ' -f1)
SRC=$T/src; mkdir -p $SRC/seed-19 $SRC/seed-20; echo w > $SRC/seed-19/weights_A.json; echo w > $SRC/seed-20/weights_A.json; SRCH=$(cd $SRC && find . -type f | sort | xargs sha256sum | sha256sum)
check() { if [ "$2" = "$3" ]; then echo "PASS $1"; else echo "FAIL $1 (got '$2' want '$3')"; fail=1; fi; }
run() { env "$@" BIN=$FB BIN_SHA=$SHA SRC_ROOT=$SRC bash "$HERE/transfer_run.sh"; }
R=$T/r1; out=$(run ROOT=$R SEEDS="19 20" STAGE_CAP_S=10 GLOBAL_BUDGET_S=200 RSS_MIB=900 2>&1); rc=$?
check "happy: exit 0" "$rc" 0; check "happy: 6 calls" "$(wc -l < $R/calls.txt)" 6; check "happy: all COMPLETED" "$(cat $R/seed-*/status_* | grep -c outcome=COMPLETED)" 6
check "happy: json files" "$(ls $R/seed-*/transfer-*.json | wc -l)" 6; check "happy: no tmp left" "$(ls $R/seed-*/*.tmp* 2>/dev/null | wc -l)" 0
check "happy: source untouched" "$(cd $SRC && find . -type f | sort | xargs sha256sum | sha256sum)" "$SRCH"
check "happy: clocks logged" "$(grep -c '^end operative' $R/CLOCKS.txt)" 1
R=$T/r2; out=$(run ROOT=$R SEEDS="19 20" STAGE_CAP_S=5 GLOBAL_BUDGET_S=200 FAKE_MODE=sleep FAKE_AT=A FAKE_SEED=19 2>&1)
check "timeout: A censored" "$(grep -c outcome=CENSORED_TIMEOUT $R/seed-19/status_A)" 1; check "timeout: F of seed 19 not run" "$(grep -c '^19 F' $R/calls.txt)" 0
check "timeout: seed 20 complete" "$(cat $R/seed-20/status_* | grep -c outcome=COMPLETED)" 3; check "timeout: A called once (no retry)" "$(grep -c '^19 A' $R/calls.txt)" 1
R=$T/r3; out=$(run ROOT=$R SEEDS=19 STAGE_CAP_S=10 RSS_MIB=40 FAKE_MODE=rss 2>&1)
check "rss: not COMPLETED" "$(grep -c outcome=COMPLETED $R/seed-19/status_pre 2>/dev/null)" 0; check "rss: memory outcome" "$(grep -c CENSORED_MEMORY $R/seed-19/status_pre)" 1
R=$T/r4; out=$(run ROOT=$R SEEDS="19 20" STAGE_CAP_S=10 GLOBAL_BUDGET_S=9 2>&1); rc=$?; check "budget: exit 5" "$rc" 5; check "budget: message" "$(echo "$out" | grep -c GLOBAL_BUDGET_EXHAUSTED)" 1
R=$T/r5; out=$(run ROOT=$R SEEDS=19 STAGE_CAP_S=10 CIVIL_DEADLINE_S=5 2>&1); rc=$?; check "civil before stage: exit 6" "$rc" 6; check "civil: nothing ran" "$(ls $R/calls.txt 2>/dev/null | wc -l)" 0
R=$T/r6/seed-19; mkdir -p $T/r6; n=$(date +%s)
env DIR=$R SRC=$SRC/seed-19 BIN=$FB BIN_SHA=$SHA SEED=19 STAGE_CAP_S=10 TOTAL_BUDGET_S=50 GUARD_CIVIL_END=$((n+2)) FAKE_MODE=sleep FAKE_AT=pre FAKE_SEED=19 bash "$HERE/transfer_stage.sh" pre >/dev/null 2>&1; rc=$?
check "civil INSIDE stage: exit 6" "$rc" 6; check "civil INSIDE stage: outcome" "$(grep -c outcome=CENSORED_CIVIL $R/status_pre)" 1
R=$T/r7/seed-19; mkdir -p $T/r7
env DIR=$R SRC=$SRC/seed-19 BIN=$FB BIN_SHA=$SHA SEED=19 STAGE_CAP_S=4 TOTAL_BUDGET_S=50 FAKE_MODE=jsonsleep FAKE_AT=A bash "$HERE/transfer_stage.sh" A >/dev/null 2>&1; rc=$?
check "json then timeout: not completed" "$(grep -c outcome=CENSORED_TIMEOUT $R/status_A)" 1
R=$T/r8/seed-19; mkdir -p $T/r8
env DIR=$R SRC=$SRC/seed-19 BIN=$FB BIN_SHA=$SHA SEED=19 STAGE_CAP_S=10 TOTAL_BUDGET_S=50 FAKE_MODE=tmpleft FAKE_AT=A bash "$HERE/transfer_stage.sh" A >/dev/null 2>&1
check "leftover .tmp: not completed" "$(grep -c outcome=COMPLETED $R/status_A)" 0
R=$T/r9; mkdir -p $R/.driver.lock; out=$(run ROOT=$R SEEDS=19 STAGE_CAP_S=10 2>&1); rc=$?; check "driver lock held: exit 7" "$rc" 7; check "lock: nothing ran" "$(ls $R/calls.txt 2>/dev/null | wc -l)" 0
R=$T/r1; out=$(run ROOT=$R SEEDS="19 20" STAGE_CAP_S=10 GLOBAL_BUDGET_S=200 RSS_MIB=900 2>&1); rc=$?; check "restart rejected: exit 7" "$rc" 7; check "restart rejected: message" "$(echo "$out" | grep -c RESTART_REJECTED)" 1
R=$T/r10; out=$(env BIN=$FB BIN_SHA=deadbeef SRC_ROOT=$SRC ROOT=$R SEEDS="19 20" STAGE_CAP_S=10 bash "$HERE/transfer_run.sh" 2>&1); rc=$?; check "binary mismatch: global stop exit 7" "$rc" 7; check "binary mismatch: seed 20 never started" "$(echo "$out" | grep -c 'seed 20')" 0
mkdir -p $T/noawk; printf '#!/bin/sh\nexit 1\n' > $T/noawk/awk; chmod +x $T/noawk/awk
R=$T/r11; out=$(PATH=$T/noawk:$PATH run ROOT=$R SEEDS=19 STAGE_CAP_S=10 2>&1); rc=$?; check "empty operative clock: exit 7" "$rc" 7
mkdir -p $T/fakedate; printf '#!/bin/sh\necho oops\n' > $T/fakedate/date; chmod +x $T/fakedate/date
R=$T/r12; out=$(PATH=$T/fakedate:$PATH run ROOT=$R SEEDS=19 STAGE_CAP_S=10 2>&1); rc=$?; check "civil clock unreadable: exit 7" "$rc" 7
R=$T/r1/seed-19; out=$(env DIR=$R SRC=$SRC/seed-19 BIN=$FB BIN_SHA=$SHA SEED=19 STAGE_CAP_S=10 TOTAL_BUDGET_S=50 bash "$HERE/transfer_stage.sh" pre 2>&1); check "no overwrite: refused" "$(echo "$out" | grep -c 'already has status')" 1
R=$T/r13; out=$(run ROOT=$R SEEDS=19 STAGE_CAP_S=10 STAGE_CAP_S=11 2>&1); R2=$T/r13; out=$(env BIN=$FB BIN_SHA=$SHA SRC_ROOT=$SRC ROOT=$R2 SEEDS=19 STAGE_CAP_S=12 bash "$HERE/transfer_run.sh" 2>&1); check "second start same root rejected" "$(echo "$out" | grep -c 'RESTART_REJECTED\|CONFIG_MISMATCH')" 1
check "defaults 360/7200/14400" "$(grep -c 'STAGE_CAP_S:-360' $HERE/transfer_run.sh)$(grep -c 'GLOBAL_BUDGET_S:-7200' $HERE/transfer_run.sh)$(grep -c 'CIVIL_DEADLINE_S:-14400' $HERE/transfer_run.sh)" 111
rm -rf $T; exit $fail
