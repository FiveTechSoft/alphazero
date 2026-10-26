#!/bin/bash
# transfer_run.sh: read-only visto/nuevo evaluation driver (stages pre, A, F per seed), sequential, one CPU, no retries.
# Required env: BIN (d1a1transfer binary), BIN_SHA (its full sha256), ROOT (NEW output root), SRC_ROOT (finished A1C-B results, only read).
# Optional env (defaults are the plan): LEVELS=6,7  SEEDS="19 ... 26"  GLOBAL_BUDGET_S=7200 (operative)  STAGE_CAP_S=360
#   CIVIL_DEADLINE_S=14400 (separate civil budget, enforced between and INSIDE stages)  RSS_MIB=1024. Per-seed operative total 900 s.
# Exit codes: 0 loop finished (maybe INCOMPLETE: read status files); 5 operative budget exhausted; 6 civil deadline exhausted
#   (before a stage); 7 clock unreadable/backward, config mismatch, lock held, restart rejected, binary/state failure (stops ALL seeds).
# A stage that is not COMPLETED (CENSORED_TIMEOUT / CENSORED_CIVIL / CENSORED_CLOCK / CENSORED_MEMORY / FAILED) stops that seed only.
# Nothing is retried, nothing is re-parameterized, no budget is raised automatically. Operative clock: /proc/uptime seconds.
set -u
mono() { awk '{print int($1)}' /proc/uptime 2>/dev/null; }  # operative clock: /proc/uptime seconds (observed behaviour, not a universal guarantee)
HERE=$(cd "$(dirname "$0")" && pwd)
BIN=${BIN:?BIN}; BIN_SHA=${BIN_SHA:?BIN_SHA}; ROOT=${ROOT:?ROOT}; SRC_ROOT=${SRC_ROOT:?SRC_ROOT}; LEVELS=${LEVELS:-6,7}
SEEDS=${SEEDS:-"19 20 21 22 23 24 25 26"}
GLOBAL=${GLOBAL_BUDGET_S:-7200}; CAP=${STAGE_CAP_S:-360}; RSS=${RSS_MIB:-1024}
WRAP=${WRAP:-$HERE/transfer_stage.sh}
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
CFG="LEVELS=$LEVELS SRC_ROOT=$SRC_ROOT SEEDS=$SEEDS STAGE_CAP_S=$CAP GLOBAL_BUDGET_S=$GLOBAL CIVIL_DEADLINE_S=$CIVIL RSS_MIB=$RSS BIN_SHA=$BIN_SHA"
if [ -e "$ROOT/CONFIG.txt" ]; then [ "$(cat "$ROOT/CONFIG.txt")" = "$CFG" ] || { echo "CONFIG_MISMATCH: effective config differs from frozen $ROOT/CONFIG.txt"; exit 7; }
else echo "$CFG" > "$ROOT/CONFIG.txt"; fi
echo "start operative=$t0 civil=$c0 ($(date -u +%FT%TZ))" >> "$ROOT/CLOCKS.txt"
cprev=$c0
used_total() { local s=0 f v; for f in "$ROOT"/seed-*/budget_used_s; do [ -e "$f" ] || continue; v=$(cat "$f"); [[ "$v" =~ ^[0-9]+$ ]] || { echo 999999; return; }; s=$((s+v)); done; echo $s; }
for S in $SEEDS; do
  for ST in pre A F; do
    u=$(used_total); m=$(mono); [[ "$m" =~ ^[0-9]+$ ]] && [ "$m" -ge "$t0" ] || { echo "CLOCK_FAILED_OPERATIVE before seed $S stage $ST: stopping all"; exit 7; }; w=$((m - t0))
    cn=$(date +%s); [[ "$cn" =~ ^[0-9]+$ ]] || { echo "CLOCK_FAILED_CIVIL before seed $S stage $ST"; exit 7; }
    if [ "$cn" -lt "$cprev" ]; then echo "CIVIL_CLOCK_BACKWARD before seed $S stage $ST: $cn < $cprev: stopping fail closed"; exit 7; fi; cprev=$cn
    if [ $((cn-c0+CAP)) -gt "$CIVIL" ]; then echo "CIVIL_DEADLINE_EXHAUSTED before seed $S stage $ST: civil $((cn-c0))s of ${CIVIL}s: stopping; result will be INCOMPLETE"; exit 6; fi
    if [ $((u+CAP)) -gt "$GLOBAL" ] || [ $((w+CAP)) -gt "$GLOBAL" ]; then
      echo "GLOBAL_BUDGET_EXHAUSTED before seed $S stage $ST: reserved+used ${u}s wall ${w}s of ${GLOBAL}s (cap ${CAP}s): stopping; result will be INCOMPLETE"; exit 5
    fi
    SRC="$SRC_ROOT/seed-$S" LEVELS="$LEVELS" SEED=$S DIR="$ROOT/seed-$S" TOTAL_BUDGET_S=900 STAGE_CAP_S=$CAP RSS_MIB=$RSS BIN="$BIN" BIN_SHA="$BIN_SHA" bash "$WRAP" "$ST"
    rc=$?
    echo "seed $S stage $ST rc=$rc"
    # global stop on anything that is not a stage outcome: preflight 3, already-done 4, budget 5, corrupt state 7, binary 8, lock 9, usage 2
    case $rc in 0) ;; 6) grep -q 'outcome=CENSORED_CLOCK' "$ROOT/seed-$S/status_$ST" 2>/dev/null && { echo "CLOCK failure in stage: stopping all"; exit 7; }; break;; *) echo "GLOBAL_STOP wrapper rc=$rc (clock/binary/config/lock/state): no further seeds"; exit 7;; esac
  done
done
echo "end operative=$(mono) civil=$(date +%s)" >> "$ROOT/CLOCKS.txt"
echo "driver finished: operative_used $(used_total)s, operative_wall $(( $(mono) - t0 ))s, civil $(( $(date +%s) - c0 ))s"
