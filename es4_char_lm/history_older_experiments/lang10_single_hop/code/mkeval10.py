import os as _os
EXP_ROOT=_os.environ.get('EXP_ROOT','.')

import numpy as np,json,hashlib
import data10 as D, lang7 as L
SPL={'TR':(L.TRB,L.TRS),'E_b':(L.NEWB,L.TRS),'E_s':(L.TRB,L.RSS),'E_bs':(L.NEWB,L.RSS)}
rng=np.random.RandomState(777); out={}; base={}

mode={}
for (s,d,o) in L.TRS+L.RSS:
    cnt={}
    for b in L.TRB: p=(L.PERMS[b][s],L.PERMS[b][d]); cnt[p]=cnt.get(p,0)+1
    mode[(s,d)]=max(cnt,key=cnt.get)
for k,(bs,st) in SPL.items():
    A=D.eps(rng,bs,st,720); sc=D.true_sc(A); g=D.gold(A); x=D.prompt(A,sc)
    assert (D.oracle(x,A)==g).all() and x.shape==(720,25) and int(x.max())<20
    cfsc=D.cf_sc(A,np.random.RandomState(778),bs); xc=D.prompt(A,cfsc); gc=D.oracle(xc,A)
    assert (gc!=g).any(1).all()
    out[k+'_x']=x; out[k+'_gold']=g; out[k+'_grp']=A['b']; out[k+'_xnc']=D.prompt(A,sc,'none'); out[k+'_xcf']=xc; out[k+'_goldcf']=gc
    pr=np.array([mode[(s,d)] for s,d in zip(A['s'],A['d'])]); base[k]=dict(prior_exact=float((pr==g).all(1).mean()),prior_first=float((pr[:,0]==g[:,0]).mean()))

import math
H=0.0
for (s,d,o) in L.TRS:
    c1={};c12={}
    for b in L.TRB:
        a=L.PERMS[b][s]; bb=L.PERMS[b][d]; c1[a]=c1.get(a,0)+1; c12[(a,bb)]=c12.get((a,bb),0)+1
    h=0.0
    for (a,bb),n in c12.items(): h-=n/18*(math.log(c1[a]/18)+math.log(n/c1[a]))
    H+=h/len(L.TRS)
base['plugin_floor_loss_mean3']=H/3; base['note']='context-free plug-in estimate (TRB bindings x TRS strings), not a universal floor; STOP term taken as 0'
np.savez((EXP_ROOT+'/lang10/eval10.npz'),**out); json.dump(base,open((EXP_ROOT+'/lang10/base10.json'),'w'),indent=1)
print(json.dumps(base)); print(hashlib.sha256(open((EXP_ROOT+'/lang10/eval10.npz'),'rb').read()).hexdigest()[:12])
