import os as _os
EXP_ROOT=_os.environ.get('EXP_ROOT','.')

import sys,itertools,json,collections
sys.path.insert(0,(EXP_ROOT+'/lang8')); import lang7 as L
PM=L.PERMS; TR=set(L.TRB); NEW=set(L.NEWB)
def kind(b): return 'S' if b in TR else 'H'
pairs=collections.defaultdict(list)
for q in itertools.combinations(range(4),2):
    for b1 in range(24):
        for b2 in range(b1+1,24):
            if all(PM[b1][w]==PM[b2][w] for w in q) and any(PM[b1][w]!=PM[b2][w] for w in range(4)): pairs[q].append((b1,b2))
tot=collections.Counter(); per_q={}
for q,ps in pairs.items():
    c=collections.Counter(''.join(sorted(kind(a)+kind(b))) for a,b in ps); per_q[str(q)]=dict(c); tot.update(c)
print('unordered binding pairs per query word-set',{k:sum(v.values()) for k,v in per_q.items()}); print('by type per query set',per_q); print('total',dict(tot),sum(tot.values()))


print('support orders 24; ordered query strings per word-set: 2 (one in TRS, one in RSS)')

def dist(a,b): return sum(x!=y for x,y in zip(PM[a],PM[b]))
def nn_ans(b,q):
    ds={t:dist(b,t) for t in L.TRB}; m=min(ds.values()); c=[t for t in ds if ds[t]==m]
    ans=[(PM[t][q[0]],PM[t][q[1]]) for t in c]; return collections.Counter(ans).most_common(1)[0][0]
mode={}
for q in itertools.permutations(range(4),2):
    cnt=collections.Counter((PM[b][q[0]],PM[b][q[1]]) for b in L.TRB); mode[q]=cnt.most_common(1)[0][0]
res=collections.defaultdict(lambda: collections.Counter())
for q0,ps in pairs.items():
  for q in (q0,q0[::-1]):
    for b1,b2 in ps:
        k=''.join(sorted(kind(b1)+kind(b2))); g=(PM[b1][q[0]],PM[b1][q[1]])
        for name,f in (('modal',lambda b:mode[q]),('nearest_binding',lambda b:(PM[b][q[0]],PM[b][q[1]]) if b in TR else nn_ans(b,q))):
            a1,a2=f(b1),f(b2); r=res[(k,name)]; r['n']+=1; r['same_pred']+=a1==a2; r['both_correct']+=(a1==g and a2==g); r['one_correct']+=((a1==g)!=(a2==g)); r['both_wrong']+=(a1!=g and a2!=g)
out={}
for (k,name),r in sorted(res.items()):
    n=r['n']; out[k+'|'+name]={m:round(r[m]/n,3) for m in ('same_pred','both_correct','one_correct','both_wrong')}|{'n_ordered_query_pairs':n}
for k,v in out.items(): print(k,v)
json.dump(dict(per_q=per_q,total=dict(tot),baselines=out),open((EXP_ROOT+'/lang10/comb10.json'),'w'),indent=1)
