#!/usr/bin/env bash
# Pilot 3 TEST driver: run only after the public freeze of this package. Run from the package directory.
# Any failure or interruption: stop. Preserve every log/JSON. Report before any relaunch; no silent repeat, no algorithm change, no mixing partials.
set -euo pipefail
[ "$(go version)" = "go version go1.26.6 linux/amd64" ] || { echo "Go environment differs from the declared one: stop and report"; exit 1; }
[ "$(python3 --version)" = "Python 3.10.12" ] || { echo "Python differs from the declared one: stop and report"; exit 1; }
sha256sum -c HASHES.txt
for c in chunk_601_612 chunk_613_624 chunk_625_636 chunk_637_648; do
  if [ -e "$c.json" ] || [ -e "$c.log" ]; then echo "output already exists: $c.json or $c.log; stop and report"; exit 1; fi
done
if [ -e analysis3.txt ]; then echo "output already exists: analysis3.txt; stop and report"; exit 1; fi
go build -o he .
check() { python3 - "$1" "$2" "$3" <<'PY'
import json,sys,collections
f,lo,hi=sys.argv[1],int(sys.argv[2]),int(sys.argv[3])
r=json.load(open(f));c=collections.Counter((x['Arm'],x['Seed']) for x in r)
exp={(a,s):1 for a in('freqrule','excl_random','random') for s in range(lo,hi+1)}
assert dict(c)==exp,'chunk '+f+' is not exactly 3 arms x seeds %d-%d once each'%(lo,hi)
assert all(x['Spent']==6000000 for x in r),'Spent != 6000000 in '+f
print('chunk ok',f,len(r),'rows')
PY
}
run_chunk() {
  lo=$1; hi=$2
  j=chunk_${lo}_${hi}.json; l=chunk_${lo}_${hi}.log
  if [ -e "$j" ] || [ -e "$l" ]; then echo "output already exists: $j or $l; stop and report"; exit 1; fi
  ./he -lo "$lo" -hi "$hi" -arms freqrule,excl_random,random -out "$j" 2>&1 | tee "$l"
  check "$j" "$lo" "$hi"
}
run_chunk 601 612
run_chunk 613 624
run_chunk 625 636
run_chunk 637 648
python3 analyze3.py chunk_601_612.json chunk_613_624.json chunk_625_636.json chunk_637_648.json | tee analysis3.txt
