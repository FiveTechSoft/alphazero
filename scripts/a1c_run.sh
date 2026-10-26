#!/bin/bash
# a1c_run.sh: A1C execution plan, sequential, one CPU, no retries.
# Required env: BIN (binary), BIN_SHA (its full sha256), ROOT (output root).
# Optional env (defaults are the plan): SEEDS="19 20 21 22 23 24 25 26", GLOBAL_BUDGET_S=5400 (90 min,
# hard, counted as the sum of budget_used_s over seed dirs plus wall clock), STAGE_CAP_S=150, RSS_MIB=1024.
# Per seed 5 stages (pre-g3 train-A train-F post-A post-F), seed by seed. Each stage runs through
# pilot_a1_stage.sh (BANK=g3) under guard.sh, which polls the RSS of the whole process group and kills it
# above RSS_MIB (a polling guard: it can be overshot between polls; GOMEMLIMIT is NOT the limit).
# A censored or failed stage is kept as is: later stages of that seed refuse to run (status not COMPLETED),
# the driver goes on with the next seed, nothing is retried or re-parameterized. When the global budget
# cannot cover another stage the driver stops: the missing pairs make the primary result INCOMPLETE and
# not confirmatory. n is never reduced.
set -u
HERE=$(cd "$(dirname "$0")" && pwd)
BIN=${BIN:?BIN}; BIN_SHA=${BIN_SHA:?BIN_SHA}; ROOT=${ROOT:?ROOT}
SEEDS=${SEEDS:-"19 20 21 22 23 24 25 26"}
GLOBAL=${GLOBAL_BUDGET_S:-5400}; CAP=${STAGE_CAP_S:-150}; RSS=${RSS_MIB:-1024}
WRAP=${WRAP:-$HERE/pilot_a1_stage.sh}
export GUARD=${GUARD:-$HERE/guard.sh}
mkdir -p "$ROOT" || exit 7
t0=$(date +%s)
used_total() { local s=0 f v; for f in "$ROOT"/seed-*/budget_used_s; do [ -e "$f" ] || continue; v=$(cat "$f"); [[ "$v" =~ ^[0-9]+$ ]] || { echo 999999; return; }; s=$((s+v)); done; echo $s; }
for S in $SEEDS; do
  for ST in pre-g3 train-A train-F post-A post-F; do
    u=$(used_total); w=$(( $(date +%s) - t0 ))
    if [ $((u+CAP)) -gt "$GLOBAL" ] || [ $((w+CAP)) -gt "$GLOBAL" ]; then
      echo "GLOBAL_BUDGET_EXHAUSTED before seed $S stage $ST: reserved+used ${u}s wall ${w}s of ${GLOBAL}s (cap ${CAP}s): stopping; result will be INCOMPLETE"; exit 5
    fi
    BANK=g3 SEED=$S DIR="$ROOT/seed-$S" TOTAL_BUDGET_S=$((5*CAP)) STAGE_CAP_S=$CAP RSS_MIB=$RSS BIN="$BIN" BIN_SHA="$BIN_SHA" bash "$WRAP" "$ST"
    rc=$?
    echo "seed $S stage $ST rc=$rc"
    [ $rc -eq 0 ] || break
  done
done
echo "driver finished: used $(used_total)s, wall $(( $(date +%s) - t0 ))s"
