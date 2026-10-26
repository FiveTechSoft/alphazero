import os as _os
EXP_ROOT=_os.environ.get('EXP_ROOT','.')

import json,numpy as np,sys
sys.path.insert(0,(EXP_ROOT+'/lang11')); import data11 as D, gt as G
pool=np.load('part/pool.npy'); Z=np.load('rev/val_prompts11.npz'); val=np.load('part/val.npy')
assert pool.shape==(1351,4,2) and G.NV==20 and max(G.STOP,G.DE,G.A,G.PAD,G.SUP,G.LAY,G.Q,G.OUT)<20 and D.C0+5<12<=G.STOP
rng=np.random.RandomState(0); idx=rng.randint(0,len(pool),500); pr=np.stack([rng.permutation(4) for _ in idx]); W=pool[idx][np.arange(500)[:,None],pr,0]; C=pool[idx][np.arange(500)[:,None],pr,1]
qp=np.array([D.PAIRS12[i] for i in rng.randint(12,size=500)]); qa=W[np.arange(500),qp[:,0]]; qb=W[np.arange(500),qp[:,1]]
x=D.prompt(W,C,qa,qb); g=np.stack([C[np.arange(500),qp[:,0]],C[np.arange(500),qp[:,1]]],1)
assert (D.oracle(x)==g).all() and x.shape==(500,25) and x.max()<20 and x.min()>=0
assert (x[:,9:18]==G.PAD).all() and (x[:,19]==G.PAD).all() and (x[:,22:24]==G.PAD).all()
Wv,Cv,qav,qbv,gv=[Z[k].astype(np.int64) for k in ('W','C','qa','qb','gold')]
xv=D.prompt(Wv,Cv,qav,qbv); assert (D.oracle(xv)==gv).all() and xv.shape==(1200,25)
assert all(len(set(r))==4 for r in Wv.tolist()) and all(len(set(r))==4 for r in Cv.tolist())
vs={tuple(map(tuple,c.tolist())) for c in val}; ps={tuple(map(tuple,c.tolist())) for c in pool}
ok=0
for w,c in zip(Wv,Cv):
    t=tuple(sorted(zip(w.tolist(),c.tolist()))); assert t in vs and t not in ps; ok+=1
AI=D.assoc_index(pool); assert AI.min()>=0 and AI.max()<36
p=D.weights(AI,np.zeros(36)); assert abs(p.sum()-1)<1e-9 and np.allclose(p,1/len(p)); e=np.random.RandomState(1).rand(36); p=D.weights(AI,e); assert abs(p.sum()-1)<1e-9 and p.max()/p.min()<=10+1e-9
S=json.load(open('rev/S_perms11.json')); assert all(len(S[k])==9 and all(sorted(r)==list(range(36)) for r in S[k]) for k in ('21','22'))
print('checks ok: val prompts in val contexts and not in pool:',ok,'; ESS ratio example',round(D.ess_ratio(p),3),'ratio',round(p.max()/p.min(),2))
