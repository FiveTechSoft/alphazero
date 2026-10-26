# Pilot 4 analysis. Usage: python3 analyze4.py chunk_701_712.json chunk_713_724.json chunk_725_736.json chunk_737_748.json
# Run from the directory containing HASHES.txt, main.go, elig.go, go.mod, analyze4.py, PREREG.md, weights_elig.json.
# Validation (any failure: print reason, exit 1; preserve all files, stop; no repair, no silent rerun, no algorithm change):
#   exactly 3 arms {learned_elig, learned_elig_shuf, excl_random} x 48 seeds 701-748 = 144 rows, each (arm, seed) once, Spent == 6,000,000,
#   all required keys present, HASHES.txt has an entry for each required file and each hash matches.
# PRIMARY (only one): learned_elig - excl_random, paired by seed, 95% bootstrap CI (10000 resamples, random.Random(1)).
#   LB > 0: higher audit-solved for the learned selector than for random choice among eligible candidates, in this domain and for this package of choices.
#   CI contains 0: no advantage demonstrated (not equivalence, no cause). UB < 0: learned selector below excl_random.
# SECONDARY (descriptive, no multiplicity adjustment, no causal reading): learned_elig - learned_elig_shuf, learned_elig_shuf - excl_random, per-arm diagnostics.
import json,sys,random,collections,hashlib
def sha(p):return hashlib.sha256(open(p,'rb').read()).hexdigest()
def fail(msg):
    print('VALIDATION FAILED:',msg,'\nPreserve all files and stop for review. Do not repair, rerun silently or change the algorithm.');sys.exit(1)
if len(sys.argv)<2:fail('no chunk files given')
REQ=['main.go','elig.go','go.mod','analyze4.py','PREREG.md','weights_elig.json','train.json','run_pilot4.sh'];seen={}
for line in open('HASHES.txt'):
    h,f=line.split(None,1);seen[f.strip()]=h
for f in REQ:
    if f not in seen:fail('HASHES.txt has no entry for '+f)
    if sha(f)!=seen[f]:fail('hash mismatch for '+f)
rows=[]
for p in sys.argv[1:]:
    print('chunk',p,'sha256',sha(p));rows+=json.load(open(p))
KEYS={'Seed','Arm','AuditSolve','Rounds','Accepts','Evals','LibSize','Spent','WallMs','Ops','Fallback','Ties'}
ARMS={'learned_elig','learned_elig_shuf','excl_random'};SEEDS=list(range(701,749))
for x in rows:
    if not KEYS<=set(x):fail('row missing keys '+str(sorted(KEYS-set(x))))
if len(rows)!=144:fail('expected 144 rows, got %d'%len(rows))
if {x['Arm'] for x in rows}!=ARMS:fail('arms are not exactly '+str(sorted(ARMS)))
keys=[(x['Arm'],x['Seed']) for x in rows]
if len(set(keys))!=144:fail('duplicate (arm, seed) rows')
if {k[1] for k in keys}!=set(SEEDS):fail('seeds are not exactly 701-748')
if any(x['Spent']!=6000000 for x in rows):fail('some arm-seed did not spend exactly 6,000,000')
d=collections.defaultdict(dict)
for x in rows:d[x['Arm']][x['Seed']]=x
def m(a,k):return sum(d[a][s][k] for s in SEEDS)/48
print("arm n mean_audit rounds accepts evals lib fallback_rounds spent mean_wall_ms mean_ops   (arm order: alphabetical)")
for a in sorted(ARMS):print(a,48,*[round(m(a,k),2) for k in('AuditSolve','Rounds','Accepts','Evals','LibSize','Fallback','Spent','WallMs','Ops')])
rng=random.Random(1)
def paired(a,b):
    df=[d[a][s]['AuditSolve']-d[b][s]['AuditSolve'] for s in SEEDS];n=len(df)
    bs=sorted(sum(df[rng.randrange(n)] for _ in range(n))/n for _ in range(10000))
    return sum(df)/n,bs[250],bs[9749]
mu,lo,hi=paired('learned_elig','excl_random')
print("PRIMARY learned_elig - excl_random: %.2f [%.2f, %.2f]"%(mu,lo,hi))
print('reading:',('LB>0 criterion met' if lo>0 else 'UB<0: learned selector below excl_random' if hi<0 else 'CI contains 0: no advantage demonstrated (not equivalence, no cause)'))
for a,b in [('learned_elig','learned_elig_shuf'),('learned_elig_shuf','excl_random')]:print("secondary (descriptive) %s - %s: %.2f [%.2f, %.2f]"%((a,b)+paired(a,b)))
