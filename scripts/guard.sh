#!/bin/bash
# v2: operative time read from /proc/uptime (integer seconds); civil time from date. Observed on the machine used: uptime did
# not advance across the execution pauses seen while date jumped. This is NOT assumed to hold universally; a monotonic clock
# going backwards is NOT checked (elapsed that would be negative is reported UNKNOWN); a civil clock going backwards is detected.
# guard.sh TIMEOUT_SEC RSS_LIMIT_MIB OUTFILE cmd args...
# Runs cmd as the leader of its own process group, polls the aggregate RSS of the group every
# 0.1 s, and kills the WHOLE group on: RSS limit, timeout, supervisor signal (TERM/INT/HUP), leader exit
# with descendants left behind, or a sampler that cannot read /proc. A second layer (timeout inside the
# group) ends the time budget even if this supervisor is SIGKILLed (memory is then not enforced).
# Limits: processes that create their own session/group (setsid, double fork daemons) escape; the RSS
# limit is a polling guard and can be overshot between polls.
lim=$1; rl=$2; out=$3; shift 3
mono() { awk '{print int($1)}' /proc/uptime 2>/dev/null; }  # operative clock (see header: observed behaviour, not a universal guarantee)
P=${GUARD_PROC:-/proc}
pgid=0
members() { pgrep -g "$pgid" 2>/dev/null; }
killgroup() { [ "$pgid" -gt 0 ] && kill -KILL -- "-$pgid" 2>/dev/null; }
trap 'killgroup; sleep 0.3; echo "status=SUPERVISOR_SIGNAL residual_procs=$(members | wc -l)"; exit 143' TERM INT HUP
setsid "$@" > "$out" 2>&1 &
pid=$!; pgid=$pid; start=$(mono); [ -n "$start" ] || { kill -KILL -- "-$pgid" 2>/dev/null; echo "status=CLOCK_FAILED"; exit 98; }; pk=0; why=exit; zero=0
# Second layer (replaces coreutils timeout): a watcher on the same clocks ends the group at lim+5 (operative),
# at the absolute civil deadline, or on any unreadable clock; it keeps running while ANY group member lives, so
# a dead supervisor + dead leader + live children is still cleaned up. Empty/invalid clock => kill (fail closed).
( while :; do
    m=$(mono); [[ "$m" =~ ^[0-9]+$ ]] || { kill -KILL -- "-$pgid" 2>/dev/null; exit 0; }
    [ $((m - start)) -ge $((lim+5)) ] && { kill -KILL -- "-$pgid" 2>/dev/null; exit 0; }
    if [ -n "${GUARD_CIVIL_END:-}" ]; then c=$(date +%s 2>/dev/null); { [[ "$c" =~ ^[0-9]+$ ]] && [ "$c" -lt "$GUARD_CIVIL_END" ]; } || { kill -KILL -- "-$pgid" 2>/dev/null; exit 0; }; fi
    if ! kill -0 "$pid" 2>/dev/null && ! pgrep -g "$pgid" >/dev/null 2>&1; then exit 0; fi  # exit only when leader AND group are gone (setsid may not have created the group yet)
    sleep 1
  done ) >/dev/null 2>&1 &
wpid=$!
cs=$(date +%s 2>/dev/null); cprev=${cs:-0}; [[ "$cs" =~ ^[0-9]+$ ]] || cs=UNKNOWN
trap 'kill $wpid 2>/dev/null; killgroup' EXIT
while kill -0 $pid 2>/dev/null; do
  tot=0; seen=0
  for p in $(members); do r=$(awk '/VmRSS/{print $2}' "$P/$p/status" 2>/dev/null); if [ -n "$r" ]; then tot=$((tot+r)); seen=1; fi; done
  [ $tot -gt $pk ] && pk=$tot
  if [ $seen -eq 0 ]; then zero=$((zero+1)); else zero=0; fi
  if [ $zero -ge 5 ]; then why=SAMPLER_FAILED; break; fi
  if [ $((tot/1024)) -gt $rl ]; then why=RSS_LIMIT; break; fi
  now=$(mono); if [ -z "$now" ]; then why=CLOCK_FAILED; break; fi
  if [ $((now - start)) -ge $lim ]; then why=TIMEOUT; break; fi
  if [ -n "${GUARD_CIVIL_END:-}" ]; then cn=$(date +%s 2>/dev/null); if ! [[ "$cn" =~ ^[0-9]+$ ]]; then why=CIVIL_CLOCK_FAILED; break; fi; if [ "$cn" -lt "$cprev" ]; then why=CIVIL_CLOCK_BACKWARD; break; fi; cprev=$cn; if [ "$cn" -ge "$GUARD_CIVIL_END" ]; then why=CIVIL_DEADLINE; break; fi; fi
  sleep 0.1
done
left_before=$(members | wc -l)
killgroup
wait $pid 2>/dev/null; rc=$?
sleep 0.3
left=$(members | wc -l)
[ $why = exit ] && [ $left_before -ge 1 ] && why=exit_with_descendants_killed
em=$(mono); if [[ "$em" =~ ^[0-9]+$ ]] && [ "$em" -ge "$start" ]; then el=$((em-start)); else el=UNKNOWN; fi
ce=$(date +%s 2>/dev/null); [[ "$ce" =~ ^[0-9]+$ ]] || ce=UNKNOWN
echo "status=$why rc=$rc peakRSS_MiB=$((pk/1024)) elapsed_s=$el civil_start=$cs civil_end=$ce descendants_alive_at_exit=$left_before residual_procs=$left"
[ $left -eq 0 ] || exit 99
