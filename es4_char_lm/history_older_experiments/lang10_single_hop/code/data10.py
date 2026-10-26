import os as _os
EXP_ROOT=_os.environ.get('EXP_ROOT','.')

import sys,numpy as np
sys.path.insert(0,(EXP_ROOT+'/lang8'))
import lang7 as L, gt as G
C0=G.C0; PM=np.array(L.PERMS)
def eps(rng,bset,strings,n):
    E=[L.gen(rng,bset,strings,L.TRL,'train',None) for _ in range(n)]
    return dict(sw=np.stack([e['sw'] for e in E]),b=np.array([e['b'] for e in E]),s=np.array([e['s'] for e in E]),d=np.array([e['d'] for e in E]))
def colors(A,sc):
    ar=np.arange(len(sc)); return sc[ar,(A['sw']==A['s'][:,None]).argmax(1)],sc[ar,(A['sw']==A['d'][:,None]).argmax(1)]
def true_sc(A): return PM[A['b']][np.arange(len(A['b']))[:,None],A['sw']]
def prompt(A,sc,ctxmode='ctx'):
    N=len(A['b']); x=np.full((N,G.PL),G.PAD,dtype=np.int64); x[:,0]=G.SUP
    if ctxmode=='ctx': x[:,1:9:2]=A['sw']; x[:,2:9:2]=C0+sc
    x[:,18]=G.Q; x[:,20]=A['s']; x[:,21]=A['d']; x[:,24]=G.OUT
    return x
def gold(A):
    sc=true_sc(A); a,b=colors(A,sc); return np.stack([a,b],1)
def target(g): return np.concatenate([C0+g,np.full((len(g),1),G.STOP)],1).astype(np.int64)
def oracle(x,A):
    sw=x[:,1:9:2]; sc=x[:,2:9:2]-C0; ar=np.arange(len(x)); a=sc[ar,(sw==x[:,20:21]).argmax(1)]; b=sc[ar,(sw==x[:,21:22]).argmax(1)]
    return np.stack([a,b],1)
def shuffled_sc(A,rng_b2,bset):
    b2=np.array([bset[rng_b2.randint(len(bset))] for _ in range(len(A['b']))]); return PM[b2][np.arange(len(b2))[:,None],A['sw']]
def cf_sc(A,rng,bset):

    g0=gold(A); sc2=np.zeros((len(A['b']),4),dtype=np.int64); b2=np.zeros(len(A['b']),dtype=np.int64)
    for i in range(len(A['b'])):
        c=[b for b in bset if tuple(colors({k:v[i:i+1] for k,v in A.items()},PM[b][A['sw'][i:i+1]])[0:2][0])!=(g0[i,0],) or tuple(colors({k:v[i:i+1] for k,v in A.items()},PM[b][A['sw'][i:i+1]])[1])!=(g0[i,1],)]
        b2[i]=c[rng.randint(len(c))]
    return PM[b2][np.arange(len(b2))[:,None],A['sw']]
