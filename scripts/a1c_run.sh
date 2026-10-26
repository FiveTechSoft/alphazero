#!/bin/bash
# a1c_run.sh: A1C execution plan, sequential, one CPU, no retries.
# Required env: BIN (binary), BIN_SHA (its full sha256), ROOT (output root).
# Exit codes: 0 finished loop (maybe INCOMPLETE: read statuses); 5 operative budget exhausted; 6 civil deadline exhausted
# (before a stage); 7 clock unreadable/backward, config mismatch or setup failure. A stage not COMPLETED stops that seed
# (wrapper exit 6 and its status_* file carry the cause: CENSORED_TIMEOUT / CENSORED_CIVIL / CENSORED_CLOCK / CENSORED_MEMORY / FAILED).
# Optional env (defaults are the plan): SEEDS="19 20 21 22 23 24 25 26", GLOBAL_BUDGET_S=7200 (operative seconds,
# hard, counted as the sum of budget_used_s over seed dirs plus wall clock), STAGE_CAP_S=300, CIVIL_DEADLINE_S=14400, RSS_MIB=1024.
# Per seed 5 stages (pre-g3 train-A train-F post-A post-F), seed by seed. Each stage runs through
# pilot_a1_stage.sh (BANK=g3) under guard.sh, which polls the RSS of the whole process group and kills it
# above RSS_MIB (a polling guard: it can be overshot between polls; GOMEMLIMIT is NOT the limit).
# A censored or failed stage is kept as is: later stages of that seed refuse to run (status not COMPLETED),
# the driver goes on with the next seed, nothing is retried or re-parameterized. When the global budget
# cannot cover another stage the driver stops: the missing pairs make the primary result INCOMPLETE and
# not confirmatory. n is never reduced.
set -u
mono() { awk '{print int($1)}' /proc/uptime 2>/dev/null; }  # operative clock: /proc/uptime seconds (observed behaviour, not a universal guarantee)
HERE=$(cd "$(dirname "$0")" && pwd)
BIN=${BIN:?BIN}; BIN_SHA=${BIN_SHA:?BIN_SHA}; ROOT=${ROOT:?ROOT}
SEEDS=${SEEDS:-"19 20 21 22 23 24 25 26"}
GLOBAL=${GLOBAL_BUDGET_S:-7200}; CAP=${STAGE_CAP_S:-300}; RSS=${RSS_MIB:-1024}
WRAP=${WRAP:-$HERE/pilot_a1_stage.sh}
export GUARD=${GUARD:-$HERE/guard.sh}
CIVIL=${CIVIL_DEADLINE_S:-14400}  # civil (wall) deadline in seconds, a SEPARATE budget from the operative one
c0=$(date +%s); [[ "$c0" =~ ^[0-9]+$ ]] || { echo CLOCK_FAILED_CIVIL; exit 7; }
export GUARD_CIVIL_END=$((c0+CIVIL))  # guard also enforces the civil deadline DURING each stage (a suspension jumps date, so it fires on resume)
t0=$(mono); [ -n "$t0" ] || { echo CLOCK_FAILED; exit 7; }
mkdir -p "$ROOT" || exit 7
# exclusive atomic driver lock (kept after a crash: fail closed) and refusal to restart a root that already started (no civil/operative reset, no silent resume)
mkdir "$ROOT/.driver.lock" 2>/dev/null || { echo "DRIVER_LOCK_HELD: $ROOT/.driver.lock"; exit 7; }
trap 'rmdir "$ROOT/.driver.lock" 2>/dev/null' EXIT
[ -e "$ROOT/CLOCKS.txt" ] && { echo "RESTART_REJECTED: $ROOT already started (CLOCKS.txt present); new root required, no resume"; exit 7; }
# frozen effective config: written once; a later start with different effective values fails closed (exit 7)
CFG="SEEDS=$SEEDS STAGE_CAP_S=$CAP GLOBAL_BUDGET_S=$GLOBAL CIVIL_DEADLINE_S=$CIVIL RSS_MIB=$RSS BIN_SHA=$BIN_SHA"
if [ -e "$ROOT/CONFIG.txt" ]; then [ "$(cat "$ROOT/CONFIG.txt")" = "$CFG" ] || { echo "CONFIG_MISMATCH: effective config differs from frozen $ROOT/CONFIG.txt"; exit 7; }
else echo "$CFG" > "$ROOT/CONFIG.txt"; fi
echo "start operative=$t0 civil=$c0 ($(date -u +%FT%TZ))" >> "$ROOT/CLOCKS.txt"
cprev=$c0
used_total() { local s=0 f v; for f in "$ROOT"/seed-*/budget_used_s; do [ -e "$f" ] || continue; v=$(cat "$f"); [[ "$v" =~ ^[0-9]+$ ]] || { echo 999999; return; }; s=$((s+v)); done; echo $s; }
for S in $SEEDS; do
  for ST in pre-g3 train-A train-F post-A post-F; do
    u=$(used_total); m=$(mono); [[ "$m" =~ ^[0-9]+$ ]] && [ "$m" -ge "$t0" ] || { echo "CLOCK_FAILED_OPERATIVE before seed $S stage $ST: stopping all"; exit 7; }; w=$((m - t0))
    cn=$(date +%s); [[ "$cn" =~ ^[0-9]+$ ]] || { echo "CLOCK_FAILED_CIVIL before seed $S stage $ST"; exit 7; }
    if [ "$cn" -lt "$cprev" ]; then echo "CIVIL_CLOCK_BACKWARD before seed $S stage $ST: $cn < $cprev: stopping fail closed"; exit 7; fi; cprev=$cn
    if [ $((cn-c0+CAP)) -gt "$CIVIL" ]; then echo "CIVIL_DEADLINE_EXHAUSTED before seed $S stage $ST: civil $((cn-c0))s of ${CIVIL}s: stopping; result will be INCOMPLETE"; exit 6; fi
    if [ $((u+CAP)) -gt "$GLOBAL" ] || [ $((w+CAP)) -gt "$GLOBAL" ]; then
      echo "GLOBAL_BUDGET_EXHAUSTED before seed $S stage $ST: reserved+used ${u}s wall ${w}s of ${GLOBAL}s (cap ${CAP}s): stopping; result will be INCOMPLETE"; exit 5
    fi
    BANK=g3 SEED=$S DIR="$ROOT/seed-$S" TOTAL_BUDGET_S=$((5*CAP)) STAGE_CAP_S=$CAP RSS_MIB=$RSS BIN="$BIN" BIN_SHA="$BIN_SHA" bash "$WRAP" "$ST"
    rc=$?
    echo "seed $S stage $ST rc=$rc"
    # global stop on anything that is not a stage outcome: preflight 3, already-done 4, budget 5, corrupt state 7, binary 8, lock 9, usage 2
    case $rc in 0) ;; 6) grep -q 'outcome=CENSORED_CLOCK' "$ROOT/seed-$S/status_$ST" 2>/dev/null && { echo "CLOCK failure in stage: stopping all"; exit 7; }; break;; *) echo "GLOBAL_STOP wrapper rc=$rc (clock/binary/config/lock/state): no further seeds"; exit 7;; esac
  done
done
echo "end operative=$(mono) civil=$(date +%s)" >> "$ROOT/CLOCKS.txt"
echo "driver finished: operative_used $(used_total)s, operative_wall $(( $(mono) - t0 ))s, civil $(( $(date +%s) - c0 ))s"
