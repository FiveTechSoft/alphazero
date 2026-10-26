
import itertools,json,hashlib,os,numpy as np
SEED_PART=20261025; NW=6; NC=6
rng=np.random.RandomState(SEED_PART)

ctxs=[tuple(sorted(zip(ws,cs))) for ws in itertools.combinations(range(NW),4) for cs in itertools.permutations(range(NC),4)]
ctxs=[tuple(zip(ws,cs)) for ws in itertools.combinations(range(NW),4) for cs in itertools.permutations(range(NC),4)]
assert len(ctxs)==15*360 and len(set(ctxs))==len(ctxs)
def popairs(c): return [frozenset(p) for p in itertools.combinations(c,2)]
allpp=sorted({tuple(sorted(p)) for c in ctxs for p in popairs(c)})
assert len(allpp)==15*30
perm=rng.permutation(len(allpp)); Xv={allpp[i] for i in perm[:45]}; Xt={allpp[i] for i in perm[45:90]}
assert not (Xv&Xt)
def has(c,X): return any(tuple(sorted(p)) in X for p in popairs(c))
pool=[c for c in ctxs if not has(c,Xv) and not has(c,Xt)]
val=[c for c in ctxs if has(c,Xv) and not has(c,Xt)]
test=[c for c in ctxs if has(c,Xt) and not has(c,Xv)]
P,V,T=map(lambda L:{c for c in L},(pool,val,test)); assert not (P&V) and not (P&T) and not (V&T)
def assoc_counts(L):
    cnt=np.zeros((NW,NC),dtype=int)
    for c in L:
        for w,k in c: cnt[w,k]+=1
    return cnt
cp,cv,ct=assoc_counts(pool),assoc_counts(val),assoc_counts(test)
ppool={tuple(sorted(p)) for c in pool for p in popairs(c)}
checks=dict(n_contexts=len(ctxs),n_pairs_of_assoc=len(allpp),n_withheld_Xv=len(Xv),n_withheld_Xt=len(Xt),pool=len(pool),val=len(val),test=len(test),
  pool_min_context_count_per_association=int(cp.min()),val_min_context_count_per_association=int(cv.min()),test_min_context_count_per_association=int(ct.min()),
  pool_covers_all_individual_associations=bool((cp>0).all()),pool_pairs_of_assoc_covered=len(ppool),pairs_of_assoc_not_in_pool=len(allpp)-len(ppool),
  pool_contains_no_withheld=not any(has(c,Xv|Xt) for c in pool),val_all_contain_Xv_none_Xt=all(has(c,Xv) and not has(c,Xt) for c in val),test_all_contain_Xt_none_Xv=all(has(c,Xt) and not has(c,Xv) for c in test))
os.makedirs('part',exist_ok=True); os.makedirs('sealed',exist_ok=True)
arr=lambda L:np.array(L,dtype=np.int8)
np.save('part/pool.npy',arr(pool)); np.save('part/val.npy',arr(val)); np.save('sealed/test.npy',arr(test))
sha=lambda f:hashlib.sha256(open(f,'rb').read()).hexdigest()
man=dict(seed_part=SEED_PART,rule='pool: no withheld pair-of-associations; val: >=1 in Xv and none in Xt; test: >=1 in Xt and none in Xv; Xv,Xt = 45+45 of 450 pairs-of-associations chosen by seed',checks=checks,
  hashes=dict(pool=sha('part/pool.npy'),val=sha('part/val.npy'),test_sealed=sha('sealed/test.npy'),Xv=hashlib.sha256(repr(sorted(Xv)).encode()).hexdigest(),Xt=hashlib.sha256(repr(sorted(Xt)).encode()).hexdigest(),constructor=sha('mk11.py')))
json.dump(man,open('manifest11.json','w'),indent=1); print(json.dumps(man,indent=1))
