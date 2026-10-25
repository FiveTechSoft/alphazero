#!/bin/bash
# guard.sh TIMEOUT_SEC RSS_LIMIT_MIB OUTFILE cmd args...
# Runs cmd as the leader of its own process group, polls the aggregate RSS of the group every
# 0.1 s, and kills the WHOLE group on: RSS limit, timeout, supervisor signal (TERM/INT/HUP), leader exit
# with descendants left behind, or a sampler that cannot read /proc. A second layer (timeout inside the
# group) ends the time budget even if this supervisor is SIGKILLed (memory is then not enforced).
# Limits: processes that create their own session/group (setsid, double fork daemons) escape; the RSS
# limit is a polling guard and can be overshot between polls.
lim=$1; rl=$2; out=$3; shift 3
P=${GUARD_PROC:-/proc}
pgid=0
members() { pgrep -g "$pgid" 2>/dev/null; }
killgroup() { [ "$pgid" -gt 0 ] && kill -KILL -- "-$pgid" 2>/dev/null; }
trap 'killgroup; sleep 0.3; echo "status=SUPERVISOR_SIGNAL residual_procs=$(members | wc -l)"; exit 143' TERM INT HUP
trap killgroup EXIT
setsid timeout -k 3 $((lim+5)) "$@" > "$out" 2>&1 &
pid=$!; pgid=$pid; start=$(date +%s); pk=0; why=exit; zero=0
while kill -0 $pid 2>/dev/null; do
  tot=0; seen=0
  for p in $(members); do r=$(awk '/VmRSS/{print $2}' "$P/$p/status" 2>/dev/null); if [ -n "$r" ]; then tot=$((tot+r)); seen=1; fi; done
  [ $tot -gt $pk ] && pk=$tot
  if [ $seen -eq 0 ]; then zero=$((zero+1)); else zero=0; fi
  if [ $zero -ge 5 ]; then why=SAMPLER_FAILED; break; fi
  if [ $((tot/1024)) -gt $rl ]; then why=RSS_LIMIT; break; fi
  if [ $(( $(date +%s) - start )) -ge $lim ]; then why=TIMEOUT; break; fi
  sleep 0.1
done
left_before=$(members | wc -l)
killgroup
wait $pid 2>/dev/null; rc=$?
sleep 0.3
left=$(members | wc -l)
[ $why = exit ] && [ $left_before -ge 1 ] && why=exit_with_descendants_killed
echo "status=$why rc=$rc peakRSS_MiB=$((pk/1024)) elapsed_s=$(( $(date +%s) - start )) descendants_alive_at_exit=$left_before residual_procs=$left"
[ $left -eq 0 ] || exit 99
