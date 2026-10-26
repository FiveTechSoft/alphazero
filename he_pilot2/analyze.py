import json,sys,random,collections
r=json.load(open(sys.argv[1]));d=collections.defaultdict(dict)
for x in r:d[x['Arm']][x['Seed']]=x
arms=sorted(d);seeds=sorted(d[arms[0]])
def m(a,k='AuditSolve'):return sum(d[a][s][k] for s in seeds)/len(seeds)
print("arm mean_audit mean_spent mean_wall_ms mean_ops")
for a in arms:print(a,round(m(a),2),round(m(a,'Spent')),round(m(a,'WallMs')),round(m(a,'Ops')))
rng=random.Random(1)
def paired(a,b):
    df=[d[a][s]['AuditSolve']-d[b][s]['AuditSolve'] for s in seeds];n=len(df)
    bs=sorted(sum(df[rng.randrange(n)] for _ in range(n))/n for _ in range(10000))
    return sum(df)/n,bs[250],bs[9749]
for a,b in [('transfer','random'),('transfer','freqrule'),('transfer','transfer_shuf'),('transfer','learned'),('learned','random'),('freqrule','random'),('transfer','ucb')]:
    if a in d and b in d:print("%s - %s: %.2f [%.2f, %.2f]"%((a,b)+paired(a,b)))
