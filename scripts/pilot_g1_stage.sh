#!/bin/bash
# pilot_g1_stage.sh STAGE  (G1 variant of the inc11 wrapper). ONE stage of the frozen G1 protocol per call, under guard.sh.
# Required env: BIN (binary), BIN_SHA (expected full sha256 of BIN), DIR (state dir). Fails closed on any doubt.
# Exit: 0 completed; 2 usage; 3 preflight; 4 already done/censored/failed; 5 budget exhausted; 6 stage not completed;
#       7 corrupt/inconsistent state; 8 binary mismatch; 9 lock held.
set -u
STAGE=${1:?stage}
HERE=$(cd "$(dirname "$0")" && pwd)
BIN=${BIN:?BIN}; BIN_SHA=${BIN_SHA:?BIN_SHA}; DIR=${DIR:?DIR}
GUARD=${GUARD:-$HERE/guard.sh}
RSS_MIB=${RSS_MIB:-900}; TOTAL=${TOTAL_BUDGET_S:-900}; CAP=${STAGE_CAP_S:-110}; SEED=${SEED:-101}
ARGS="-seed $SEED -rounds 6 -perround 40 -cap 100 -evalcap 100"
case "$STAGE" in pre-g1|train|post-g1|report) ;; *) echo "unknown stage"; exit 2;; esac
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
[ -e "$DIR/status_$STAGE" ] && [ "$STAGE" != report ] && { echo "stage $STAGE already has status: $(cat "$DIR/status_$STAGE")"; exit 4; }
# budget: must exist as integer, else (if absent and no stage has run) start at 0; corrupt => fail closed
if [ -e "$DIR/budget_used_s" ]; then used=$(cat "$DIR/budget_used_s"); isint "$used" || { echo "BUDGET_CORRUPT: '$used'"; exit 7; }
else ls "$DIR"/status_* >/dev/null 2>&1 && { echo "BUDGET_MISSING with prior stages"; exit 7; }; used=0; fi
remaining=$((TOTAL-used)); [ "$remaining" -gt 5 ] || { echo "BUDGET_EXHAUSTED: used ${used}s of ${TOTAL}s"; exit 5; }
lim=$(( remaining < CAP ? remaining : CAP ))
# preflight
free_mib=$(awk '/MemAvailable/{print int($2/1024)}' /proc/meminfo)
isint "$free_mib" && [ "$free_mib" -ge 1300 ] || { echo "PREFLIGHT_FAIL: ${free_mib} MiB available"; exit 3; }
pgrep -f ' (d1g1train|d1train) ' >/dev/null && { echo "PREFLIGHT_FAIL: another d1train/d1g1train running"; exit 3; }
# RESERVE the full limit before running (conservative: a signal/kill never gives time back)
atomic_write "$DIR/inflight" "stage=$STAGE reserved=${lim}s prev_used=${used}s"
atomic_write "$DIR/budget_used_s" "$((used+lim))"
start=$(date +%s)
OUT="$DIR/$STAGE.log"; [ "$STAGE" = report ] && OUT="$DIR/report.out"
"$GUARD" "$lim" "$RSS_MIB" "$OUT" "$BIN" d1g1train $ARGS -stage "$STAGE" -dir "$DIR" > "$DIR/$STAGE.guard" 2>&1 &
GPID=$!
wait "$GPID"; grc=$?
# a trapped signal interrupts wait; make sure guard really ended
if [ "$SIGNALLED" = 1 ]; then wait "$GPID" 2>/dev/null; grc=$?; fi
el=$(( $(date +%s) - start ))
gline=$(tail -n1 "$DIR/$STAGE.guard" 2>/dev/null)
if [ "$SIGNALLED" = 1 ]; then outcome="CENSORED_SIGNAL"
elif [ "$grc" -eq 0 ] && { [ "$STAGE" = report ] || [ -e "$DIR/$STAGE.json" ]; } && ! grep -q 'residual_procs=[1-9]' "$DIR/$STAGE.guard"; then outcome="COMPLETED"
elif echo "$gline" | grep -q 'status=TIMEOUT'; then outcome="CENSORED_TIMEOUT"
elif echo "$gline" | grep -q 'status=.*RSS\|status=OOM\|status=MEM'; then outcome="CENSORED_MEMORY"
else outcome="FAILED"; fi
atomic_write "$DIR/status_$STAGE" "outcome=$outcome guard_rc=$grc elapsed_s=$el limit_s=$lim guard=[$gline]"
# conservative accounting: keep reservation unless clean completion shortens it; elapsed is never under-counted
if [ "$outcome" = COMPLETED ]; then atomic_write "$DIR/budget_used_s" "$((used+el))"; fi
rm -f "$DIR/inflight"
echo "STAGE $STAGE outcome=$outcome rc=$grc elapsed=${el}s limit=${lim}s budget_used=$(cat "$DIR/budget_used_s")s/${TOTAL}s"
[ "$outcome" = COMPLETED ] || { echo "NOT COMPLETED: state kept, no retry, no parameter change"; exit 6; }
