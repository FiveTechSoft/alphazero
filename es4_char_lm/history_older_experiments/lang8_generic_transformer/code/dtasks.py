


import numpy as np
import lang7 as L, gt as G
def arrays(eps):
    return dict(sw=np.stack([e['sw'] for e in eps]),sc=np.stack([e['sc'] for e in eps]),lay=np.stack([e['lay'] for e in eps]),s=np.array([e['s'] for e in eps]),d=np.array([e['d'] for e in eps]),
                b=np.array([e['b'] for e in eps]),l=np.array([e['l'] for e in eps]),gold=np.array([e['gold'] for e in eps]),grp=np.array([e['b']*1000+e['sid'] for e in eps]))
def prompts(task,A):
    N=len(A['sw']); x=G.ser(np.zeros((N,5),dtype=np.int64),A['sw'],A['sc'],A['lay']); ar=np.arange(N)
    x[:,19]=G.PAD; x[:,22]=G.PAD; x[:,23]=G.PAD
    cs=A['sc'][ar,(A['sw']==A['s'][:,None]).argmax(1)]; cd=A['sc'][ar,(A['sw']==A['d'][:,None]).argmax(1)]
    if task=='d1': x[:,20]=G.C0+cs; x[:,21]=G.C0+cd
    else: x[:,20]=A['s']; x[:,21]=A['d']
    return x
def oracle(x,task):

    N=len(x); ar=np.arange(N); sw=x[:,1:9:2]; sc=x[:,2:9:2]-G.C0; lc=x[:,[11,13,15,17]]-G.C0
    if task=='d1': ca=x[:,20]-G.C0; cb=x[:,21]-G.C0
    else: ca=sc[ar,(sw==x[:,20:21]).argmax(1)]; cb=sc[ar,(sw==x[:,21:22]).argmax(1)]
    return np.stack([(lc==ca[:,None]).argmax(1),(lc==cb[:,None]).argmax(1)],1)
def cf_layout(x,A,rng):

    N=len(x); l2=np.zeros(N,dtype=np.int64)
    for i in range(N):
        c=[l for l in range(24) if L.answer(int(A['b'][i]),(int(A['s'][i]),int(A['d'][i])),l)!=tuple(A['gold'][i])]; l2[i]=c[rng.randint(len(c))]
    lay2=np.array([L.PERMS[l] for l in l2]); x2=x.copy()
    for i in range(4): x2[:,11+2*i]=G.C0+lay2[:,i]
    return x2,l2
def cf_support(x,A,rng,change):

    N=len(x); b2=np.zeros(N,dtype=np.int64)
    for i in range(N):
        c=[b for b in range(24) if (L.answer(b,(int(A['s'][i]),int(A['d'][i])),int(A['l'][i]))!=tuple(A['gold'][i]))==change]; b2[i]=c[rng.randint(len(c))]
    sc2=np.array([[L.PERMS[b][w] for w in A['sw'][i]] for i,b in enumerate(b2)]); x2=x.copy(); x2[:,2:9:2]=G.C0+sc2
    return x2,b2
