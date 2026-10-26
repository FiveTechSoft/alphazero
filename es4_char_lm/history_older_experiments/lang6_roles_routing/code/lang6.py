
import numpy as np, itertools, time
from collections import Counter
PAIRS=[(x,y) for x in range(4) for y in range(4) if x!=y]
RES=[(0,1),(1,2),(2,3),(3,0)]; SEEN=[p for p in PAIRS if p not in RES]
PERMS=list(itertools.permutations(range(4)))
perm0=np.random.RandomState(0).permutation(24); NEWB=[int(i) for i in perm0[:6]]; TRB=[int(i) for i in perm0[6:]]
perm5=np.random.RandomState(5).permutation(24); TRL=[int(i) for i in perm5[:18]]; HL=[int(i) for i in perm5[18:]]
PAD=4
def answer(b,p,l):
    L=PERMS[l]; return (L.index(PERMS[b][p[0]]),L.index(PERMS[b][p[1]]))
def gen(rng,bset,pairs,layouts,perm_rng=None):
    b=int(bset[rng.randint(len(bset))]); p=pairs[rng.randint(len(pairs))]; l=int(layouts[rng.randint(len(layouts))])
    order=rng.permutation(4); sw=order.astype(np.int64)
    sc=np.array([PERMS[b][w] for w in sw],dtype=np.int64)
    sc2=None
    if perm_rng is not None:
        b2=int(perm_rng.choice([x for x in range(24) if x!=b])); sc2=np.array([PERMS[b2][w] for w in sw],dtype=np.int64)
    return dict(b=b,p=p,l=l,sw=sw,sc=sc,sc2=sc2,lay=np.array(PERMS[l],dtype=np.int64),gold=answer(b,p,l))
def validate(ep,bset,pairs,layouts):
    v=[]
    if ep['b'] not in bset: v.append('binding_split')
    if ep['p'] not in pairs: v.append('pair_split')
    if ep['l'] not in layouts: v.append('layout_split')
    if sorted(ep['sw'])!=[0,1,2,3] or sorted(ep['sc'])!=[0,1,2,3]: v.append('support_not_four_primitives')

    col={int(w):int(c) for w,c in zip(ep['sw'],ep['sc'])}
    L=list(ep['lay']); g=(L.index(col[ep['p'][0]]),L.index(col[ep['p'][1]]))
    if g!=ep['gold']: v.append('oracle_wrong')
    return v
def pack(eps,arm):
    sw=np.stack([e['sw'] for e in eps]); lay=np.stack([e['lay'] for e in eps])
    sc=np.stack([e['sc2'] if (arm=='perm') else e['sc'] for e in eps])
    if arm=='empty': sw=np.full_like(sw,PAD); sc=np.full_like(sc,PAD)
    X=np.array([e['p'][0] for e in eps]); Y=np.array([e['p'][1] for e in eps]); gold=np.array([e['gold'] for e in eps])
    return sw,sc,lay,X,Y,gold
def init(seed,d=16):
    r=np.random.RandomState(seed); f=lambda *s,sc=1.0:(r.randn(*s)*sc).astype(np.float32)
    return dict(Ew=f(5,d),Ec=f(5,d),Wq=f(d,d,sc=1/np.sqrt(d)),Wk=f(d,d,sc=1/np.sqrt(d)),Wv=f(d,d,sc=1/np.sqrt(d)),Wp=f(d,d,sc=1/np.sqrt(d)))
def fwd(P,sw,sc,lay,X,Y,gold):
    B=len(X); d=P['Ew'].shape[1]; scl=1.0/np.sqrt(d)
    w=np.concatenate([X,Y]); swN=np.concatenate([sw,sw]); scN=np.concatenate([sc,sc]); layN=np.concatenate([lay,lay]); tg=np.concatenate([gold[:,0],gold[:,1]])
    qe=P['Ew'][w]; q=qe@P['Wq']; ke=P['Ew'][swN]; K=ke@P['Wk']; ve=P['Ec'][scN]; V=ve@P['Wv']
    s=np.einsum('nd,nid->ni',q,K)*scl; s=s-s.max(1,keepdims=True); a=np.exp(s); a/=a.sum(1,keepdims=True)
    r=np.einsum('ni,nid->nd',a,V); ce=P['Ec'][layN]; Pc=ce@P['Wp']
    lg=np.einsum('nd,nkd->nk',r,Pc)*scl; lg=lg-lg.max(1,keepdims=True); p=np.exp(lg); p/=p.sum(1,keepdims=True)
    n=np.arange(2*B); loss=-np.log(p[n,tg]+1e-30).sum()/B
    return loss,(w,swN,scN,layN,tg,qe,q,ke,K,ve,V,a,r,ce,Pc,p,scl,B)
def bwd(P,c):
    w,swN,scN,layN,tg,qe,q,ke,K,ve,V,a,r,ce,Pc,p,scl,B=c
    g={k:np.zeros_like(v) for k,v in P.items()}; N=2*B; n=np.arange(N)
    dl=p.copy(); dl[n,tg]-=1; dl/=B
    dr=scl*np.einsum('nk,nkd->nd',dl,Pc); dPc=scl*np.einsum('nk,nd->nkd',dl,r)
    g['Wp']=np.einsum('nkd,nke->de',ce,dPc); np.add.at(g['Ec'],layN,dPc@P['Wp'].T)
    da=np.einsum('nd,nid->ni',dr,V); dV=a[:,:,None]*dr[:,None,:]
    ds=a*(da-(a*da).sum(1,keepdims=True))
    dq=scl*np.einsum('ni,nid->nd',ds,K); dK=scl*ds[:,:,None]*q[:,None,:]
    g['Wv']=np.einsum('nid,nie->de',ve,dV); np.add.at(g['Ec'],scN,dV@P['Wv'].T)
    g['Wk']=np.einsum('nid,nie->de',ke,dK); np.add.at(g['Ew'],swN,dK@P['Wk'].T)
    g['Wq']=qe.T@dq; np.add.at(g['Ew'],w,dq@P['Wq'].T)
    return g
def predict(P,sw,sc,lay,X,Y):
    z=np.zeros((len(X),2),dtype=np.int64); _,c=fwd(P,sw,sc,lay,X,Y,z); B=len(X); p=c[15]
    return np.stack([p[:B].argmax(1),p[B:].argmax(1)],1)
def gradcheck():
    P=init(5,4); P={k:v.astype(np.float64) for k,v in P.items()}; rng=np.random.RandomState(6)
    eps=[gen(rng,TRB,SEEN,TRL) for _ in range(3)]; sw,sc,lay,X,Y,gold=pack(eps,'full')
    loss,c=fwd(P,sw,sc,lay,X,Y,gold); g=bwd(P,c); worst=0.0; n=0; rr=np.random.RandomState(9)
    for k in P:
        flat=P[k].reshape(-1)
        for i in rr.choice(flat.size,min(8,flat.size),replace=False):
            o=flat[i]; e=1e-5; flat[i]=o+e; lp=fwd(P,sw,sc,lay,X,Y,gold)[0]; flat[i]=o-e; lm=fwd(P,sw,sc,lay,X,Y,gold)[0]; flat[i]=o
            num=(lp-lm)/(2*e); an=g[k].reshape(-1)[i]; worst=max(worst,abs(num-an)/(abs(num)+abs(an)+1e-8)); n+=1
    return worst,n
def bound(bs,ps,ls):
    s=0;m=0
    for p in ps:
        for l in ls:
            s+=Counter(answer(b,p,l) for b in bs).most_common(1)[0][1]/len(bs); m+=1
    return s/m
SPLITS={'TR':(TRB,SEEN,TRL,5,'TR'),'E1':(NEWB,SEEN,HL,6,'E1'),'E2':(TRB,RES,HL,3,'E2'),'E3':(NEWB,RES,HL,36,'E3')}

def build_eval(seed):
    rng=np.random.RandomState(seed); prng=np.random.RandomState(seed+1); out={}
    for name,(bs,ps,ls,K,_) in SPLITS.items():
        eps=[]
        for b in bs:
            for p in ps:
                for _ in range(K):
                    e=gen(rng,[b],[p],ls,prng)
                    v=validate(e,bs,ps,ls)
                    if v: raise RuntimeError('eval validation '+name+str(v))
                    e['grp']=b*100+PAIRS.index(p); eps.append(e)
        out[name]=eps
    return out
