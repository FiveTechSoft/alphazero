import json,sys,random
r=json.load(open(sys.argv[1]));arms=sorted({x['Arm'] for x in r});seeds=sorted({x['Seed'] for x in r})
d={(x['Seed'],x['Arm']):x['AuditSolve'] for x in r}
print("mean audit solved /300:")
for a in arms: print(f"  {a:17s} {sum(d[s,a] for s in seeds)/len(seeds):.1f}")
random.seed(0)
for b in arms:
  if b=='learned':continue
  diffs=[d[s,'learned']-d[s,b] for s in seeds]
  bs=sorted(sum(random.choice(diffs) for _ in diffs)/len(diffs) for _ in range(5000))
  w=sum(x>0 for x in diffs);l=sum(x<0 for x in diffs)
  print(f"learned - {b:17s} mean {sum(diffs)/len(diffs):+.2f} CI95 [{bs[125]:+.2f},{bs[4875]:+.2f}] wins {w} losses {l} n={len(diffs)}")
