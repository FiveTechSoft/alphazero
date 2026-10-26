#!/bin/bash
# transfer_stage.sh STAGE (pre|A|F): ONE read-only evaluation stage of d1a1transfer per call, under guard.sh.
# Extra required env: SRC (finished seed results dir, only read), SEED. Output: $DIR/transfer-$STAGE.json (written atomically by the binary).
# Required env: BIN (binary), BIN_SHA (expected full sha256 of BIN), DIR (state dir). Fails closed on any doubt.
# Exit: 0 completed; 2 usage; 3 preflight; 4 already done/censored/failed; 5 budget exhausted; 6 stage not completed;
#       7 corrupt/inconsistent state; 8 binary mismatch; 9 lock held.
set -u
mono() { awk '{print int($1)}' /proc/uptime 2>/dev/null; }  # operative clock: /proc/uptime seconds (observed behaviour, not a universal guarantee)
STAGE=${1:?stage}
HERE=$(cd "$(dirname "$0")" && pwd)
BIN=${BIN:?BIN}; BIN_SHA=${BIN_SHA:?BIN_SHA}; DIR=${DIR:?DIR}
GUARD=${GUARD:-$HERE/guard.sh}
RSS_MIB=${RSS_MIB:-900}; TOTAL=${TOTAL_BUDGET_S:-900}; CAP=${STAGE_CAP_S:-360}; SEED=${SEED:-101}
ARGS="-seed $SEED -perround 40 -evalcap 100 -levels ${LEVELS:-6,7} -net $STAGE -dir ${SRC:?SRC}"
case "$STAGE" in pre|A|F) ;; *) echo "unknown stage"; exit 2;; esac
mkdir -p "$DIR" || exit 7
# exclusive atomic lock (mkdir). Stale lock after SIGKILL is NOT auto-cleared: fail closed.
mkdir "$DIR/.lock" 2>/dev/null || { echo "LOCK_HELD: $DIR/.lock (another run, or a killed run: investigate)"; exit 9; }
GPID=""
cleanup(){ rmdir "$DIR/.lock" 2>/dev/null; }
on_sig(){ SIGNALLED=1; [ -n "$GPID" ] && kill -TERM "$GPID" 2>/dev/null; }
SIGNALLED=0
trap on_sig INT TERM HUP
trap cleanup EXIT
atomic_write(){ printf '%s\n' "$2" > "$1.tmp.$$" && sync "$1.tmp.$$" 2>/dev/null; mv -f "$1.tmp.$$" "$1"; }
isint(){ [[ "$1" =~ ^[0-9]+$ ]]; }
# binary fingerprint: must equal BIN_SHA and any earlier recorded one
real=$(sha256sum "$BIN" 2>/dev/null | cut -d' ' -f1)
[ -n "$real" ] && [ "$real" = "$BIN_SHA" ] || { echo "BINARY_MISMATCH: got '$real' expected '$BIN_SHA'"; exit 8; }
if [ -e "$DIR/binary.sha256" ]; then [ "$(cat "$DIR/binary.sha256")" = "$real" ] || { echo "BINARY_MISMATCH vs recorded"; exit 8; }
else atomic_write "$DIR/binary.sha256" "$real"; fi
# stage-state checks
[ -e "$DIR/inflight" ] && { echo "STATE_INCONSISTENT: previous stage ended abnormally (inflight marker): $(cat "$DIR/inflight")"; exit 7; }
for f in "$DIR"/status_*; do [ -e "$f" ] || continue; case "$(cat "$f")" in *outcome=COMPLETED*) ;; *) echo "PRIOR_STAGE_NOT_COMPLETED: $f"; exit 4;; esac; done
[ -e "$DIR/status_$STAGE" ] && { echo "stage $STAGE already has status: $(cat "$DIR/status_$STAGE")"; exit 4; }
# budget: must exist as integer, else (if absent and no stage has run) start at 0; corrupt => fail closed
if [ -e "$DIR/budget_used_s" ]; then used=$(cat "$DIR/budget_used_s"); isint "$used" || { echo "BUDGET_CORRUPT: '$used'"; exit 7; }
else ls "$DIR"/status_* >/dev/null 2>&1 && { echo "BUDGET_MISSING with prior stages"; exit 7; }; used=0; fi
remaining=$((TOTAL-used)); [ "$remaining" -gt 5 ] || { echo "BUDGET_EXHAUSTED: used ${used}s of ${TOTAL}s"; exit 5; }
lim=$(( remaining < CAP ? remaining : CAP ))
# preflight
free_mib=$(awk '/MemAvailable/{print int($2/1024)}' /proc/meminfo)
isint "$free_mib" && [ "$free_mib" -ge 1300 ] || { echo "PREFLIGHT_FAIL: ${free_mib} MiB available"; exit 3; }
pgrep -f ' (d1a1train|d1a1transfer|d1g1train|d1train) ' >/dev/null && { echo "PREFLIGHT_FAIL: another d1train-family process running"; exit 3; }
# RESERVE the full limit before running (conservative: a signal/kill never gives time back)
atomic_write "$DIR/inflight" "stage=$STAGE reserved=${lim}s prev_used=${used}s"
atomic_write "$DIR/budget_used_s" "$((used+lim))"
start=$(mono); [ -n "$start" ] || { echo "CLOCK_FAILED"; exit 7; }
OUT="$DIR/transfer-$STAGE.log"
bash "$GUARD" "$lim" "$RSS_MIB" "$OUT" "$BIN" d1a1transfer $ARGS -out "$DIR/transfer-$STAGE.json" > "$DIR/$STAGE.guard" 2>&1 &
GPID=$!
wait "$GPID"; grc=$?
# a trapped signal interrupts wait; make sure guard really ended
if [ "$SIGNALLED" = 1 ]; then wait "$GPID" 2>/dev/null; grc=$?; fi
em=$(mono); if [[ "$em" =~ ^[0-9]+$ ]] && [ "$em" -ge "$start" ]; then el=$((em-start)); else el=$lim; ELNOTE=" elapsed_unknown_reserved_counted"; fi
gline=$(tail -n1 "$DIR/$STAGE.guard" 2>/dev/null)
if [ "$SIGNALLED" = 1 ]; then outcome="CENSORED_SIGNAL"
elif [ "$grc" -eq 0 ] && echo "$gline" | grep -q '^status=exit rc=0 ' && [ -e "$DIR/transfer-$STAGE.json" ] && [ ! -e "$DIR/transfer-$STAGE.json.tmp" ] && echo "$gline" | grep -q 'residual_procs=0$'; then outcome="COMPLETED"
elif echo "$gline" | grep -q 'status=TIMEOUT'; then outcome="CENSORED_TIMEOUT"
elif echo "$gline" | grep -q 'status=CIVIL_DEADLINE'; then outcome="CENSORED_CIVIL"
elif echo "$gline" | grep -q 'status=CIVIL_CLOCK\|status=CLOCK'; then outcome="CENSORED_CLOCK"
elif echo "$gline" | grep -q 'status=.*RSS\|status=OOM\|status=MEM'; then outcome="CENSORED_MEMORY"
else outcome="FAILED"; fi
atomic_write "$DIR/status_$STAGE" "outcome=$outcome guard_rc=$grc elapsed_s=$el limit_s=$lim guard=[$gline]${ELNOTE:-}"
# conservative accounting: keep reservation unless clean completion shortens it; elapsed is never under-counted
if [ "$outcome" = COMPLETED ]; then atomic_write "$DIR/budget_used_s" "$((used+el))"
else atomic_write "$DIR/budget_used_s" "$((used + (el > lim ? el : lim)))"; fi  # failure: max(observed, reserved)
rm -f "$DIR/inflight"
echo "STAGE $STAGE outcome=$outcome rc=$grc elapsed=${el}s limit=${lim}s budget_used=$(cat "$DIR/budget_used_s")s/${TOTAL}s"
[ "$outcome" = COMPLETED ] || { echo "NOT COMPLETED: state kept, no retry, no parameter change"; exit 6; }
