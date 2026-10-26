#!/usr/bin/env bash
# Pilot 4 TEST driver. Run from the package directory only after the public freeze. Not run during preparation.
# Any failure or interruption: stop, preserve every log/JSON, report before any relaunch. No silent repeat, no algorithm change, no mixing of partials.
set -euo pipefail
[ "$(go version)" = "go version go1.26.6 linux/amd64" ] || { echo "Go environment differs from the declared one: stop and report"; exit 1; }
[ "$(python3 --version)" = "Python 3.10.12" ] || { echo "Python differs from the declared one: stop and report"; exit 1; }
sha256sum -c HASHES.txt
for c in chunk_701_712 chunk_713_724 chunk_725_736 chunk_737_748; do
  if [ -e "$c.json" ] || [ -e "$c.log" ]; then echo "output already exists: $c.json or $c.log; stop and report"; exit 1; fi
done
for f in analysis4.txt he he.sha256; do
  if [ -e "$f" ]; then echo "output already exists: $f; stop and report"; exit 1; fi
done
go vet ./...
go test ./...
go build -o he .
sha256sum he | tee he.sha256
check() { python3 - "$1" "$2" "$3" <<'PY'
import json,sys,collections
f,lo,hi=sys.argv[1],int(sys.argv[2]),int(sys.argv[3])
r=json.load(open(f));c=collections.Counter((x['Arm'],x['Seed']) for x in r)
exp={(a,s):1 for a in('learned_elig','learned_elig_shuf','excl_random') for s in range(lo,hi+1)}
assert dict(c)==exp,'chunk '+f+' is not exactly 3 arms x seeds %d-%d once each'%(lo,hi)
assert all(x['Spent']==6000000 for x in r),'Spent != 6000000 in '+f
print('chunk ok',f,len(r),'rows')
PY
}
run_chunk() {
  lo=$1; hi=$2
  j=chunk_${lo}_${hi}.json; l=chunk_${lo}_${hi}.log
  if [ -e "$j" ] || [ -e "$l" ]; then echo "output already exists: $j or $l; stop and report"; exit 1; fi
  ./he -lo "$lo" -hi "$hi" -arms learned_elig,learned_elig_shuf,excl_random -weights weights_elig.json -out "$j" 2>&1 | tee "$l"
  check "$j" "$lo" "$hi"
}
run_chunk 701 712
run_chunk 713 724
run_chunk 725 736
run_chunk 737 748
python3 analyze4.py chunk_701_712.json chunk_713_724.json chunk_725_736.json chunk_737_748.json | tee analysis4.txt
